PROJECT_ID=$(shell gcloud config get-value project)
DATASET=cloud_trace_spans

service-account.json:
	gcloud iam service-accounts keys create --iam-account=trace-agent@${PROJECT_ID}.iam.gserviceaccount.com service-account.json

env:
	@echo export GOOGLE_APPLICATION_CREDENTIALS=$(shell pwd)/service-account.json
	@echo export PROJECT_ID=${PROJECT_ID}
	@echo export DATASET=${DATASET}

.PHONY: service-account
service-account:
	gcloud iam service-accounts create trace-agent --display-name "Service Account for trace cli tools"

.PHONY: service-account-permissions
service-account-permissions:
	gcloud projects add-iam-policy-binding ${PROJECT_ID} \
	    --member serviceAccount:trace-agent@${PROJECT_ID}.iam.gserviceaccount.com \
	    --role roles/cloudtrace.user
	gcloud projects add-iam-policy-binding ${PROJECT_ID} \
	    --member serviceAccount:trace-agent@${PROJECT_ID}.iam.gserviceaccount.com \
	    --role roles/bigquery.dataEditor

.PHONY: dataset
dataset:
	bq --project ${PROJECT_ID} \
		mk \
	    --dataset \
	    --default_table_expiration 3600 \
	    --description "Cloud Trace Spans" \
	    ${PROJECT_ID}:${DATASET}

.PHONY: clean-dataset
clean-dataset:
	bq --project ${PROJECT_ID} \
		rm \
	    -r \
	    -d ${PROJECT_ID}:${DATASET}
