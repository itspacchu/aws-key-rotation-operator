# incluster-aws-key-rotation

A simple utility to do aws key rotation for Amazon's ECR Private repositories. Uses client-go informers to watch for deployment or pod status changes and applies specified secret in the corresponding namespace.


### TODO

- [x] ECR URL detection (improvement needed)
- [x] Generate Kubernetes secret
- [x] Use Informers instead of Cronjobs
- [ ] Annotation for Namespace filtering
- [ ] Use kustomization


### Usage

Create secret with AWS credentials:

```sh
kubectl create secret generic aws-secrets \
  --from-literal=aws-access-key-id=ABCD \
  --from-literal=aws-secret-access-key=xyz/abc
```

Apply manifest files in `./manifests`


```sh
kubectl apply -f ./manifests/
```

WIP
