package sqs

import (
	"fmt"
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
	QueueName  string             `json:"queueName"`
	Region     string             `json:"region"`
	QueueUrl   string             `json:"queueUrl"`
	Attributes map[string]*string `json:"attributes"`
}

type AWSSqsQueueList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []AWSSqsQueue `json:"items"`
}

func (c *Controller) AddQueue(obj interface{}) {
	queue := obj.(*AWSSqsQueue)
	fmt.Printf("Creating queue %s\n", queue.Spec.QueueName)

	c.regionHost.ConfigureRegion(c, queue.Spec.Region)

	input := &sqs.CreateQueueInput{
		QueueName:  &queue.Spec.QueueName,
		Attributes: queue.Spec.Attributes,
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
	if queueCopy.ObjectMeta.Annotations == nil {
		queueCopy.ObjectMeta.Annotations = make(map[string]string)
	}
	queueCopy.ObjectMeta.Annotations[AWSSqsQueueCRDGroup+"/sqs-autocreated"] = "true"

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
	c.regionHost.ConfigureRegion(c, queue.Spec.Region)
	fmt.Printf("Deleting queue %s\n", queue.Spec.QueueUrl)

	input := &sqs.DeleteQueueInput{
		QueueUrl: &queue.Spec.QueueUrl,
	}

	if queue.ObjectMeta.Annotations[AWSSqsQueueCRDGroup+"/sqs-autocreated"] != "true" {
		fmt.Printf("Refusing to delete %s - queue not created by us\n", queue.Spec.QueueUrl)
		return
	}

	_, err := c.svc.DeleteQueue(input)

	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	return
}

func (c *Controller) UpdateQueue(old, new interface{}) {
	queue := new.(*AWSSqsQueue)
	c.regionHost.ConfigureRegion(c, queue.Spec.Region)

	input := &sqs.SetQueueAttributesInput{
		QueueUrl:   &queue.Spec.QueueUrl,
		Attributes: queue.Spec.Attributes,
	}

	_, err := c.svc.SetQueueAttributes(input)

	if err != nil {
		fmt.Printf("UpdateQueue - Encountered error setting queue attributes: %v", err)
		return
	}

	return
}
