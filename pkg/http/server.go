package handlers

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/celestiaorg/torch/config"
	"github.com/celestiaorg/torch/pkg/db/redis"
	"github.com/celestiaorg/torch/pkg/k8s"
	"github.com/celestiaorg/torch/pkg/metrics"
	"github.com/celestiaorg/torch/pkg/nodes"
)

const (
	retryInterval = 10 * time.Second // retryInterval Retry interval in seconds to generate the consensus metric.
)

// GetHttpPort GetPort retrieves the namespace where the service will be deployed
func GetHttpPort() string {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		log.Info("Using the default port: 8080")
		return "8080"
	}

	// Ensure that the provided port is a valid numeric value
	_, err := strconv.Atoi(port)
	if err != nil {
		log.Error("Invalid HTTP_PORT [", os.Getenv("HTTP_PORT"), "] ,using default port 8080")
		return "8080"
	}

	return port
}

// Run initializes the HTTP server, registers metrics for all nodes in the configuration,
// and starts the server.
func Run(cfg config.MutualPeersConfig) {
	// Get http port
	httpPort := GetHttpPort()

	// Set up the HTTP server
	r := mux.NewRouter()
	// Get the routers
	r = Router(r, cfg)
	// Use the middleware
	r.Use(LogRequest)

	// Initialize the config and register the metrics for all nodes
	err := metrics.InitConfig()
	if err != nil {
		log.Errorf("Error initializing metrics: %v", err)
		return
	}

	// Create the server
	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: r,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Listening on: %v", err)
		}
	}()
	log.Info("Server Started...")
	log.Info("Listening on port: " + httpPort)

	// check if Torch has to generate the metric or not.
	BackgroundGenerateHashMetric(cfg)

	// Initialize the goroutine to check the nodes in the queue.
	log.Info("Initializing queues to process the nodes...")
	// Create a new context without timeout as we want to keep this goroutine running forever, if we specify a timeout,
	// it will be canceled at some point.c
	go func() {
		go nodes.ProcessTaskQueue()
	}()

	log.Info("Initializing goroutine to watch over the StatefulSets...")
	// Initialize a goroutine to watch for changes in StatefulSets in the namespace.
	go func() {
		// Call the WatchStatefulSets function and capture any potential error.
		err := k8s.WatchStatefulSets()
		if err != nil {
			// Log an error message if WatchStatefulSets encounters an error.
			log.Error("Error in WatchStatefulSets: ", err)
		}
	}()

	// Initialize the goroutine to add a watcher to the StatefulSets in the namespace.
	log.Info("Initializing Redis consumer")
	go func() {
		nodes.ConsumerInit("k8s")
	}()

	// Check if we already have some multi addresses in the DB and expose them, there might be a situation where Torch
	// get restarted, and we already have the nodes IDs, so we can expose them.
	err = RegisterMetrics(cfg)
	if err != nil {
		log.Error("Couldn't generate the metrics...", err)
	}

	<-done
	log.Info("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Server Shutdown Failed: %v", err)
	}
	log.Info("Server Exited Properly")
}

// BackgroundGenerateHashMetric check if we have defined the consensus in the config, if so, it creates a goroutine
// to generate the metric
func BackgroundGenerateHashMetric(cfg config.MutualPeersConfig) {
	// Check if the config has the consensusNode field defined to generate the metric from the Genesis Hash data.
	if cfg.MutualPeers[0].ConsensusNode != "" {
		// Initialise the goroutine to generate the metric in the background, only if we specify the node in the config.
		go func() {
			log.Info("Consensus node defined to get the first block")
			for {
				err := GenerateHashMetrics(cfg)
				// check if err is nil, if so, that means that Torch was able to generate the metric.
				if err == nil {
					log.Info("Metric generated for the first block...")
					// The metric was successfully generated, stop the retries.
					break
				}

				// Wait for the retry interval before the next execution
				time.Sleep(retryInterval)
			}
		}()
	}
}

// GenerateHashMetrics generates the metric by getting the first block and calculating the days.
func GenerateHashMetrics(cfg config.MutualPeersConfig) error {
	log.Info("Generating the metric for the first block generated...")

	// Get the genesisHash
	blockHash, earliestBlockTime, err := nodes.GenesisHash(cfg.MutualPeers[0].ConsensusNode)
	if err != nil {
		return err
	}

	// check if earliestBlockTime is not empty, otherwise torch skips this process for now.
	err = metrics.WithMetricsBlockHeight(
		blockHash,
		earliestBlockTime,
		cfg.MutualPeers[0].ConsensusNode,
		os.Getenv("POD_NAMESPACE"),
	)
	if err != nil {
		log.Error("Error registering metric block_height_1: ", err)
		return err
	}

	return nil
}

// RegisterMetrics generates and registers the metrics for all nodes in case they already exist in the DB.
func RegisterMetrics(cfg config.MutualPeersConfig) error {
	red := redis.InitRedisConfig()
	// Create a new context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)

	// Make sure to call the cancel function to release resources when you're done
	defer cancel()

	log.Info("Generating metrics from existing nodes...")

	// Adding nodes from config to register the initial metrics
	for _, n := range cfg.MutualPeers {
		for _, no := range n.Peers {
			// checking the node in the DB first
			ma, err := redis.CheckIfNodeExistsInDB(red, ctx, no.NodeName)
			if err != nil {
				log.Error("Error CheckIfNodeExistsInDB : [", no.NodeName, "]", err)
				return err
			}

			// check if the multi address is not empty
			if ma != "" {
				log.Info("Node: [", no.NodeName, "], found in the DB generating metric: ", " [", ma, "]")

				// Register a multi-address metric
				m := metrics.MultiAddrs{
					ServiceName: "torch",
					NodeName:    no.NodeName,
					MultiAddr:   ma,
					Namespace:   no.Namespace,
					Value:       1,
				}
				metrics.RegisterMetric(m)
			}
		}
	}

	return nil
}
