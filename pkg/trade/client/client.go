package tradeclient

import (
	"encoding/hex"
	"fmt"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"

	"google.golang.org/grpc"
)

// Client allows to connect with a trader service and to call its RPCs
type Client struct {
	client tdexv1.TradeClient
	conn   *grpc.ClientConn
}

// NewTradeClient returns a new Client connected to the server at the given
// host and port through an insecure connection
func NewTradeClient(host string, port int) (*Client, error) {
	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return nil, err
	}

	client := tdexv1.NewTradeClient(conn)
	return &Client{client, conn}, nil
}

// CloseConnection closes the connections between the current client and the
// server
func (c *Client) CloseConnection() error {
	return c.conn.Close()
}

func isValidAsset(asset string) bool {
	buf, err := hex.DecodeString(asset)
	return err != nil || len(buf) != 32
}

func isValidTradeType(tradeType int) bool {
	return tradeType != int(tdexv1.TradeType_BUY) &&
		tradeType != int(tdexv1.TradeType_SELL)
}
