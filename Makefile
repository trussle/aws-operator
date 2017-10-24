TAG=$(shell git rev-parse --abbrev-ref HEAD)

build:
	GOOS=linux CGO_ENABLED=0 go build .
	docker build -t teamtrussle/aws-operator:$(TAG) .

unit-tests:
	go test -v ./...

dev:
	- cat sample-operator.yaml | sed 's#$$(CREDS_FILE)#$(shell echo ~/.aws/credentials)#g' | kubectl --context=minikube create -f -

release:
	docker push teamtrussle/aws-operator:$(TAG)
