ACCOUNT_ID?=304955174079
REGION?=eu-west-1

REPO?=metadata-endpoint
IMAGE_NAME?=endpoint-resolver
REGISTRY?=$(ACCOUNT_ID).dkr.ecr.$(REGION).amazonaws.com/$(REPO)
LOG_GROUP?=metadata-logs
ENV_EKS+=REGISTRY IMAGE_NAME
ENV_ECS+= ACCOUNT_ID REGION
ECS_VARS?=$(foreach v,$(ENV_ECS),$(v)='$($(v))' )
KUBE_VARS?=$(foreach v,$(ENV_EKS),$(v)='$($(v))' )


all:
    ifeq ($(BUILD_CLUSTER),EKS)
all: repository build tag push eksdeploy
    else ifeq ($(BUILD_CLUSTER),ECS)
all: repository build tag push loggroup taskdefinition
    else ifeq  ($(BUILD_CLUSTER),ALL)
all: repository build tag push loggroup taskdefinition eksdeploy
    else 
		echo 'Invalid parameter provided.'
    endif


# Image preparation 
repository:
	@echo 'Creating ECR repository $(REPO)...'
	AWS_PAGER="" aws ecr create-repository --repository-name $(REPO) --region $(REGION)

build:
	@echo 'Building image $(IMAGE_NAME)...'
	docker build --no-cache -t $(IMAGE_NAME):latest .

tag:
	@echo 'Tagging image $(IMAGE_NAME)...'
	docker tag $(IMAGE_NAME):latest $(REGISTRY)

push: 
	@echo 'Logging to the registry and pushing image...'
	aws ecr get-login-password --region $(REGION)| docker login --username AWS --password-stdin $(REGISTRY)
	docker push $(REGISTRY)

# ECS Deployment
loggroup:
	@echo 'Create CloudWatch log group $(LOG_GROUP)...'
	aws logs create-log-group --log-group-name $(LOG_GROUP) --region $(REGION)

.ONESHELL:
taskdefinition:
	@echo 'Register task definition in ECS...'
	export AWS_PAGER=""
	$(ECS_VARS) envsubst < deploy/endpoint-metadata-task.template > deploy/endpoint-metadata-task.json
	aws ecs register-task-definition --network-mode none --requires-compatibilities "[\"EC2\"]" --cli-input-json file://deploy/endpoint-metadata-task.json
	aws ecs register-task-definition --network-mode bridge --requires-compatibilities "[\"EC2\"]" --cli-input-json file://deploy/endpoint-metadata-task.json
	aws ecs register-task-definition --network-mode host --requires-compatibilities "[\"EC2\"]" --cli-input-json file://deploy/endpoint-metadata-task.json
	aws ecs register-task-definition --network-mode awsvpc --requires-compatibilities "[\"EC2\",\"FARGATE\"]" --cli-input-json file://deploy/endpoint-metadata-task.json

# EKS Deployment
eksdeploy:
	@echo 'Deploy pod to to EKS...'
	$(KUBE_VARS) envsubst < deploy/pod.yaml | kubectl apply -f -

# Clean up everything
SHELL:=/bin/bash
.ONESHELL:
clean:
	@echo 'Cleaning up all the created resources...'
	export AWS_PAGER=""
	aws logs delete-log-group --log-group-name metadata-logs
	aws ecr delete-repository --repository-name metadata-endpoint --force
	tasks=$$(aws ecs list-task-definitions --family-prefix metadata-endpoint --query "taskDefinitionArns[]" | cut -f7 -d :| sed 's/[^0-9]//g' | sed '1d; $$d')
	while IFS= read -r version; do
		aws ecs deregister-task-definition --task-definition metadata-endpoint:$$version
	done <<< "$$tasks"
	kubectl delete ns meta-store	

.PHONY: repository build tag push loggroup taskdefinition eksdeploy clean

