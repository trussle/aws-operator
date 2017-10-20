package main

import (
	"github.com/trussle/aws-operator/iam"
	"github.com/trussle/aws-operator/sqs"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
)

func main() {
	restCfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := apiextcs.NewForConfig(restCfg)
	if err != nil {
		panic(err.Error())
	}

	crdcs, _, err := iam.NewClient(restCfg)
	if err != nil {
		panic(err.Error())
	}

	sqsCrd, sqsScheme, err := sqs.NewClient(restCfg)
	if err != nil {
		panic(err.Error())
	}

	iamController, err := iam.New()
	if err != nil {
		panic(err.Error())
	}

	sqsController, err := sqs.New(sqsScheme)
	if err != nil {
		panic(err.Error())
	}

	go sqsController.Run(clientset, sqsCrd)
	go iamController.Run(clientset, crdcs)

	select {}
}
