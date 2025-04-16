package kinesis

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/service/kinesis"
)

// Consumer wraps the Kinesis client and stream info.
type Consumer struct {
	client     *kinesis.Kinesis
	streamName string
}

// NewConsumer instantiates a new Kinesis consumer.
func NewConsumer(client *kinesis.Kinesis, streamName string) *Consumer {
	return &Consumer{
		client:     client,
		streamName: streamName,
	}
}

// GetRecords retrieves records from Kinesis (simplified example).
func (c *Consumer) GetRecords(ctx context.Context) ([]*kinesis.Record, error) {
	shardIteratorOutput, err := c.client.GetShardIterator(&kinesis.GetShardIteratorInput{
		StreamName:        &c.streamName,
		ShardId:           aws.String("shardId-000000000000"), // In production, iterate across shards.
		ShardIteratorType: aws.String("TRIM_HORIZON"),
	})
	if err != nil {
		return nil, err
	}

	shardIterator := shardIteratorOutput.ShardIterator
	output, err := c.client.GetRecords(&kinesis.GetRecordsInput{
		ShardIterator: shardIterator,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("fetched %d records", len(output.Records))
	return output.Records, nil
}
