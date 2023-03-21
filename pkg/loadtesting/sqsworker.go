package loadtesting

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SqsWorker struct {
	queueName string
	queueUrl  *string
	Client    *sqs.Client
}

func NewSqsWorker(ctx context.Context, queueName string) SqsWorker {
	s := SqsWorker{
		queueName: queueName,
	}
	client, err := s.getSqsClient(ctx)
	if err != nil {
		panic(err)
	} else {
		s.Client = client
		s.queueUrl, err = s.GetQueueUrl(ctx, s.Client, s.queueName)
		if err != nil {
			panic(err)
		}
	}
	return s
}

func (s SqsWorker) Process(ctx context.Context, input []string) (*string, error) {
	return s.SendMessage(ctx, s.Client, *s.queueUrl, input[0], map[string]types.MessageAttributeValue{})
}

func (s SqsWorker) getSqsClient(ctx context.Context) (*sqs.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sqs client, configuration error %v", err)
	}

	return sqs.NewFromConfig(cfg), nil

}

func (s SqsWorker) GetQueueUrl(ctx context.Context, client *sqs.Client, queueName string) (*string, error) {
	// Get URL of queue
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	}

	result, err := client.GetQueueUrl(ctx, gQInput)
	if err != nil {
		return nil, err
	}

	return result.QueueUrl, nil
}

func (s SqsWorker) SendMessage(ctx context.Context, client *sqs.Client, queueUrl string, msgBody string, attributes map[string]types.MessageAttributeValue) (*string, error) {
	sMInput := &sqs.SendMessageInput{
		MessageBody: &msgBody,
		QueueUrl:    &queueUrl,
		//DelaySeconds:            10,
		MessageAttributes: attributes,
		//MessageDeduplicationId:  new(string),
		//MessageGroupId:          new(string),
		//MessageSystemAttributes: map[string]types.MessageSystemAttributeValue{},
	}

	resp, err := client.SendMessage(ctx, sMInput)

	if err != nil {
		return nil, err
	}

	return resp.MessageId, nil
}
