package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/sqs"

	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"os"
	"time"
)

func New(scheme *runtime.Scheme) (*Controller, error) {
	return &Controller{
		scheme: scheme,
	}, nil
}

type Controller struct {
	svc        *sqs.SQS
	restClient *rest.RESTClient
	scheme     *runtime.Scheme
}

func (c *Controller) Run(clientset apiextcs.Interface, client *rest.RESTClient) {
	err := Register(clientset, AWSSqsQueue{}, AWSSqsQueueCRDNamePlural, CRDGroup, CRDVersion)
	if err != nil {
		fmt.Printf("error while registring CRD: %v", err)
	}

	c.restClient = client

	time.Sleep(3 * time.Second)
	_, queueInf := cache.NewInformer(
		cache.NewListWatchFromClient(client, AWSSqsQueueCRDNamePlural, os.Getenv("NAMESPACE"), fields.Everything()),
		&AWSSqsQueue{},
		time.Minute*1,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.AddQueue,
			DeleteFunc: c.DeleteQueue,
			UpdateFunc: c.UpdateQueue,
		},
	)

	stop := make(chan struct{})

	go queueInf.Run(stop)
}
