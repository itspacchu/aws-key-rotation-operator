package cmd

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func DeploymentAdded(inf interface{}) {
	deployment := inf.(*appsv1.Deployment)
	log.Printf("Deployment added: %s/%s", deployment.Namespace, deployment.Name)
}

func PodUpdated(clientset *kubernetes.Clientset, old interface{}, new interface{}) {
	_ = old.(*corev1.Pod)
	newPod := new.(*corev1.Pod)

	for _, container := range newPod.Status.ContainerStatuses {
		if container.State.Waiting != nil && container.State.Waiting.Reason == "ImagePullBackOff" {
			log.Warnf("Pod %s is in ImagePullBackOff\n", newPod.Name)
			accountID, region, _, err := parseDockerURI(container.Image)
			if err != nil {
				log.Warnf("skipping! with %s", err)
				continue
			}
			token, err := GetTokenForRegion(region)

			if err != nil {
				log.Warnf("unable to fetch for region %s", region)
				continue
			}
			ApplySecretObject(token, newPod.Namespace, accountID, region, clientset)
		}

	}
}

func DeploymentUpdated(clientset *kubernetes.Clientset, old interface{}, new interface{}) {
	oldDeploy := old.(*appsv1.Deployment)
	newDeploy := new.(*appsv1.Deployment)
	if oldDeploy.Status.ObservedGeneration == newDeploy.Status.ObservedGeneration {
		return
	}

	log.Infof("Deployment Updated: %s/%s [%d] -> %s", oldDeploy.Namespace, oldDeploy.Name, oldDeploy.Status.ObservedGeneration, newDeploy.Status.Conditions[len(newDeploy.Status.Conditions)-1].Message)
	podSpecs := newDeploy.Spec.Template.Spec
	for _, container := range podSpecs.Containers {
		accountID, region, _, err := parseDockerURI(container.Image)
		if err != nil {
			log.Warnf("skipping! with %s", err)
			continue
		}
		token, err := GetTokenForRegion(region)

		if err != nil {
			log.Warnf("unable to fetch for region %s", region)
			continue
		}

		ApplySecretObject(token, newDeploy.Namespace, accountID, region, clientset)
	}
}

func DeploymentDeleted(inf interface{}) {
	deployment := inf.(*appsv1.Deployment)
	log.Warnf("Deployment deleted: %s/%s", deployment.Namespace, deployment.Name)
}

func HandlePodRestarts(clientset *kubernetes.Clientset, ctx context.Context) {
	factory := informers.NewSharedInformerFactory(clientset, time.Hour*24)
	podInformer := factory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			PodUpdated(clientset, oldObj, newObj)
		},
	})

	factory.Start(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced) {
		log.Fatalf("failed to sync cache")
	}
}

func HandleDeploymentChanges(clientset *kubernetes.Clientset, ctx context.Context) {
	factory := informers.NewSharedInformerFactory(clientset, time.Hour*24)
	deploymentInformer := factory.Apps().V1().Deployments().Informer()

	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: DeploymentAdded,
		UpdateFunc: func(oldObj, newObj interface{}) {
			DeploymentUpdated(clientset, oldObj, newObj)
		},
		DeleteFunc: DeploymentDeleted,
	})

	factory.Start(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), deploymentInformer.HasSynced) {
		log.Fatalf("failed to sync cache")
	}
}
