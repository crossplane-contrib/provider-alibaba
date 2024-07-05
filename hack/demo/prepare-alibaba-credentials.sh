#!/bin/bash

kubectl create secret generic alibaba-account-creds -n crossplane-system \
    --from-literal=accessKeyId=${ALICLOUD_ACCESS_KEY} \
    --from-literal=accessKeySecret=${ALICLOUD_SECRET_KEY}
