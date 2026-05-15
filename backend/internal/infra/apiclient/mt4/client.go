// Package mt4 provides a gRPC client for the MT4 gateway.
// Connects via Connect-RPC (protoc-gen-connect-go) to the external mt4grpc service.
package mt4

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	mt4grpc "github.com/antclaw/antclaw/gen/go/third_party/mt4grpc"
	"github.com/antclaw/antclaw/gen/go/third_party/mt4grpc/mt4grpcconnect"
)

// Client wraps the mt4grpc Connection and MT4 services.
type Client struct {
	conn       mt4grpcconnect.ConnectionClient
	mt4        mt4grpcconnect.MT4Client
	service    mt4grpcconnect.ServiceClient
	baseURL    string
}

// NewClient creates a new MT4 gateway client.
// baseURL is the MT4 gRPC gateway address, e.g. "http://mt4-gateway:8080".
func NewClient(baseURL string) *Client {
	httpClient := &http.Client{}

	return &Client{
		conn:    mt4grpcconnect.NewConnectionClient(httpClient, baseURL),
		mt4:     mt4grpcconnect.NewMT4Client(httpClient, baseURL),
		service: mt4grpcconnect.NewServiceClient(httpClient, baseURL),
		baseURL: baseURL,
	}
}

// Connect establishes a connection to an MT4 account and returns a session token.
func (c *Client) Connect(ctx context.Context, login int32, password, host string, port int32) (string, error) {
	req := connect.NewRequest(&mt4grpc.ConnectRequest{
		User:     login,
		Password: password,
		Host:     host,
		Port:     port,
	})
	resp, err := c.conn.Connect(ctx, req)
	if err != nil {
		return "", fmt.Errorf("mt4 connect: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return "", fmt.Errorf("mt4 connect error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// ConnectProxy establishes a connection with proxy support.
func (c *Client) ConnectProxy(ctx context.Context, login int32, password, host string, port int32, proxy *mt4grpc.ConnectProxyRequest) (string, error) {
	if proxy == nil {
		proxy = &mt4grpc.ConnectProxyRequest{}
	}
	proxy.User = login
	proxy.Password = password
	proxy.Host = host
	proxy.Port = port
	req := connect.NewRequest(proxy)
	resp, err := c.conn.ConnectProxy(ctx, req)
	if err != nil {
		return "", fmt.Errorf("mt4 connect proxy: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return "", fmt.Errorf("mt4 connect proxy error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// Disconnect closes a connection by token.
func (c *Client) Disconnect(ctx context.Context, token string) error {
	req := connect.NewRequest(&mt4grpc.DisconnectRequest{Id: token})
	resp, err := c.conn.Disconnect(ctx, req)
	if err != nil {
		return fmt.Errorf("mt4 disconnect: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return fmt.Errorf("mt4 disconnect error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return nil
}

// CheckConnect verifies the connection is alive.
func (c *Client) CheckConnect(ctx context.Context, token string) (string, error) {
	req := connect.NewRequest(&mt4grpc.CheckConnectRequest{Id: token})
	resp, err := c.conn.CheckConnect(ctx, req)
	if err != nil {
		return "", fmt.Errorf("mt4 check connect: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return "", fmt.Errorf("mt4 check connect error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// AccountSummary returns the account trading summary.
func (c *Client) AccountSummary(ctx context.Context, token string) (*mt4grpc.AccountSummary, error) {
	req := connect.NewRequest(&mt4grpc.AccountSummaryRequest{Id: token})
	resp, err := c.mt4.AccountSummary(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("mt4 account summary: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return nil, fmt.Errorf("mt4 account summary error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// IsInvestor checks if the account is in investor (read-only) mode.
func (c *Client) IsInvestor(ctx context.Context, token string) (bool, error) {
	req := connect.NewRequest(&mt4grpc.IsInvestorRequest{Id: token})
	resp, err := c.mt4.IsInvestor(ctx, req)
	if err != nil {
		return false, fmt.Errorf("mt4 is investor: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return false, fmt.Errorf("mt4 is investor error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// Search returns broker companies matching the given name.
func (c *Client) Search(ctx context.Context, company string) ([]*mt4grpc.Company, error) {
	req := connect.NewRequest(&mt4grpc.SearchRequest{Company: company})
	resp, err := c.service.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("mt4 search: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return nil, fmt.Errorf("mt4 search error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// PingHost tests connectivity to a host:port.
func (c *Client) PingHost(ctx context.Context, host string, port int32) (int32, error) {
	req := connect.NewRequest(&mt4grpc.PingHostRequest{Host: host, Port: port})
	resp, err := c.service.PingHost(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("mt4 ping host: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return 0, fmt.Errorf("mt4 ping host error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}
