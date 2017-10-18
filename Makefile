

build:
	GOOS=linux CGO_ENABLED=0 go build .
	docker build -t calum/operator:test .



test:
	kubectl delete pods -l app=oper

reset:
	kubectl delete iamroles --all
	kubectl create -f sample-resource
