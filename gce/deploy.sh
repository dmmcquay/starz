#!/bin/bash

echo "creating coreos instance on GCE"
gcloud compute instances create core1 \
    --image https://www.googleapis.com/compute/v1/projects/coreos-cloud/global/images/coreos-alpha-1000-0-0-v20160328 \
    --zone us-central1-a \
    --machine-type n1-standard-1 \
    --metadata-from-file user-data=cloud-config.yaml

echo "grabbing ip to create static IP for coreos instance"
staticIP=$(gcloud compute instances list | awk -F ' ' '{print $5}' | awk 'FNR == 2 {print}')

echo "creating static ip for coreos instance"
gcloud compute addresses create --addresses $staticIP --region us-central1

echo $staticIP
