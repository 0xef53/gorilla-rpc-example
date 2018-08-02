package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/0xef53/gorilla-rpc-example/common"
)

func DoViaHttp2(addr string) error {
	fmt.Println("Via HTTP/2 connection:")

	//client, err := common.NewTlsClient("https://"+addr+":9395/rpc/v1", "client.crt", "client.key")
	client, err := common.NewTlsClient(addr, "/rpc/v1", "client.crt", "client.key")
	if err != nil {
		return fmt.Errorf("NewClient() error: %s", err)
	}

	res := &common.ServerSummary{}

	if err := client.Run("RPC.GetServerSummary", &common.ServerSummaryQuery{"superserver"}, &res); err != nil {
		return fmt.Errorf("client.Run() error: %s", err)
	}

	b, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		return fmt.Errorf("json.Marshal() error: %s", err)
	}

	fmt.Println(string(b))

	return nil
}

func DoViaUnixSocket(addr string) error {
	fmt.Println("Via unix socket connection:")

	client, err := common.NewClient("@/tmp/server.sock", "/rpc/v1")
	if err != nil {
		return fmt.Errorf("NewClient() error: %s", err)
	}

	res := &common.ServerSummary{}

	if err := client.Run("RPC.GetServerSummary", &common.ServerSummaryQuery{"superserver"}, &res); err != nil {
		return fmt.Errorf("client.Run() error: %s", err)
	}

	b, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		return fmt.Errorf("json.Marshal() error: %s", err)
	}

	fmt.Println(string(b))

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client IPADDR")
		return
	}

	addr := os.Args[1]

	if err := DoViaHttp2(addr); err != nil {
		fmt.Println("DoViaHttp2() error:", err)
	}

	fmt.Printf("\n\n")

	if err := DoViaUnixSocket(addr); err != nil {
		fmt.Println("DoViaHttp2() error:", err)
	}

}
