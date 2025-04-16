package main

import (
	"context"
	"log"
	"math"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsKinesis "github.com/aws/aws-sdk-go/service/kinesis"
	"iot-insighthub/pkg/kinesis"
	"iot-insighthub/pkg/telemetry"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ingestCounter counts processed records.
var ingestCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "ingested_events_total",
	Help: "Total number of ingested events",
})

// init registers Prometheus metrics.
func init() {
	prometheus.MustRegister(ingestCounter)
}

// startPprofServer starts an HTTP server that exposes pprof endpoints.
func startPprofServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	log.Println("Starting pprof server on :6060")
	if err := http.ListenAndServe(":6060", mux); err != nil {
		log.Fatalf("pprof server failed: %v", err)
	}
}

// listShards lists all the shard IDs for a given stream.
func listShards(kc *awsKinesis.Kinesis, streamName string) ([]*string, error) {
	input := &awsKinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
	}
	result, err := kc.DescribeStream(input)
	if err != nil {
		return nil, err
	}

	var shardIDs []*string
	for _, shard := range result.StreamDescription.Shards {
		shardIDs = append(shardIDs, shard.ShardId)
	}
	return shardIDs, nil
}

// processShard continuously fetches records from a given shard.
// It applies exponential backoff on failures and sends records to recordChan.
func processShard(ctx context.Context, kc *awsKinesis.Kinesis, streamName string, shardID *string, recordChan chan<- *awsKinesis.Record) {
	// Get the initial shard iterator.
	input := &awsKinesis.GetShardIteratorInput{
		StreamName:        aws.String(streamName),
		ShardId:           shardID,
		ShardIteratorType: aws.String("TRIM_HORIZON"),
	}
	output, err := kc.GetShardIterator(input)
	if err != nil {
		log.Printf("Error getting shard iterator for shard %s: %v", *shardID, err)
		return
	}
	iterator := output.ShardIterator

	// Initialize exponential backoff variables.
	backoff := 1 * time.Second
	const maxBackoff = 30 * time.Second

	for {
		if iterator == nil {
			log.Printf("Shard iterator expired for shard %s", *shardID)
			break
		}

		getRecordsInput := &awsKinesis.GetRecordsInput{
			ShardIterator: iterator,
			Limit:         aws.Int64(100), // Up to 100 records per call.
		}
		recordsOutput, err := kc.GetRecords(getRecordsInput)
		if err != nil {
			log.Printf("Error fetching records from shard %s: %v", *shardID, err)
			time.Sleep(backoff)
			// Increase backoff exponentially, up to maxBackoff.
			backoff = time.Duration(math.Min(float64(maxBackoff), float64(backoff)*2))
			continue
		}
		// Reset backoff after success.
		backoff = 1 * time.Second

		// Send each record to the processing channel (with non-blocking send due to buffering).
		for _, record := range recordsOutput.Records {
			select {
			case recordChan <- record:
			case <-ctx.Done():
				return
			}
		}

		iterator = recordsOutput.NextShardIterator

		// Slow down the loop if no records were returned.
		if len(recordsOutput.Records) == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// recordWorker processes records received on recordChan.
// It leverages the telemetry service to handle each record.
func recordWorker(ctx context.Context, workerID int, recordChan <-chan *awsKinesis.Record, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case record, ok := <-recordChan:
			if !ok {
				// Channel closed; exit worker.
				return
			}
			if err := telemetry.ProcessRecord(record); err != nil {
				log.Printf("Worker %d: Error processing record: %v", workerID, err)
			} else {
				ingestCounter.Inc()
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	ctx := context.Background()

	// Start Prometheus metrics server on port 9090.
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Prometheus metrics server running on :9090")
		log.Fatal(http.ListenAndServe(":9090", nil))
	}()

	// Start pprof server for runtime profiling on port 6060.
	go startPprofServer()

	// Create AWS session and Kinesis client.
	sess := session.Must(session.NewSession())
	kc := awsKinesis.New(sess)
	streamName := "YourKinesisStreamName"

	// List all shards in the stream.
	shardIDs, err := listShards(kc, streamName)
	if err != nil {
		log.Fatalf("Error listing shards: %v", err)
	}
	log.Printf("Found %d shards", len(shardIDs))

	// Create a buffered channel for records for backpressure.
	recordChan := make(chan *awsKinesis.Record, 1000)

	// Start a worker pool to process records concurrently.
	numWorkers := 10
	var workerWG sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		workerWG.Add(1)
		go recordWorker(ctx, i, recordChan, &workerWG)
	}

	// Spawn a goroutine for each shard to consume records concurrently.
	var shardWG sync.WaitGroup
	for _, shardID := range shardIDs {
		shardWG.Add(1)
		go func(shardID *string) {
			defer shardWG.Done()
			processShard(ctx, kc, streamName, shardID, recordChan)
		}(shardID)
	}

	// Wait for shard processing to complete (in production, this may run indefinitely).
	shardWG.Wait()

	// Close the record channel and wait for workers to finish.
	close(recordChan)
	workerWG.Wait()
}
