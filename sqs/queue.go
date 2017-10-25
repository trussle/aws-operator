package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/sqs"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AWSSqsQueueCRDName       = "awssqsqueue"  // The name of this CRD
	AWSSqsQueueCRDNamePlural = "awssqsqueues" // Pluralised name of the CRD
	AWSSqsQueueCRDGroup      = "trussle.com"  // CRD Group name
)

type AWSSqsQueue struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               AWSSqsQueueSpec `json:"spec"`
}

type AWSSqsQueueSpec struct {
	QueueName  string             `json:"queueName"`
	Region     string             `json:"region"`
	QueueURL   string             `json:"queueUrl"`
	Attributes map[string]*string `json:"attributes"`
}

type AWSSqsQueueList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []AWSSqsQueue `json:"items"`
}

func (c *Controller) AddQueue(obj interface{}) error {
	queue := obj.(*AWSSqsQueue)
	fmt.Printf("Creating queue %s\n", queue.Spec.QueueName)

	err := c.regionHost.ConfigureRegion(c, queue.Spec.Region)

	if err != nil {
		fmt.Printf("Error calling ConfigureRegion: %+v", err)
		return err
	}
	input := &sqs.CreateQueueInput{
		QueueName:  &queue.Spec.QueueName,
		Attributes: queue.Spec.Attributes,
	}

	response, err := c.svc.CreateQueue(input)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}

	if response.QueueUrl == nil {
		return fmt.Errorf("Encountered empty QueueUrl on AddQueue response")
	}
	copiedObject, err := c.scheme.Copy(queue)

	if err != nil {
		fmt.Printf("Failed to create a copy of queue object: %v\n", err)
		return err
	}

	queueCopy := copiedObject.(*AWSSqsQueue)
	queueCopy.Spec.QueueURL = *response.QueueUrl
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
		return err
	}

	return nil
}

func (c *Controller) DeleteQueue(obj interface{}) error {
	queue := obj.(*AWSSqsQueue)
	err := c.regionHost.ConfigureRegion(c, queue.Spec.Region)
	if err != nil {
		fmt.Printf("Error calling ConfigureRegion: %+v", err)
		return err
	}

	input := &sqs.DeleteQueueInput{
		QueueUrl: &queue.Spec.QueueURL,
	}

	if queue.ObjectMeta.Annotations[AWSSqsQueueCRDGroup+"/sqs-autocreated"] != "true" {
		return fmt.Errorf("Refusing to delete %s - queue not created by us\n", queue.Spec.QueueURL)
	}

	_, deleteQueueErr := c.svc.DeleteQueue(input)

	if deleteQueueErr != nil {
		fmt.Printf(deleteQueueErr.Error())
		return deleteQueueErr
	}

	return nil
}

func (c *Controller) UpdateQueue(old, new interface{}) error {
	queue := new.(*AWSSqsQueue)
	err := c.regionHost.ConfigureRegion(c, queue.Spec.Region)

	if err != nil {
		fmt.Printf("Error calling ConfigureRegion: %+v", err)
		return err
	}

	input := &sqs.SetQueueAttributesInput{
		QueueUrl:   &queue.Spec.QueueURL,
		Attributes: queue.Spec.Attributes,
	}

	_, setQueueAttributesErr := c.svc.SetQueueAttributes(input)

	if setQueueAttributesErr != nil {
		fmt.Printf("UpdateQueue - Encountered error setting queue attributes: %v", setQueueAttributesErr)
		return setQueueAttributesErr
	}

	return nil
}
