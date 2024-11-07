package nodes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/celestiaorg/torch/config"
	"github.com/celestiaorg/torch/pkg/db/redis"
	"github.com/celestiaorg/torch/pkg/k8s"
	"github.com/celestiaorg/torch/pkg/metrics"
)

const (
	errRemoteCommand = "Error executing remote command: "
	timeoutDuration  = 60 * time.Second // timeoutDuration we specify the max time to run the func.
	nodeIdMaxLength  = 52               // nodeIdMaxLength Specify the max length for the nodes ids.
)

var (
	daContainerSetupName = "da-setup"                     // daContainerSetupName initContainer that we use to configure the nodes.
	daContainerName      = "da"                           // daContainerName container name which the pod runs.
	fPathDA              = "/tmp/celestia-config/TP-ADDR" // fPathDA path to the file where Torch will write.
	ns                   = k8s.GetCurrentNamespace()      // ns namespace of the node.
)

// SetDaNodeDefault sets all the default values in case they are empty
func SetDaNodeDefault(peer config.Peer) config.Peer {
	if peer.ContainerSetupName == "" {
		peer.ContainerSetupName = daContainerSetupName
	}
	if peer.ContainerName == "" {
		peer.ContainerName = daContainerName
	}
	if peer.Namespace == "" {
		peer.Namespace = ns
	}
	peer.ConnectsAsEnvVar = false
	peer.ConnectsTo = nil
	peer.DnsConnections = nil
	return peer
}

// SetupDANodeWithConnections configure a DA node with connections
func SetupDANodeWithConnections(peer config.Peer) error {
	red := redis.InitRedisConfig()
	// Create a new context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	connString := ""
	addPrefix := true

	// Make sure to call the cancel function to release resources when you're done
	defer cancel()

	// read the connection list
	for index, nodeName := range peer.ConnectsTo {
		log.Info(peer.NodeName, " , connection: [", index, "] to node: [", nodeName, "]")

		// checking the node in the DB first
		ma, err := redis.CheckIfNodeExistsInDB(red, ctx, nodeName)
		if err != nil {
			log.Error("Error CheckIfNodeExistsInDB for full-node: [", peer.NodeName, "]", err)
			return err
		}

		// check if the MA is already in the config
		ma, addPrefix = VerifyAndUpdateMultiAddress(peer, index, ma, addPrefix)

		// if the node is not in the db, then we generate it
		if ma == "" {
			log.Info("Node ", "["+nodeName+"]"+" NOT found in DB, let'nodeName generate it")
			ma, err = GenerateNodeIdAndSaveIt(peer, peer.ConnectsTo[index], red, ctx)
			if err != nil {
				log.Error("Error GenerateNodeIdAndSaveIt for full-node: [", peer.NodeName, "]", err)
				return err
			}
		}

		// if we have the address already, lets continue the process, otherwise, means we couldn't get the node id
		if ma != "" && addPrefix {
			// adding the node prefix
			ma, err = SetIdPrefix(peer, ma, index)
			if err != nil {
				log.Error("Error SetIdPrefix for full-node: [", peer.NodeName, "]", err)
				return err
			}
			log.Info("Peer connection prefix: ", ma)
		}

		// check the connection index and concatenate it in case we have more than one node
		if index > 0 {
			connString = connString + "," + ma
		} else {
			connString = ma
		}

		// validate the MA, must start with /ip4/ || /dns/
		if !strings.HasPrefix(ma, "/ip4/") && !strings.HasPrefix(ma, "/dns/") {
			errorMessage := fmt.Sprintf("Error generating the MultiAddress, must begin with /ip4/ || /dns/: [%s]", ma)
			log.Error(errorMessage)
			return errors.New(errorMessage)
		}

		log.Info("Registering metric for node: [", nodeName, "]")

		// Register a multi-address metric
		m := metrics.MultiAddrs{
			ServiceName: "torch",
			NodeName:    nodeName,
			MultiAddr:   ma,
			Namespace:   peer.Namespace,
			Value:       1,
		}
		metrics.RegisterMetric(m)

		// get the command to write in a file and execute the command against the node
		command := k8s.WriteToFile(connString, fPathDA)
		output, err := k8s.RunRemoteCommand(
			peer.NodeName,
			peer.ContainerSetupName,
			k8s.GetCurrentNamespace(),
			command)
		if err != nil {
			log.Error(errRemoteCommand, err)
			return err
		}

		log.Info("MultiAddr for node ", peer.NodeName, " is: [", output, "]")

		log.Info("Adding node to the queue: [", peer.NodeName, "]")
		go AddToQueue(peer)
	}

	return nil
}

