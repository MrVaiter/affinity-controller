#!/bin/bash

helm uninstall metacontroller

kubectl delete -f ./yaml/decoratorcontroller.yaml

kubectl delete -f ./yaml/deploy.yaml
