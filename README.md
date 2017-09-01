# Mars Images App

Displays the latest images from the Mars rover Curiosity.

## Create CloudSQL Instance

```
gcloud sql instances create mars-images --region us-central1
```

Generate and set a password for the `mars` user:

```
export CLOUDSQL_USER=mars
export CLOUDSQL_PASSWORD=$(openssl rand -base64 15)

gcloud sql users create ${CLOUDSQL_USER} '%' --instance mars-images --password=${CLOUDSQL_PASSWORD}
```

```
GOOGLE_PROJECT=$(gcloud config get-value project)
CLOUDSQL_CONNECTION_NAME="${GOOGLE_PROJECT}:us-central1:mars-images"
```

Create the `mars-images` database:

```
gcloud sql databases create mars-images --instance=mars-images
```

## Deploy to App Engine

```
gcloud app deploy
```

## Load data

Use the admin route to load the initial data set.

```
open $(gcloud app browse --no-launch-browser)/update
```

> Enter your admin password to authenticate when prompted.

## Local development

Start the Close SQL Proxy:

```
GOOGLE_PROJECT=$(gcloud config get-value project)
CLOUDSQL_CONNECTION_NAME="${GOOGLE_PROJECT}:us-central1:mars-images"
```

```
./cloud_sql_proxy -instances=${CLOUDSQL_CONNECTION_NAME}=tcp:3306
```

Edit the app.yaml and uncomment these lines:

```
SQL_CONNECTION_PROTO: tcp
CLOUDSQL_CONNECTION_NAME: 127.0.0.1:3306
```

Start the dev server:

```
dev_appserver.py --port 9999 .
```

Open the dev webpage:

```
open http://localhost:9999
```
