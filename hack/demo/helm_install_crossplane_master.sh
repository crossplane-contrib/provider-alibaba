#!/bin/bash

set -x
set errexit

echo "Install Crossplane..."
kubectl create namespace crossplane-system

# Command to add helm repo first:
#   helm repo add crossplane-master https://charts.crossplane.io/master/
version=$(helm search repo crossplane --devel | awk '$1 == "crossplane-master/crossplane" {print $2}')
helm install crossplane --namespace crossplane-system crossplane-master/crossplane --version $version --devel


echo "Install OAM Runtime..."
kubectl create namespace oam-system
helm install oam --namespace oam-system crossplane-master/oam-kubernetes-runtime --devel

echo "Applying the application’s Components..."
files=(
  "tracker-data-component.yaml"
  "tracker-flights-component.yaml"
  "tracker-quakes-component.yaml"
  "tracker-weather-component.yaml"
  "tracker-ui-component.yaml"
  "tracker-db-component.yaml"
 )

for myfile in ${files[@]}; do
  kubectl apply -f hack/demo/deploy/components/${myfile}
done

