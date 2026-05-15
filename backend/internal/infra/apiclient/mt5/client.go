// Package mt5 provides a gRPC client for the MT5 gateway.
// Connects via Connect-RPC (protoc-gen-connect-go) to the external mt5grpc service.
package mt5

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	mt5grpc "github.com/antclaw/antclaw/gen/go/third_party/mt5grpc"
	"github.com/antclaw/antclaw/gen/go/third_party/mt5grpc/mt5grpcconnect"
)

// Client wraps the mt5grpc Connection and MT5 services.
type Client struct {
	conn    mt5grpcconnect.ConnectionClient
	mt5     mt5grpcconnect.MT5Client
	service mt5grpcconnect.ServiceClient
	baseURL string
}

// NewClient creates a new MT5 gateway client.
// baseURL is the MT5 gRPC gateway address, e.g. "http://mt5-gateway:8080".
func NewClient(baseURL string) *Client {
	httpClient := &http.Client{}

	return &Client{
		conn:    mt5grpcconnect.NewConnectionClient(httpClient, baseURL),
		mt5:     mt5grpcconnect.NewMT5Client(httpClient, baseURL),
		service: mt5grpcconnect.NewServiceClient(httpClient, baseURL),
		baseURL: baseURL,
	}
}

// Connect establishes a connection to an MT5 account and returns a session token.
func (c *Client) Connect(ctx context.Context, login uint64, password, host string, port int32) (string, error) {
	req := connect.NewRequest(&mt5grpc.ConnectRequest{
		User:     login,
		Password: password,
		Host:     host,
		Port:     port,
	})
	resp, err := c.conn.Connect(ctx, req)
	if err != nil {
		return "", fmt.Errorf("mt5 connect: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return "", fmt.Errorf("mt5 connect error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// ConnectEx establishes a connection using server name (alternative to host:port).
func (c *Client) ConnectEx(ctx context.Context, login uint64, password, server string) (string, error) {
	req := connect.NewRequest(&mt5grpc.ConnectExRequest{
		User:     login,
		Password: password,
		Server:   server,
	})
	resp, err := c.conn.ConnectEx(ctx, req)
	if err != nil {
		return "", fmt.Errorf("mt5 connect ex: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return "", fmt.Errorf("mt5 connect ex error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// ConnectProxy establishes a connection with proxy support.
func (c *Client) ConnectProxy(ctx context.Context, login uint64, password, host string, port int32, proxy *mt5grpc.ConnectProxyRequest) (string, error) {
	if proxy == nil {
		proxy = &mt5grpc.ConnectProxyRequest{}
	}
	proxy.User = login
	proxy.Password = password
	proxy.Host = host
	proxy.Port = port
	req := connect.NewRequest(proxy)
	resp, err := c.conn.ConnectProxy(ctx, req)
	if err != nil {
		return "", fmt.Errorf("mt5 connect proxy: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return "", fmt.Errorf("mt5 connect proxy error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// Disconnect closes a connection by token.
func (c *Client) Disconnect(ctx context.Context, token string) error {
	req := connect.NewRequest(&mt5grpc.DisconnectRequest{Id: token})
	resp, err := c.conn.Disconnect(ctx, req)
	if err != nil {
		return fmt.Errorf("mt5 disconnect: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return fmt.Errorf("mt5 disconnect error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return nil
}

// CheckConnect verifies the connection is alive.
func (c *Client) CheckConnect(ctx context.Context, token string) (string, error) {
	req := connect.NewRequest(&mt5grpc.CheckConnectRequest{Id: token})
	resp, err := c.conn.CheckConnect(ctx, req)
	if err != nil {
		return "", fmt.Errorf("mt5 check connect: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return "", fmt.Errorf("mt5 check connect error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// AccountSummary returns the account trading summary.
func (c *Client) AccountSummary(ctx context.Context, token string) (*mt5grpc.AccountSummary, error) {
	req := connect.NewRequest(&mt5grpc.AccountSummaryRequest{Id: token})
	resp, err := c.mt5.AccountSummary(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("mt5 account summary: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return nil, fmt.Errorf("mt5 account summary error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// Search returns broker companies matching the given name.
func (c *Client) Search(ctx context.Context, company string) ([]*mt5grpc.Company, error) {
	req := connect.NewRequest(&mt5grpc.SearchRequest{Company: company})
	resp, err := c.service.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("mt5 search: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return nil, fmt.Errorf("mt5 search error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}

// PingHost tests connectivity to a host:port.
func (c *Client) PingHost(ctx context.Context, host string, port int32) (int32, error) {
	req := connect.NewRequest(&mt5grpc.PingHostRequest{Host: host, Port: &port})
	resp, err := c.service.PingHost(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("mt5 ping host: %w", err)
	}
	if resp.Msg.Error != nil && resp.Msg.Error.Code != 0 {
		return 0, fmt.Errorf("mt5 ping host error [%d]: %s", resp.Msg.Error.Code, resp.Msg.Error.Message)
	}
	return resp.Msg.Result, nil
}
