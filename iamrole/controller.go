package iamrole

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
	err := Register(clientset)
	if err != nil {
		fmt.Printf("error while registring CRD: %v", err)
	}

	time.Sleep(3 * time.Second)
	_, inf := cache.NewInformer(
		cache.NewListWatchFromClient(client, CRDNamePlural, os.Getenv("NAMESPACE"), fields.Everything()),
		&IamRole{},
		time.Minute*1,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.Add,
			DeleteFunc: c.Delete,
			UpdateFunc: c.Update,
		},
	)

	stop := make(chan struct{})

	go inf.Run(stop)
}

func (c *Controller) Add(obj interface{}) {
	role := obj.(*IamRole)

	iroles, err := c.svc.ListRoles(&iam.ListRolesInput{})
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	for _, ir := range iroles.Roles {
		if *ir.RoleName == role.Spec.Name {
			fmt.Printf("Skipping due to existing IAM role: %s\n", role.Spec.Name)
			return
		}
	}

	_, err = c.svc.CreateRole(&iam.CreateRoleInput{
		RoleName:                 &role.Spec.Name,
		AssumeRolePolicyDocument: &role.Spec.RolePolicyDocument,
	})
	if err != nil {
		fmt.Printf("Failed to create role: %s\n", role.Spec.Name)
		return
	}

	ipolicies, err := c.svc.ListPolicies(&iam.ListPoliciesInput{})
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	for _, ip := range ipolicies.Policies {
		if *ip.PolicyName == role.Spec.Name {
			fmt.Printf("Skipping policy due to existing policy: %s\n", role.Spec.Name)
			return
		}
	}
	_, err = c.svc.CreatePolicy(&iam.CreatePolicyInput{
		PolicyName:     &role.Spec.Name,
		PolicyDocument: &role.Spec.PolicyDocument,
	})
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	fmt.Printf("Add: %s", obj)
}

func (c *Controller) Delete(obj interface{}) {
	fmt.Printf("Delete: %s", obj)
}

func (c *Controller) Update(old, new interface{}) {
	fmt.Printf("Update: %s %s", old, new)
}
