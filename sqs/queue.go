package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AWSSqsQueueCRDName       = "awssqsqueue"
	AWSSqsQueueCRDNamePlural = "awssqsqueues"
	AWSSqsQueueCRDGroup      = "trussle.com"
)

type AWSSqsQueue struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               AWSSqsQueueSpec `json:"spec"`
}

type AWSSqsQueueSpec struct {
	QueueName string `json:"queueName"`
	Region    string `json:"region"`
	QueueUrl  string `json:"queueUrl"`
}

type AWSSqsQueueList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []AWSSqsQueue `json:"items"`
}

func (c *Controller) SetupService(region string) {
	fmt.Printf("Setting up service with region %s\n", region)

	sess, err := session.NewSession()
	if err != nil {
		fmt.Printf("Error creating AWS session: %v\n", err)
		return
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
		return
	}

	sqsSvc := sqs.New(sess)

	c.svc = sqsSvc
}

func (c *Controller) AddQueue(obj interface{}) {
	queue := obj.(*AWSSqsQueue)
	fmt.Printf("Creating queue %s\n", queue.Spec.QueueName)

	c.SetupService(queue.Spec.Region)

	input := &sqs.CreateQueueInput{
		QueueName: &queue.Spec.QueueName,
	}

	response, err := c.svc.CreateQueue(input)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if response.QueueUrl == nil {
		fmt.Printf("Encountered empty QueueUrl on AddQueue response")
		return
	}
	copiedObject, err := c.scheme.Copy(queue)

	if err != nil {
		fmt.Printf("Failed to create a copy of queue object: %v\n", err)
		return
	}

	queueCopy := copiedObject.(*AWSSqsQueue)
	queueCopy.Spec.QueueUrl = *response.QueueUrl

	err = c.restClient.Put().
		Name(queue.ObjectMeta.Name).
		Namespace(queue.ObjectMeta.Namespace).
		Resource(AWSSqsQueueCRDNamePlural).
		Body(queueCopy).
		Do().
		Error()

	if err != nil {
		fmt.Printf("Failed to PUT resource to kube-api %v\n", err)
		return
	}

	return
}

func (c *Controller) DeleteQueue(obj interface{}) {
	queue := obj.(*AWSSqsQueue)
	c.SetupService(queue.Spec.Region)
	fmt.Printf("Deleting queue %s\n", queue.Spec.QueueUrl)

	input := &sqs.DeleteQueueInput{
		QueueUrl: &queue.Spec.QueueUrl,
	}

	_, err := c.svc.DeleteQueue(input)

	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	return
}

func (c *Controller) UpdateQueue(old, new interface{}) {
	fmt.Printf("Update not implemented")
}
