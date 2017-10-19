TAG=$(shell git rev-parse --abbrev-ref HEAD)

build:
	GOOS=linux CGO_ENABLED=0 go build .
	docker build -t teamtrussle/aws-operator:$(TAG) .

release:
	docker push teamtrussle/aws-operator:$(TAG)
