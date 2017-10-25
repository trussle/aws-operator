package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/sqs"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"os"
	"time"
)

type Controller struct {
	svc        sqsiface.SQSAPI
	restClient *rest.RESTClient
	scheme     *runtime.Scheme
	regionHost IAWSRegionHost
}

func New(scheme *runtime.Scheme) (*Controller, error) {
	return &Controller{
		scheme:     scheme,
		regionHost: &AWSRegionHost{},
	}, nil
}

type IAWSRegionHost interface {
	ConfigureRegion(c *Controller, region string) error
}
type AWSRegionHost struct {
	IAWSRegionHost
}

func (rc AWSRegionHost) ConfigureRegion(c *Controller, region string) error {
	fmt.Printf("Setting up service with region %s\n", region)

	sess, err := session.NewSession()
	if err != nil {
		fmt.Printf("Error creating AWS session: %v\n", err)
		return err
	}

	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.SharedCredentialsProvider{},
			&credentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		})

	sess, err = session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      &region,
	})

	if err != nil {
		fmt.Printf("Error creating AWS session: %v\n", err)
		return err
	}

	sqsSvc := sqs.New(sess)

	c.svc = sqsSvc
	return nil
}

type controllerMethodDelegate func(obj interface{}) error
type controllerUpdateMethodDelegate func(obj interface{}, new interface{}) error

func wrapUpdateMethod(fn controllerUpdateMethodDelegate) func(obj interface{}, new interface{}) {
	return func(obj interface{}, new interface{}) {
		err := fn(obj, new)
		if err != nil {
			fmt.Printf("%+v", fn)
			return
		}

		return
	}
}

func wrapMethod(fn controllerMethodDelegate) func(obj interface{}) {
	return func(obj interface{}) {

		err := fn(obj)
		if err != nil {
			fmt.Printf("%+v", fn)
			return
		}

		return
	}
}

func (c *Controller) Run(clientset apiextcs.Interface, client *rest.RESTClient) {
	err := Register(clientset, AWSSqsQueue{}, AWSSqsQueueCRDNamePlural, CRDGroup, CRDVersion)
	if err != nil {
		fmt.Printf("error while registering CRD: %v", err)
	}

	c.restClient = client

	time.Sleep(3 * time.Second)
	_, queueInf := cache.NewInformer(
		cache.NewListWatchFromClient(client, AWSSqsQueueCRDNamePlural, os.Getenv("NAMESPACE"), fields.Everything()),
		&AWSSqsQueue{},
		time.Minute*1,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    wrapMethod(c.AddQueue),
			DeleteFunc: wrapMethod(c.DeleteQueue),
			UpdateFunc: wrapUpdateMethod(c.UpdateQueue),
		},
	)

	stop := make(chan struct{})

	go queueInf.Run(stop)
}
