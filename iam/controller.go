package iam

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"

	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"os"
	"time"
)

func New() (*Controller, error) {
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
		})

	sess, err = session.NewSession(&aws.Config{
		Credentials: creds,
	})
	if err != nil {
		return nil, err
	}

	iamSvc := iam.New(sess)

	return &Controller{
		svc: iamSvc,
	}, nil
}

type Controller struct {
	svc *iam.IAM
}

func (c *Controller) Run(clientset apiextcs.Interface, client *rest.RESTClient) {
	err := Register(clientset, AWSIamPolicy{}, AWSIamPolicyCRDNamePlural, CRDGroup, CRDVersion)
	if err != nil {
		fmt.Printf("error while registring CRD: %v", err)
	}

	err = Register(clientset, AWSIamRole{}, AWSIamRoleCRDNamePlural, CRDGroup, CRDVersion)
	if err != nil {
		fmt.Printf("error while registring CRD: %v", err)
	}

	time.Sleep(3 * time.Second)
	_, policyinf := cache.NewInformer(
		cache.NewListWatchFromClient(client, AWSIamPolicyCRDNamePlural, os.Getenv("NAMESPACE"), fields.Everything()),
		&AWSIamPolicy{},
		time.Minute*1,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.AddPolicy,
			DeleteFunc: c.DeletePolicy,
			UpdateFunc: c.UpdatePolicy,
		},
	)

	_, roleinf := cache.NewInformer(
		cache.NewListWatchFromClient(client, AWSIamRoleCRDNamePlural, os.Getenv("NAMESPACE"), fields.Everything()),
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
}
