package iam

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"

	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	defaultWaitTime = 3 * time.Second
)

type Controller struct {
	namespace string
	svc       *iam.IAM
	clientSet apiextcs.Interface
	client    *rest.RESTClient
	stop      chan chan struct{}
}

func New(namespace string, clientSet apiextcs.Interface, client *rest.RESTClient) (*Controller, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		},
	)

	sess, err = session.NewSession(&aws.Config{
		Credentials: creds,
	})
	if err != nil {
		return nil, err
	}

	return &Controller{
		namespace: namespace,
		svc:       iam.New(sess),
		clientSet: clientSet,
		client:    client,
		stop:      make(chan chan struct{}),
	}, nil
}

func (c *Controller) Run() error {
	err := Register(c.clientSet, AWSIamPolicy{}, AWSIamPolicyCRDNamePlural, CRDGroup, CRDVersion)
	if err != nil {
		return errors.Wrap(err, "registering CRD")
	}

	err = Register(c.clientSet, AWSIamRole{}, AWSIamRoleCRDNamePlural, CRDGroup, CRDVersion)
	if err != nil {
		return errors.Wrap(err, "registering CRD")
	}

	time.Sleep(defaultWaitTime)

	_, policyinf := cache.NewInformer(
		cache.NewListWatchFromClient(c.client, AWSIamPolicyCRDNamePlural, c.namespace, fields.Everything()),
		&AWSIamPolicy{},
		time.Minute*1,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.AddPolicy,
			DeleteFunc: c.DeletePolicy,
			UpdateFunc: c.UpdatePolicy,
		},
	)

	_, roleinf := cache.NewInformer(
		cache.NewListWatchFromClient(c.client, AWSIamRoleCRDNamePlural, c.namespace, fields.Everything()),
		&AWSIamRole{},
		time.Minute*1,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.AddRole,
			DeleteFunc: c.DeleteRole,
			UpdateFunc: c.UpdateRole,
		},
	)

	stop := make(chan struct{})

	go roleinf.Run(stop)
	go policyinf.Run(stop)

	for {
		select {
		case q := <-c.stop:
			stop <- struct{}{}
			close(q)
			return nil
		}
	}
}

func (c *Controller) Stop() {
	q := make(chan struct{})
	c.stop <- q
	<-q
}
