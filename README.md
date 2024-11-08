# K8S Namespace Protection Webhook

## Overview
Namespace Protection is a Kubernetes webhook that aims to prevent any accidental deletion of a namespace which could in the loss of Pods,Secrets,PVCs...

It provides a better alternative than finalizers for namespaced resources that require to stay untouched in the event of a delete event on the namespace
## How it works
The Helm Chart deploys a Validating Webhook that listens to delete operations on the Kind Namespace object. 

The webhook cancels the deletion operation on the namespace when the **protect-deletion** annotation is set to True. This results in the following error:
```
Error from server: admission webhook "namespace-protection.kube-system.svc.cluster.local" denied the request: This namespace is protected against delete operations
```

## Install
```
helm install -n kube-system namespace-protection ./chart
```

## How to use
Add the following annotation to the namespace:
```yaml
metadata:
  annotations:
    protect-deletion: "true"
```

