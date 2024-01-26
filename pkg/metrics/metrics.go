package metrics

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Get the meter from the global meter provider with the name "torch".
var meter = otel.GetMeterProvider().Meter("torch")

// MultiAddrs represents the information for a Multi Addresses.
type MultiAddrs struct {
	ServiceName string  // ServiceName Name of the service associated with the Multi Addresses.
	NodeName    string  // NodeName Name of the node.
	MultiAddr   string  // MultiAddr Multi Addresses value.
	Namespace   string  // Namespace where the service is deployed.
	Value       float64 // Value to be observed for the Multi Addresses.
}

// WithMetricsMultiAddress creates a callback function to observe metrics for multiple Multi Addresses.
func WithMetricsMultiAddress(multiAddrs []MultiAddrs) error {
	log.Info("registering metric: ", multiAddrs)
	// Create a Float64ObservableGauge named "Multi Addresses" with a description for the metric.
	multiAddressesGauge, err := meter.Float64ObservableGauge(
		"multiaddr",
		metric.WithDescription("Torch - MultiAddresses"),
	)
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}

	// Define the callback function that will be called periodically to observe metrics.
	callback := func(ctx context.Context, observer metric.Observer) error {
		for _, ma := range multiAddrs {
			// Create labels with attributes for each Multi Addresses.
			labels := metric.WithAttributes(
				attribute.String("service_name", ma.ServiceName),
				attribute.String("node_name", ma.NodeName),
				attribute.String("multiaddress", ma.MultiAddr),
				attribute.String("namespace", ma.Namespace),
			)
			// Observe the float64 value for the current Multi Addresses with the associated labels.
			observer.ObserveFloat64(multiAddressesGauge, ma.Value, labels)
		}

		return nil
	}

	// Register the callback with the meter and the Float64ObservableGauge.
	_, err = meter.RegisterCallback(callback, multiAddressesGauge)
	return err
}

// BlockHeight represents the information for the block height 1.
type BlockHeight struct {
	ServiceName string  // ServiceName Name of the service associated with the multi-address.
	BlockHeight string  // BlockHeight height of the block.
	Value       float64 // Value to be observed for the multi-address.
}

// WithMetricsBlockHeight creates a callback function to observe metrics for block_height_1.
// consensus-node:26657/block?height=1
func WithMetricsBlockHeight(blockHeight, earliestBlockTime, serviceName, namespace string) error {
	log.Info("registering metric: ", blockHeight)
	// Create a Float64ObservableGauge named "block_height_1" with a description for the metric.
	blockHeightGauge, err := meter.Float64ObservableGauge(
		"block_height_1",
		metric.WithDescription("Torch - BlockHeight"),
	)
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}

	// Calculate the days that the chain is live.
	daysRunning, err := calculateDaysDifference(earliestBlockTime)
	if err != nil {
		log.Error("ERROR: ", err)
		return err
	}

	callback := func(ctx context.Context, observer metric.Observer) error {
		// Define the callback function that will be called periodically to observe metrics.
		// Create labels with attributes for each block_height_1.
		labels := metric.WithAttributes(
			attribute.String("service_name", serviceName),
			attribute.String("block_height_1", blockHeight),
			attribute.String("earliest_block_time", earliestBlockTime),
			attribute.Int("days_running", daysRunning),
			attribute.String("namespace", namespace),
		)
		// Observe the float64 value for the current block_height_1 with the associated labels.
		observer.ObserveFloat64(blockHeightGauge, 1, labels)

		return nil
	}

	// Register the callback with the meter and the Float64ObservableGauge.
	_, err = meter.RegisterCallback(callback, blockHeightGauge)
	return err
}

// calculateDaysDifference based on the date received, returns the number of days since this day.
func calculateDaysDifference(inputTimeString string) (int, error) {
	layout := "2006-01-02T15:04:05.999999999Z"
	inputTime, err := time.Parse(layout, inputTimeString)
	if err != nil {
		log.Error("Error parsing time: [", inputTimeString, "]", err)
		return -1, err
	}

	currentTime := time.Now()
	timeDifference := currentTime.Sub(inputTime)
	daysDifference := int(timeDifference.Hours() / 24)

	return daysDifference, nil
}

// LoadBalancer represents the information for a load balancer.
type LoadBalancer struct {
	ServiceName      string  // ServiceName Name of the service associated with the load balancer.
	LoadBalancerName string  // LoadBalancerName Name of the load balancer.
	LoadBalancerIP   string  // LoadBalancerIP IP address of the load balancer.
	Namespace        string  // Namespace where the service is deployed.
	Value            float64 // Value to be observed for the load balancer.
}

// WithMetricsLoadBalancer creates a callback function to observe metrics for multiple load balancers.
func WithMetricsLoadBalancer(loadBalancers []LoadBalancer) error {
	// Create a Float64ObservableGauge named "load_balancer" with a description for the metric.
	loadBalancersGauge, err := meter.Float64ObservableGauge(
		"load_balancer",
		metric.WithDescription("Torch - Load Balancers"),
	)
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}

	// Define the callback function that will be called periodically to observe metrics.
	callback := func(ctx context.Context, observer metric.Observer) error {
		for _, lb := range loadBalancers {
			// Create labels with attributes for each load balancer.
			labels := metric.WithAttributes(
				attribute.String("service_name", lb.ServiceName),
				attribute.String("load_balancer_name", lb.LoadBalancerName),
				attribute.String("load_balancer_ip", lb.LoadBalancerIP),
				attribute.String("namespace", lb.Namespace),
			)
			// Observe the float64 value for the current load balancer with the associated labels.
			observer.ObserveFloat64(loadBalancersGauge, lb.Value, labels)
		}

		return nil
	}

	// Register the callback with the meter and the Float64ObservableGauge.
	_, err = meter.RegisterCallback(callback, loadBalancersGauge)
	if err != nil {
		log.Error("Error registering callback: ", err)
		return err
	}

	return nil
}

// ConsensusNodeMetric represents the information for consensus node metrics.
type ConsensusNodeMetric struct {
	NodeName  string // NodeName is the name of the node.
	NodeID    string // NodeID is the ID of the node.
	Namespace string // Namespace of the node.
}

// RegisterConsensusNodeMetric creates and registers metrics for consensus nodes.
func RegisterConsensusNodeMetric(nodeID, nodeName, namespace string) error {
	log.Info("Registering metric for consensus node: ", nodeName)

	// Create an ObservableGauge for consensus node metrics.
	consensusNodeGauge, err := meter.Float64ObservableGauge(
		"consensus_node_ids_metric",
		metric.WithDescription("Metric for Consensus Node IDs"),
	)
	if err != nil {
		log.Error("Error creating metric: ", err)
		return err
	}

	callback := func(ctx context.Context, observer metric.Observer) error {
		// Define the callback function that will be called periodically to observe metrics.
		labels := metric.WithAttributes(
			attribute.String("node_name", nodeName),
			attribute.String("node_id", nodeID),
			attribute.String("namespace", namespace),
		)
		// Observe the value for current consensus node metrics with associated labels.
		observer.ObserveFloat64(consensusNodeGauge, 1, labels)

		return nil
	}

	// Register the callback with the meter and the ObservableGauge.
	_, err = meter.RegisterCallback(callback, consensusNodeGauge)
	if err != nil {
		log.Error("Error registering callback: ", err)
		return err
	}

	return nil
}
