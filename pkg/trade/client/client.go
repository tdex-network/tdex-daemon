package tradeclient

import (
	"fmt"

	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client allows to connect with a trader service and to call its RPCs
type Client struct {
	client tdexv2.TradeServiceClient
	conn   *grpc.ClientConn
}

// NewTradeClient returns a new Client connected to the server at the given
// host and port through an insecure connection
func NewTradeClient(host string, port int) (*Client, error) {
	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return nil, err
	}

	client := tdexv2.NewTradeServiceClient(conn)
	return &Client{client, conn}, nil
}

// CloseConnection closes the connections between the current client and the
// server
func (c *Client) CloseConnection() error {
	return c.conn.Close()
}
