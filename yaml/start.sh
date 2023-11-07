#!/bin/bash

helm install metacontroller ./metacontroller-helm-v2.2.5.tgz

kubectl apply -f ./decoratorcontroller.yaml

kubectl apply -f ./deploy.yaml
