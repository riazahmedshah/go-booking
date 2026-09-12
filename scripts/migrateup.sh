#!/bin/bash
set -e # Exit immediately if tern fails

if [ -z "$DB_DSN" ]; then
  echo "Error: DB_DSN environment variable is empty."
  exit 1
fi

tern migrate -m ./internal/database/migrations --conn-string "$DB_DSN"