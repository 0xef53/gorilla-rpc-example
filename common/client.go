package common

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"runtime"

	"golang.org/x/net/http2"

	jsonrpc "github.com/gorilla/rpc/v2/json2"
)

type TlsClient struct {
	client   *http.Client
	endpoint string
}

func NewTlsClient(addr, endpoint, certFile, keyFile string) (*TlsClient, error) {
	tlsConfig, err := TlsConfig(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	if err := http2.ConfigureTransport(transport); err != nil {
		return nil, err
	}

	c := TlsClient{
		client:   &http.Client{Transport: transport},
		endpoint: "https://" + addr + ":9395" + endpoint,
	}

	return &c, nil
}

func (c *TlsClient) Run(method string, args interface{}, res interface{}) error {
	message, err := jsonrpc.EncodeClientRequest(method, args)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(message))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if res == nil {
		return nil
	}

	if err := jsonrpc.DecodeClientResponse(resp.Body, &res); err != nil {
		panic(err)
	}

	return nil
}

type Client struct {
	client   *http.Client
	endpoint string
}

func NewClient(sockpath, endpoint string) (*Client, error) {
	if sockpath[0] == '@' && runtime.GOOS == "linux" {
		sockpath = sockpath + string(0)
	}

	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", sockpath)
		},
	}

	c := Client{
		client:   &http.Client{Transport: transport},
		endpoint: "http://127.0.0.1" + endpoint,
	}

	return &c, nil
}

func (c *Client) Run(method string, args interface{}, res interface{}) error {
	message, err := jsonrpc.EncodeClientRequest(method, args)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(message))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if res == nil {
		return nil
	}

	if err := jsonrpc.DecodeClientResponse(resp.Body, &res); err != nil {
		panic(err)
	}

	return nil
}
