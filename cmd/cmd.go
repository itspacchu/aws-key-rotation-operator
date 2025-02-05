package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"encoding/base64"
	"encoding/json"

	"github.com/charmbracelet/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	HashTokenLookup map[string]string
	SecretName      string
	Namespace       string
)

func Run() error {
	if _, present := os.LookupEnv("KUBERNETES_SERVICE_HOST"); !present {
		log.Warn("Not running in Kubernetes pod!")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Warn(err.Error() + "...Attempting to fetch kubeconfig file!")
		home := homedir.HomeDir()
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
		if err != nil {
			return fmt.Errorf("unable to connect to cluster : %s", err.Error())
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create client")
	}
	log.Info("Connected with Kubernetes API!")

	// Some code stealing was done https://pkg.go.dev/k8s.io/client-go/informers#SharedInformerFactory
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go HandleDeploymentChanges(clientset, ctx)
	go HandlePodRestarts(clientset, ctx)

	wait.Until(func() {}, time.Second, ctx.Done())
	return nil
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

func ApplySecretObject(token string, namespace string, accountID string, region string, clientset *kubernetes.Clientset) error {
	secret := GenerateSecretObject(
		SecretName,
		token,
		fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", accountID, region),
		namespace,
	)

	err := clientset.CoreV1().Secrets(secret.Namespace).Delete(context.TODO(), secret.Name, metav1.DeleteOptions{})
	if err == nil {
		log.Warnf("Deleted %s:%s", secret.Namespace, secret.Name)
	}

	_, err = clientset.CoreV1().Secrets(secret.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Error %s:%s -> %v", secret.Namespace, secret.Name, err)
		return err
	}
	log.Infof("Created %s:%s for registry %s", secret.Namespace, secret.Name, fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", accountID, region))
	return err
}

func GenerateSecretObject(name string, token string, registry string, namespace string) *corev1.Secret {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("AWS:%s", token)))
	dockerConfig := map[string]interface{}{
		"auths": map[string]interface{}{
			registry: map[string]string{
				"auth":     auth,
				"password": token + "\n", // \n present in other secrets
				"username": "AWS",
			},
		},
	}
	dockerConfigJson, err := json.Marshal(dockerConfig)
	if err != nil {
		log.Fatalf("Error marshaling DockerConfigJson: %v", err)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerConfigJson,
		},
	}
	return secret
}
