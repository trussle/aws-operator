package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"testing"
)

type mockedCreateController struct {
	sqsiface.SQSAPI
	Resp sqs.CreateQueueOutput
}

func (mcq mockedCreateController) CreateQueue(in *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	return &mcq.Resp, nil
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
	testUrl := "http://test/"
	Resp := sqs.CreateQueueOutput{
		QueueUrl: &testUrl,
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
