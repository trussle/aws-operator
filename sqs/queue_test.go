package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"testing"
)

type mockedCreateController struct {
	sqsiface.SQSAPI
	Resp interface{}
}

func (mcq mockedCreateController) CreateQueue(in *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	response := mcq.Resp.(sqs.CreateQueueOutput)
	return &response, nil
}

func (mcq mockedCreateController) DeleteQueue(in *sqs.DeleteQueueInput) (*sqs.DeleteQueueOutput, error) {
	response := mcq.Resp.(sqs.DeleteQueueOutput)

	return &response, nil
}

func (mcq mockedCreateController) SetQueueAttributes(in *sqs.SetQueueAttributesInput) (*sqs.SetQueueAttributesOutput, error) {
	response := mcq.Resp.(sqs.SetQueueAttributesOutput)

	return &response, nil
}

type mockedRegionHost struct {
	AWSRegionHost
	Resp AWSRegionHost
}

func (mrh mockedRegionHost) ConfigureRegion(c *Controller, region string) error {
	return nil
}

func mockRestClient() *rest.RESTClient {
	client, err := rest.RESTClientFor(&rest.Config{
		Host: "test",
		ContentConfig: rest.ContentConfig{
			GroupVersion:         &v1.SchemeGroupVersion,
			NegotiatedSerializer: serializer.DirectCodecFactory{CodecFactory: scheme.Codecs},
		},
	})

	if err != nil {
		fmt.Printf("%+v", err)
	}
	return client
}

func TestCreateQueue(t *testing.T) {
	testURL := "http://test/"
	Resp := sqs.CreateQueueOutput{
		QueueUrl: &testURL,
	}
	QueueName := "test"

	controller := &Controller{
		svc:        mockedCreateController{Resp: Resp},
		regionHost: mockedRegionHost{},
		scheme:     runtime.NewScheme(),
		restClient: mockRestClient(),
	}

	controller.AddQueue(&AWSSqsQueue{Spec: AWSSqsQueueSpec{QueueName: QueueName}})
}

func TestDeleteQueue(t *testing.T) {
	testURL := "http://test-queue/"
	Resp := sqs.DeleteQueueOutput{}

	controller := &Controller{
		svc:        mockedCreateController{Resp: Resp},
		regionHost: mockedRegionHost{},
		scheme:     runtime.NewScheme(),
		restClient: mockRestClient(),
	}

	annotations := map[string]string{AWSSqsQueueCRDGroup + "/sqs-autocreated": "true"}

	err := controller.DeleteQueue(&AWSSqsQueue{ObjectMeta: meta_v1.ObjectMeta{Annotations: annotations}, Spec: AWSSqsQueueSpec{QueueURL: testURL}})
	if err != nil {
		t.Fatalf("DeleteQueue failed with error: %+v", err)
		return
	}
}

func TestDeleteQueueNoAnnotation(t *testing.T) {
	testURL := "http://test-queue/"
	Resp := sqs.DeleteQueueOutput{}

	controller := &Controller{
		svc:        mockedCreateController{Resp: Resp},
		regionHost: mockedRegionHost{},
		scheme:     runtime.NewScheme(),
		restClient: mockRestClient(),
	}

	err := controller.DeleteQueue(&AWSSqsQueue{Spec: AWSSqsQueueSpec{QueueURL: testURL}})
	if err == nil {
		t.Fatalf("Expecting DeleteQueue to have returned an error due to missing annotation")
		return
	}
}

func TestUpdateQueue(t *testing.T) {
	testURL := "http://test-queue/"
	Resp := sqs.SetQueueAttributesOutput{}

	controller := &Controller{
		svc:        mockedCreateController{Resp: Resp},
		regionHost: mockedRegionHost{},
		scheme:     runtime.NewScheme(),
		restClient: mockRestClient(),
	}
	DelaySeconds := "20"
	attributes := map[string]*string{"DelaySeconds": &DelaySeconds}
	oldObject := &AWSSqsQueue{}
	err := controller.UpdateQueue(oldObject, &AWSSqsQueue{Spec: AWSSqsQueueSpec{QueueURL: testURL, Attributes: attributes}})
	if err != nil {
		t.Fatalf("%+v", err)
		return
	}
}