// VerifyAndUpdateMultiAddress checks if the configuration contains a Multi Address at the specified index
// and updates it if found. It returns the verified Multi Address and a boolean indicating if an update was performed.
func VerifyAndUpdateMultiAddress(peer config.Peer, index int, currentAddr string, addPrefix bool) (string, bool) {
	// verify that we have the multi addr already specify in the config
	if strings.Contains(peer.ConnectsTo[index], "dns") || strings.Contains(peer.ConnectsTo[index], "ip4") {
		// Use the address from the configuration
		currentAddr = peer.ConnectsTo[index]
		addPrefix = false
	}
	return currentAddr, addPrefix
}

// SetIdPrefix generates the prefix depending on dns or ip
func SetIdPrefix(peer config.Peer, c string, i int) (string, error) {
	// check if we are using DNS or IP
	if len(peer.DnsConnections) > 0 {
		c = "/dns/" + peer.DnsConnections[i] + "/tcp/2121/p2p/" + c
		c = "," + "/dns/" + peer.DnsConnections[i] + "/udp/2121/quic-v1/p2p/" + c
	} else {
		comm := k8s.GetNodeIP()
		output, err := k8s.RunRemoteCommand(
			peer.ConnectsTo[i],
			peer.ContainerName,
			k8s.GetCurrentNamespace(),
			comm)
		if err != nil {
			log.Error(errRemoteCommand, err)
			return "", err
		}
		log.Info("command - ip is: ", output)
		c = output + c
	}
	return c, nil
}

// GenerateNodeIdAndSaveIt generates the node id and store it
func GenerateNodeIdAndSaveIt(
	pod config.Peer,
	connNode string,
	red *redis.RedisClient,
	ctx context.Context,
) (string, error) {
	// Generate the command and run it against the connection node + it's running container
	command := k8s.CreateTrustedPeerCommand()
	output, err := k8s.RunRemoteCommand(
		connNode,
		pod.ContainerName,
		k8s.GetCurrentNamespace(),
		command)
	if err != nil {
		log.Error(errRemoteCommand, err)
		return "", err
	}

	// if the output of the generation is not empty, that means that we could generate the node id successfully, so let's
	// store it into the DB.
	if output != "" {
		log.Info("Adding pod id to Redis: ", connNode, " [", output, "] ")

		// check that the node id generate has the right length
		output, err = TruncateString(output, nodeIdMaxLength)
		if err != nil {
			log.Error("Error TruncateString: ", err)
			return "", err
		}

		// save node in redis
		err = redis.SetNodeId(connNode, red, ctx, output)
		if err != nil {
			log.Error("Error SetNodeId: ", err)
			return "", err
		}
	} else {
		log.Error("Output is empty for pod: ", " [", pod.NodeName, "] ")
		return "", err
	}

	return output, nil
}

// TruncateString receives and input and a maxLength and returns a string with the size specified.
func TruncateString(input string, maxLength int) (string, error) {
	if len(input) == maxLength {
		return input, nil
	}
	if len(input) < maxLength {
		log.Error("Error: The node id received is not valid, too short: , ", " - [", len(input), "]", " - [", input, "]")
		return input, errors.New("error: The node id received is not valid")
	}

	log.Info("Node ID with bigger size found: ", input)

	return input[:maxLength], nil
}
