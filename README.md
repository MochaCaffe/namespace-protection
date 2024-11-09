# K8S Namespace Protection Webhook

## Overview
Namespace Protection is a Kubernetes webhook that aims to prevent accidental deletion of a namespace which could result in the loss of Pods,Secrets,PVCs...

It provides a better alternative than finalizers: namespaced resources will stay untouched in the event of a delete event on the namespace
## How it works
The Helm Chart deploys a Validating Webhook that listens to delete operations on the Kind Namespace object. 

The webhook rejects the deletion operation on the namespace when the **protect-deletion** annotation is set to True. This results in the following error:
```
Error from server: admission webhook "namespace-protection.kube-system.svc.cluster.local" denied the request: This namespace is protected against deletion
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

