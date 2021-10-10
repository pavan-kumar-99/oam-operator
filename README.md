
# Oam Operator ( The Cloud Journey Accelerator )

![Unit Tests master workflow](https://github.com/pavan-kumar-99/oam-operator/actions/workflows/unit-tests.yaml/badge.svg)
![Linting master workflow](https://github.com/pavan-kumar-99/oam-operator/actions/workflows/lint.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/pavan-kumar-99/oam-operator)](https://goreportcard.com/report/github.com/pavan-kumar-99/oam-operator)

A Kubernetes Operator, that would help all the DevOps teams to accelerate their Journey into the cloud and K8s. OAM operator scaffolds all of the code required to create resources across various cloud provides, which includes both K8s and Non-K8s resources. For example an user can create all the required resources for the application ( K8s resources like Deployments, Statefulsets, Ingresses, Non-k8s resources like S3, RDS, EKS clusters ) with 10 lines of YAML. See [Example Usage](#example-usage)


## Overview
* [Architecture](#oam-operator-architecture)
* [Deploy the CRD's](#deploy-the-crds)
  - [Using make](#using-make)
  - [Using kubectl](#using-kubectl)
* [Deploy the Operator](#deploy-the-operator)
  - [Using make](#using-make)
* [Example Usage](#example-usage)
  - [Create custom resource](#create-custom-resource)
  - [Describe status](#describe-status)
* [Cleanup](#cleanup)
* [Future Enhancements](#future-enhancements)

### OAM Operator Architecture
![oam-operator](https://user-images.githubusercontent.com/54094196/134817952-8af98e13-768b-4d20-a34a-2aa744498844.png)

### Supported Versions
*  Kubernetes 1.18-1.22
*  OpenShift 3.11, 4.4-4.8
*  kubebuilder 3.1.0
*  controller-gen 0.4.1
*  kustomize 3.8.7

### Deploy the CRDs
### Using make
```bash
$ make install
```

### Using kubectl

Deploy the CustomResourceDefinitions (CRDs) for the operator.

```bash
kubectl apply -k config/crds/
```

### Deploy the Operator

```bash
$ make deploy
```

### Example Usage
#### Create custom resource
```
$ kubectl create -f config/samples/apps_v1beta1_application.yaml
```
#### Describe status
```
$ kubectl describe Application web-app
Name:         web-app
API Version:  apps.oam.cfcn.io/v1beta1
Kind:         Application
Metadata:
  Finalizers:
    finalizer.app
Spec:
  Application Name:  web-app
  Cloud:
    Aws:
      s3:  true
Events:
  Type    Reason   Age   From         Message
  ----    ------   ----  ----         -------
  Normal  Created  24s   Application  Created S3 Bucket
  Normal  Created  24s   Application  Created HPA
  Normal  Created  24s   Application  Created Service
  Normal  Created  23s   Application  Created Deployment
  Normal  Created  23s   Application  Created Ingress
```

#### Cleanup
```
$ make uninstall 
$ make undeploy
```
#### Future Enhancements
- [ ] Add Support to Multiple Cloud Providers
- [ ] Support Consul Injectors 
- [ ] Add Support to External Providers
- [ ] Add Support to GitOps Deployments
