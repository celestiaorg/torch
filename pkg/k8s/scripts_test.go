package k8s

import (
	"reflect"
	"testing"
)

// case1 common message.
const case1 = "Case 1: Successfully script generated."

// TestCreateFileWithEnvVar validates the node types and their path
func TestCreateFileWithEnvVar(t *testing.T) {
	t.Parallel()

	type args struct {
		nodeToFile string
		nodeType   string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Case 1: Check [consensus] nodes",
			args: args{
				nodeToFile: "/home/celestia/config/TP-ADDR",
				nodeType:   "consensus",
			},
			want: []string{"sh", "-c", `
#!/bin/sh
echo -n "/home/celestia/config/TP-ADDR" > "/home/celestia/config/TP-ADDR"`},
		},
		{
			name: "Case 2: Check [da] nodes",
			args: args{
				nodeToFile: "/tmp/CONSENSUS_NODE_SERVICE",
				nodeType:   "da",
			},
			want: []string{"sh", "-c", `
#!/bin/sh
echo -n "/tmp/CONSENSUS_NODE_SERVICE" > "/tmp/CONSENSUS_NODE_SERVICE"`},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CreateFileWithEnvVar(tt.args.nodeToFile, tt.args.nodeType)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateFileWithEnvVar() = \ngot:  %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}

// TestCreateTrustedPeerCommand checks the script to generate the multi address.
func TestCreateTrustedPeerCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want []string
	}{
		{
			name: case1,
			want: []string{"sh", "-c", `
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

echo -n "${TP_ADDR}" >> "/tmp/TP-ADDR"
cat "/tmp/TP-ADDR"
`},
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CreateTrustedPeerCommand()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateTrustedPeerCommand() = \ngot:  %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}

// TestGetNodeIP gets the IP of the node and add it to a file
func TestGetNodeIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want []string
	}{
		{
			name: case1,
			want: []string{"sh", "-c", `
#!/bin/sh
echo -n "/ip4/$(ifconfig | grep -oE 'inet addr:([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)' | grep -v '127.0.0.1' | awk '{print substr($2, 6)}')/tcp/2121/p2p/,/ip4/$(ifconfig | grep -oE 'inet addr:([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)' | grep -v '127.0.0.1' | awk '{print substr($2, 6)}')/udp/2121/quic-v1/p2p/" >> "/tmp/NODE_IP"
cat "/tmp/NODE_IP"`},
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GetNodeIP()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNodeIP() = \ngot:  %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}

// TestWriteToFile writes content to a file
func TestWriteToFile(t *testing.T) {
	t.Parallel()

	type args struct {
		content string
		file    string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: case1,
			args: args{
				content: "THIS IS A TEST",
				file:    "/tmp/test_file",
			},
			want: []string{"sh", "-c", `
#!/bin/sh
echo -n "THIS IS A TEST" > "/tmp/test_file"
cat "/tmp/test_file"`},
		},
		{
			name: "Case 2: Successfully script generated.",
			args: args{
				content: "content in file",
				file:    "/tmp/file_with_content",
			},
			want: []string{"sh", "-c", `
#!/bin/sh
echo -n "content in file" > "/tmp/file_with_content"
cat "/tmp/file_with_content"`},
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := WriteToFile(tt.args.content, tt.args.file)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WriteToFile() = \ngot:  %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}
