#!/bin/bash

while IFS='=' read -r key value; do
  # skip empty lines and comments
  [[ -z "$key" || "$key" == \#* ]] && continue
  
  echo "Creating secret: $key"
  echo -n "$value" | gcloud secrets create "$key" --data-file=- --replication-policy="automatic"
done < .env.prod