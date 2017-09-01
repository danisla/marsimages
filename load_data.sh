#!/bin/bash

CLOUDSQL_USER=${CLOUDSQL_USER:-mars}
[[ -z "${CLOUDSQL_PASSWORD}" ]] && read -p "SQL Password: " CLOUDSQL_PASSWORD

GOOGLE_PROJECT=$(gcloud config get-value project)
CLOUDSQL_CONNECTION_NAME="${GOOGLE_PROJECT}:us-central1:mars-images"

if [[ ! -e ./cloud_sql_proxy ]]; then
  echo "INFO: Downloading cloud_sql_proxy"

  URL="https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64"
  if [[ "$(uname)" =~ Darwin ]]; then
    URL="https://dl.google.com/cloudsql/cloud_sql_proxy.darwin.amd64"
  fi
  curl -L "${URL}" -o cloud_sql_proxy
  chmod +x cloud_sql_proxy
fi

echo "INFO: Starting cloud_sql_proxy"

./cloud_sql_proxy -instances=${CLOUDSQL_CONNECTION_NAME}=tcp:3306 >/dev/null 2>&1 &
export fork=%1

# Cleanup fork on exit
function finish {
  echo "INFO: Stopping cloud_sql_proxy"
  kill $fork
}
trap finish EXIT

echo "INFO: Loading data..."

cd ./load-data-sql/
go run main.go -user ${CLOUDSQL_USER} -password ${CLOUDSQL_PASSWORD}
cd - >/dev/null

echo "INFO: Done"