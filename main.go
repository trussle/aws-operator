package main

import (
	"github.com/trussle/aws-operator/iam"
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

	cnt, err := iam.New()
	if err != nil {
		panic(err.Error())
	}

	go cnt.Run(clientset, crdcs)

	select {}
}
