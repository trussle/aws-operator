

## Development

Make sure you have Glide installed.  Run `glide i -v` to build the vendor directory.

1. Install [MiniKube](https://kubernetes.io/docs/getting-started-guides/minikube/) to get a local Kubernetes environment
2. `$ eval $(minikube docker-env)`
3. `$ make build`
4. `$ make dev` to install the Operator.  It will mount your local ~/.aws/credentials for testing - for production container the credentials are retrieved from the EC2 environment instead.

