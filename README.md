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

## Load data

Use the helper script to load data into the SQL database through the [cloud_sql_proxy](https://cloud.google.com/sql/docs/mysql/sql-proxy):

```
./load_data.sh
```

This script will prompt you for the SQL password if it hasn't been exported and can be run multiple times to update the data.

## Deploy to App Engine

```
gcloud app deploy
```

Open in browser:

```
gcloud app browse
```

## Local development

Start the Close SQL Proxy:

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
