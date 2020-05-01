#!/bin/bash

set -x
set errexit


echo "Install Crossplane..."
kubectl create namespace crossplane-system

# Command to add helm repo first:
#   helm repo add crossplane-master https://charts.crossplane.io/master/
version=$(helm search repo crossplane --devel | awk '$1 == "crossplane-master/crossplane" {print $2}')
helm install crossplane --namespace crossplane-system crossplane-master/crossplane --version $version --devel

echo "Install Cert Manager..."
kubectl create namespace cert-manager

until kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.14.0/cert-manager.yaml; do
  sleep 5
done

echo "Install OAM Core types operator..."

kubectl create namespace oam-system
# If repo doesn't exist, do:
#   git clone git@github.com:crossplane/addon-oam-kubernetes-local.git
helm install oam-local -n oam-system ../addon-oam-kubernetes-local/charts/oam-core-resources/

echo "Waiting to setup oam-sytem ..."
sleep 60

kubectl create namespace oam-system

helm install controller -n oam-system ~/code/crossplane/addon-oam-kubernetes-local/charts/oam-core-resources/


echo "Applying oam definitions..."
kubectl apply -f https://raw.githubusercontent.com/oam-dev/samples/1.0.0-alpha2/2.ServiceTracker_App/Definitions/containerized-workload.yaml
kubectl apply -f https://raw.githubusercontent.com/oam-dev/samples/1.0.0-alpha2/2.ServiceTracker_App/Definitions/managed-postgres-workload.yaml
kubectl apply -f https://raw.githubusercontent.com/oam-dev/samples/1.0.0-alpha2/2.ServiceTracker_App/Definitions/manual-scaler-trait.yaml

echo "Applying the application’s Components (except db component)..."
files=(
  "tracker-data-component.yaml"
  "tracker-flights-component.yaml"
  "tracker-quakes-component.yaml"
  "tracker-weather-component.yaml"
  "tracker-ui-component.yaml"
 )

for myfile in ${files[@]}; do
  kubectl apply -f https://raw.githubusercontent.com/oam-dev/samples/1.0.0-alpha2/2.ServiceTracker_App/Components/${myfile}
done

echo "Applying the Application Configuration..."
kubectl apply -f https://raw.githubusercontent.com/oam-dev/samples/1.0.0-alpha2/2.ServiceTracker_App/ApplicationConfiguration/tracker-app-config-managed.yaml

