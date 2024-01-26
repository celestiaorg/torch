package nodes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/celestiaorg/torch/config"
	"github.com/celestiaorg/torch/pkg/k8s"
)

var (
	consContainerSetupName = "consensus-setup"         // consContainerSetupName initContainer that we use to configure the nodes.
	consContainerName      = "consensus"               // consContainerName container name which the pod runs.
	namespace              = k8s.GetCurrentNamespace() // namespace of the node.
	KeyResult              = "result"                  // KeyResult result field in the JSON response
	KeyBlockID             = "block_id"                // KeyBlockID block_id field within the 'result' field
	KeyHash                = "hash"                    // KeyHash hash field within the 'block_id' field
	KeyNodeInfo            = "node_info"               // KeyNodeInfo node_info field within the 'result' field
	KeyID                  = "id"                      // KeyID id field within the 'node_info' field

)

// SetConsNodeDefault sets all the default values in case they are empty
func SetConsNodeDefault(peer config.Peer) config.Peer {
	if peer.ContainerSetupName == "" {
		peer.ContainerSetupName = consContainerSetupName
	}
	if peer.ContainerName == "" {
		peer.ContainerName = consContainerName
	}
	if peer.Namespace == "" {
		peer.Namespace = namespace
	}
	return peer
}

// GenesisHash connects to the specified consensus node, makes a request to the API,
// and retrieves information about the genesis block including its hash and time.
func GenesisHash(consensusNode string) (string, string, error) {
	url := fmt.Sprintf("http://%s:26657/block?height=1", consensusNode)
	jsonResponse, err := makeAPIRequest(url)
	if err != nil {
		return "", "", err
	}

	blockIDHash, ok := jsonResponse[KeyResult].(map[string]interface{})[KeyBlockID].(map[string]interface{})[KeyHash].(string)
	if !ok {
		log.Error("Unable to access .block_id.hash")
		return "", "", errors.New("error accessing block ID hash")
	}

	blockTime, ok := jsonResponse["result"].(map[string]interface{})["block"].(map[string]interface{})["header"].(map[string]interface{})["time"].(string)
	if !ok {
		log.Error("Unable to access .block.header.time")
		return "", "", errors.New("error accessing block time")
	}

	return blockIDHash, blockTime, nil
}

// ConsensusNodesIDs connects to the specified consensus node, makes a request to the API,
// and retrieves the node ID from the status response.
func ConsensusNodesIDs(consensusNode string) (string, error) {
	url := fmt.Sprintf("http://%s:26657/status?", consensusNode)
	jsonResponse, err := makeAPIRequest(url)
	if err != nil {
		return "", err
	}

	nodeID, ok := jsonResponse[KeyResult].(map[string]interface{})[KeyNodeInfo].(map[string]interface{})[KeyID].(string)
	if !ok {
		log.Error("Unable to access .result.node_info.id")
		return "", errors.New("error accessing node ID")
	}

	log.Info("Consensus Node [", consensusNode, "] ID: [", nodeID, "]")

	return nodeID, nil
}

// makeAPIRequest handles the common task of making an HTTP request to a given URL
// and parsing the JSON response. It returns a map representing the JSON response or an error.
func makeAPIRequest(url string) (map[string]interface{}, error) {
	response, err := http.Get(url)
	if err != nil {
		log.Error("Error making the request: ", err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Error("Non-OK response:", response.Status)
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error("Error reading response body:", err)
		return nil, err
	}

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(bodyBytes, &jsonResponse)
	if err != nil {
		log.Error("Error parsing JSON:", err)
		return nil, err
	}

	return jsonResponse, nil
}
