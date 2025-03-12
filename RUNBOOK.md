# How to run the SRS application

## Running in K8s with Minikube and Helm

This process describe how to run the application inside k8s with minikube. When the application is deployed it runs automatically the database migrations. The process to load the stock rating should be triggered manually.

Prerequisites:

- Install docker
- Install minikube
- Install helm
- Install kubectl
- Install cockroachdb or create an account in cockroach db cloud and get the database credentials

Steps:

1. Start minikube
2. Change the docker context to minukube running `eval $(minikube -p minikube docker-env)`
3. Build the docker image `dk build . -t yourrepository/srs`
4. Package the helm charts `helm package helm`
5. Create a credentials-values.yaml file with the env vars and secrets **Do not commit this file**
6. Deploy the helm charts `helm upgrade --install -n srs srs srs-0.1.0.tgz --create-namespace -f ./credentials-values.yaml`
7. Expose the srs service with minikube `mk service -n srs srs-nginx --url`
8. Go to the provide url

### Example credentials-values.yaml file
```yaml
env:
  - name: SRS_DATABASE
    value:
  - name: SRS_DATABASE_PORT
    value:
  - name: SRS_DATABASE_USER
    value:
  - name: SRS_DATABASE_PASSWORD
    value:
  - name: SRS_DATABASE_HOST
    value:
  - name: STOCK_RATING_API_URL
    value:
  - name: STOCK_RATING_API_AUTH_TOKEN
    value:
  - name: FINNHUB_API_KEY
    value:
  - name: GIN_MODE
    value: release
```

## Running the backend and frontend independently for development

### Backend
1. Run `cd backend`
2. Export the env variables
    ```sh
    export SRS_DATABASE_HOST=mycockroachclusterurl \
    export SRS_DATABASE_PORT=26257 \
    export SRS_DATABASE=mycockroachdb \
    export SRS_DATABASE_USER=mydatabaseuser \
    export SRS_DATABASE_PASSWORD=mydatabasepassword \
    export STOCK_RATING_API_AUTH_TOKEN=apitoken \
    export STOCK_RATING_API_URL=https://apiurl \
    export FINNHUB_API_KEY=finnhubapikey
    ```
3. go run cmd/api/main.go

### Frontend
1. Run `cd frontend`
2. Run `npm run dev`

