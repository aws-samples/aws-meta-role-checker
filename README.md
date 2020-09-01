# aws-meta-role-checker
A tool to run on your AWS ECS and/or EKS clusters with the purpose to check whether account keys are accessible through the metadata service and identify whether any over-privileged roles are being used run the workloads. This way you can take measures in order to switch to less privileged roles or block access to metadata or credential endpoints altogether.

This tool will build a container, store the image in AWS ECR and deploy it in an already running EKS cluster. For ECS, it will register tasks for none, bridge, host and awsvpc (EC2 and Fargate) modes so that you can then freely choose to run any or all of them in any existing or new ECS cluster.

The logs containing the exposed data are written to CloudWatch for ECS and to the container logs for EKS.


## Prerequisites
It is assumed that docker, kubectl and aws-cli v2 are installed and correctly configured in your PATH.

NOTE: envsubst is not available by default on mac
To install it use:
```
brew install gettext
brew link --force gettext
```


## Building
Before executing the build, make sure ACCOUNT_ID and REGION variables in Makefile are set.

Replace the ACCOUNT_ID? and REGION? values with your own:
```
ACCOUNT_ID?=1112222333444
REGION?=eu-west-1
```

To build the artifacts on an existing EKS cluster, run the following command:
```shell
make BUILD_CLUSTER=EKS
```

To build the necassary artifacts for an ECS cluster, to which you can manually assign the resulting tasks, run the following command:
```shell
make BUILD_CLUSTER=ECS
```

To build the artifacts on an existing EKS cluster as well for ECS, run the following command:
```shell
make BUILD_CLUSTER=ALL
```

## Viewing logs
To view logs for the EKS deployment just run:
```shell
kubectl logs -n meta-store eks-metadata-endpoint-[xxxxxx]
```

For ECS deployments, search the generated logs prefixed by metadata/containerlogs in the metadata CloudWatch log group.

## Cleaning
To clean up the built resources, run:

```shell
make clean
```

