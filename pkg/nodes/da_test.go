package nodes

import (
	"reflect"
	"testing"

	"github.com/celestiaorg/torch/config"
)

func TestHasAddrAlready(t *testing.T) {
	t.Parallel()
	type args struct {
		peer      config.Peer
		i         int
		c         string
		addPrefix bool
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "Case 1.0: DNS - Multi address specified",
			args: args{
				peer: config.Peer{
					NodeName:   "da-full-1",
					NodeType:   "da",
					ConnectsTo: []string{"/dns/da-bridge-1/tcp/2121/p2p/12D3KooWH1pTTJR5NXPYs2huVcJ9srmmiyGU4txHm2qgdaUVPYAw,/dns/da-bridge-1/udp/2121/quic-v1/p2p/12D3KooWH1pTTJR5NXPYs2huVcJ9srmmiyGU4txHm2qgdaUVPYAw"},
				},
				i:         0,
				c:         "",
				addPrefix: false,
			},
			want:  "/dns/da-bridge-1/tcp/2121/p2p/12D3KooWH1pTTJR5NXPYs2huVcJ9srmmiyGU4txHm2qgdaUVPYAw,/dns/da-bridge-1/udp/2121/quic-v1/p2p/12D3KooWH1pTTJR5NXPYs2huVcJ9srmmiyGU4txHm2qgdaUVPYAw",
			want1: false,
		},
		{
			name: "Case 1.1: IP - Multi address specified",
			args: args{
				peer: config.Peer{
					NodeName:   "da-full-1",
					NodeType:   "da",
					ConnectsTo: []string{"/ip4/192.168.1.100/tcp/2121/p2p/12D3KooWH1pTTJR5NXPYs2huVcJ9srmmiyGU4txHm2qgdaUVPYAw,/ip4/192.168.1.100/udp/2121/quic-v1/p2p/12D3KooWH1pTTJR5NXPYs2huVcJ9srmmiyGU4txHm2qgdaUVPYAw"},
				},
				i:         0,
				c:         "",
				addPrefix: false,
			},
			want:  "/ip4/192.168.1.100/tcp/2121/p2p/12D3KooWH1pTTJR5NXPYs2huVcJ9srmmiyGU4txHm2qgdaUVPYAw,/ip4/192.168.1.100/udp/2121/quic-v1/p2p/12D3KooWH1pTTJR5NXPYs2huVcJ9srmmiyGU4txHm2qgdaUVPYAw",
			want1: false,
		},
		{
			name: "Case 2: No multi address specified - one node",
			args: args{
				peer: config.Peer{
					NodeName:   "da-full-1",
					NodeType:   "da",
					ConnectsTo: []string{"da-bridge-1"},
				},
				i:         0,
				c:         "da-bridge-1",
				addPrefix: false,
			},
			want:  "da-bridge-1",
			want1: false,
		},
		{
			name: "Case 3: No multi address specified - more than one node",
			args: args{
				peer: config.Peer{
					NodeName:   "da-full-1",
					NodeType:   "da",
					ConnectsTo: []string{"da-bridge-1", "da-bridge-2"},
				},
				i:         1,
				c:         "da-bridge-1,da-bridge-2",
				addPrefix: false,
			},
			want:  "da-bridge-1,da-bridge-2",
			want1: false,
		},
		{
			name: "Case 4: Empty ConnectsTo slice",
			args: args{
				peer: config.Peer{
					NodeName:   "da-full-1",
					NodeType:   "da",
					ConnectsTo: []string{},
				},
				i:         0,
				c:         "",
				addPrefix: false,
			},
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.args.peer.ConnectsTo == nil || len(tt.args.peer.ConnectsTo) <= tt.args.i {
				t.Skip("Skipping test case with invalid index or empty ConnectsTo")
			}
			got, got1 := VerifyAndUpdateMultiAddress(tt.args.peer, tt.args.i, tt.args.c, tt.args.addPrefix)
			if got != tt.want {
				t.Errorf("VerifyAndUpdateMultiAddress() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("VerifyAndUpdateMultiAddress() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSetDaNodeDefault(t *testing.T) {
	t.Parallel()
	type args struct {
		peer config.Peer
	}
	tests := []struct {
		name string
		args args
		want config.Peer
	}{
		{
			name: "Case 1: Tests default values",
			args: args{
				peer: config.Peer{
					NodeName: "da-full-1",
					NodeType: "da",
				},
			},
			want: config.Peer{
				NodeName:           "da-full-1",
				NodeType:           "da",
				ContainerName:      "da",
				ContainerSetupName: "da-setup",
				Namespace:          ns,
				ConnectsAsEnvVar:   false,
				ConnectsTo:         nil,
				DnsConnections:     nil,
			},
		},
		{
			name: "Case 2: Tests default values already specified",
			args: args{
				peer: config.Peer{
					NodeName:           "da-bridge-1",
					NodeType:           "da",
					ContainerName:      "da",
					ContainerSetupName: "da-setup",
					Namespace:          "custom-namespace",
				},
			},
			want: config.Peer{
				NodeName:           "da-bridge-1",
				NodeType:           "da",
				ContainerName:      "da",
				ContainerSetupName: "da-setup",
				Namespace:          "custom-namespace",
				ConnectsAsEnvVar:   false,
				ConnectsTo:         nil,
				DnsConnections:     nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SetDaNodeDefault(tt.args.peer)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetDaNodeDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	t.Parallel()
	type args struct {
		input     string
		maxLength int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Case 1: Tests valid node ID: 12D3KooWPB3thXCYyr6Jid49d5DDaRL63inzVagaQswCcgUARg5W - 52",
			args: args{
				input:     "12D3KooWPB3thXCYyr6Jid49d5DDaRL63inzVagaQswCcgUARg5W",
				maxLength: 52,
			},
			want: "12D3KooWPB3thXCYyr6Jid49d5DDaRL63inzVagaQswCcgUARg5W",
		},
		{
			name: "Case 2: Tests not valid node ID.",
			args: args{
				input:     "12D3KooWPB3thXCYyr6Jid49d5DDaRL63inzVagaQswCcgUARg5W12D3KooWPB3thXCYyr6Jid49d5DDaRL63inzVagaQswCcgUARg5W",
				maxLength: 52,
			},
			want: "12D3KooWPB3thXCYyr6Jid49d5DDaRL63inzVagaQswCcgUARg5W",
		},
		{
			name: "Case 3: Tests valid node ID: 12D3KooWMGSh8pLvQYn5zYcdRhVfNAcMZrDt71iyq6eSVtrgjKb8 - 52",
			args: args{
				input:     "12D3KooWMGSh8pLvQYn5zYcdRhVfNAcMZrDt71iyq6eSVtrgjKb8",
				maxLength: 52,
			},
			want: "12D3KooWMGSh8pLvQYn5zYcdRhVfNAcMZrDt71iyq6eSVtrgjKb8",
		},
		{
			name: "Case 4: Tests not valid node ID.",
			args: args{
				input:     "12D3KooWMGSh8pLvQYn5zYcdRhVfNAcMZrDt71iyq6eSVtrgjKb812D3KooWMGSh8pLvQYn5zYcdRhVfNAcMZrDt71iyq6eSVtrgjKb8",
				maxLength: 52,
			},
			want: "12D3KooWMGSh8pLvQYn5zYcdRhVfNAcMZrDt71iyq6eSVtrgjKb8",
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := TruncateString(tt.args.input, tt.args.maxLength)
			if err != nil {
				t.Fatalf("TruncateString() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Errorf("TruncateString() = %v, want %v", got, tt.want)
			}
		})
	}
}
