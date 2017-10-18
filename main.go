package main

import (
	"github.com/chronojam/aws-operator/iamrole"
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

	crdcs, _, err := iamrole.NewClient(restCfg)
	if err != nil {
		panic(err.Error())
	}

	cnt, err := iamrole.New()
	if err != nil {
		panic(err.Error())
	}

	go cnt.Run(clientset, crdcs)

	select {}
}
