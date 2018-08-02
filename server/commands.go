package main

import (
	"fmt"
	"net/http"

	"github.com/0xef53/gorilla-rpc-example/common"
)

func (x *RPC) GetServerSummary(r *http.Request, args *common.ServerSummaryQuery, result *common.ServerSummary) error {
	if args.ServerName != "superserver" {
		return fmt.Errorf("Unknown server name: %s", args.ServerName)
	}

	s := common.ServerSummary{ServerName: args.ServerName}

	s.Qemu.Major = 2
	s.Qemu.Minor = 11
	s.Qemu.Micro = 3

	*result = s

	return nil
}
