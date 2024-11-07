package k8s

import (
	"fmt"
)

var (
	trustedPeerFile          = "/tmp/TP-ADDR"
	trustedPeerFileConsensus = "/home/celestia/config/TP-ADDR"
	trustedPeerFileDA        = "/tmp/CONSENSUS_NODE_SERVICE"
	nodeIpFile               = "/tmp/NODE_IP"
	cmd                      = `$(ifconfig | grep -oE 'inet addr:([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)' | grep -v '127.0.0.1' | awk '{print substr($2, 6)}')`
	trustedPeerPrefix        = "/ip4/" + cmd + "/tcp/2121/p2p/,/ip4/" + cmd + "/udp/2121/quic-v1/p2p/"
)

// CreateFileWithEnvVar creates the file in the FS with the node to connect.
func CreateFileWithEnvVar(nodeToFile, nodeType string) []string {
	f := ""
	if nodeType == "consensus" {
		f = trustedPeerFileConsensus
	}
	if nodeType == "da" {
		f = trustedPeerFileDA
	}

	script := fmt.Sprintf(`
#!/bin/sh
echo -n "%[2]s" > "%[1]s"`, f, nodeToFile)

	return []string{"sh", "-c", script}
}

// CreateTrustedPeerCommand generates the command for creating trusted peers.
// we have to use the shell script because we can only get the token and the
// nodeID from the node itself.
func CreateTrustedPeerCommand() []string {
	script := fmt.Sprintf(`
#!/bin/sh
# generate the token
export AUTHTOKEN=$(celestia bridge auth admin --node.store /home/celestia)

# remove the first warning line...
export AUTHTOKEN=$(echo $AUTHTOKEN|rev|cut -d' ' -f1|rev)

# make the request and parse the response
TP_ADDR=$(wget --header="Authorization: Bearer $AUTHTOKEN" \
   --header="Content-Type: application/json" \
   --post-data='{"jsonrpc":"2.0","id":0,"method":"p2p.Info","params":[]}' \
   --output-document - \
   http://localhost:26658 | grep -o '"ID":"[^"]*"' | sed 's/"ID":"\([^"]*\)"/\1/')

echo -n "${TP_ADDR}" >> "%[1]s"
cat "%[1]s"
`, trustedPeerFile, trustedPeerPrefix)

	return []string{"sh", "-c", script}
}

// GetNodeIP adds the node IP to a file.
func GetNodeIP() []string {
	script := fmt.Sprintf(`
#!/bin/sh
echo -n "%[2]s" > "%[1]s"
cat "%[1]s"`, nodeIpFile, trustedPeerPrefix)

	return []string{"sh", "-c", script}
}

// WriteToFile writes content into a file.
func WriteToFile(content, file string) []string {
	script := fmt.Sprintf(`
#!/bin/sh
echo -n "%[1]s" > "%[2]s"
cat "%[2]s"`, content, file)

	return []string{"sh", "-c", script}
}
