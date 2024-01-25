package config

// MutualPeersConfig represents the configuration structure.
type MutualPeersConfig struct {
	MutualPeers []*MutualPeer `yaml:"mutualPeers"` // MutualPeers list of mutual peers.
}

// MutualPeer represents a mutual peer structure.
type MutualPeer struct {
	ConsensusNode    string `yaml:"consensusNode,omitempty"`    // ConsensusNode name
	Peers            []Peer `yaml:"peers"`                      // Peer list of peers.
	TrustedPeersPath string `yaml:"trustedPeersPath,omitempty"` // TrustedPeersPath specify the path to keep the files
}

// Peer represents a peer structure.
type Peer struct {
	NodeName           string   `yaml:"nodeName"`                     // NodeName name of the sts/deployment
	ServiceName        string   `yaml:"serviceName,omitempty"`        // ServiceName name of the service
	NodeType           string   `yaml:"nodeType"`                     // NodeType specify the type of node
	Namespace          string   `yaml:"namespace,omitempty"`          // Namespace of the node
	ContainerName      string   `yaml:"containerName,omitempty"`      // ContainerName name of the main container
	ContainerSetupName string   `yaml:"containerSetupName,omitempty"` // ContainerSetupName initContainer name
	ConnectsAsEnvVar   bool     `yaml:"connectsAsEnvVar,omitempty"`   // ConnectsAsEnvVar use the value as env var
	ConnectsTo         []string `yaml:"connectsTo,omitempty"`         // ConnectsTo list of nodes that it will connect to
	DnsConnections     []string `yaml:"dnsConnections,omitempty"`     // DnsConnections list of DNS records
	RetryCount         int      `yaml:"retryCount,omitempty"`         // RetryCount number of retries
}
