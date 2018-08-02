package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	jsonrpc "github.com/gorilla/rpc/v2/json2"

	"golang.org/x/net/http2"

	"github.com/0xef53/gorilla-rpc-example/common"
)

type RPC struct{}

func serve(ctx context.Context, addr net.IP, rcvr *RPC) error {
	var bindAddr string

	if addr.To4() != nil {
		bindAddr = addr.String()
	} else {
		bindAddr = "[" + addr.String() + "]"
	}

	var listener net.Listener

	if addr.IsLoopback() {
		l, err := net.Listen("tcp", bindAddr+":9394")
		if err != nil {
			return err
		}
		listener = l
	} else {
		config, err := common.TlsConfig("server.crt", "server.key")
		if err != nil {
			return err
		}

		config.Rand = rand.Reader

		l, err := tls.Listen("tcp", bindAddr+":9395", config)
		if err != nil {
			return err
		}
		listener = l
	}
	defer listener.Close()

	rpcSrv := rpc.NewServer()
	rpcSrv.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	rpcSrv.RegisterService(rcvr, "")

	r := mux.NewRouter()
	r.Handle("/rpc/v1", rpcSrv)
	r.Use(Logging)

	httpSrv := http.Server{Handler: r}
	http2.ConfigureServer(&httpSrv, &http2.Server{})

	idleConnsClosed := make(chan struct{})
	go func() {
		<-ctx.Done()
		log.Printf("Closing %s ...\n", bindAddr)
		if err := httpSrv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("Listening %s\n", bindAddr)

	if err := httpSrv.Serve(listener); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server Serve: %v", err)
	}

	<-idleConnsClosed

	return nil
}

func serveUnixSocket(ctx context.Context, sockpath string, rcvr *RPC) error {
	if sockpath[0] == '@' && runtime.GOOS == "linux" {
		sockpath = sockpath + string(0)
	}

	listener, err := net.Listen("unix", sockpath)
	if err != nil {
		return err
	}
	defer listener.Close()

	rpcSrv := rpc.NewServer()
	rpcSrv.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	rpcSrv.RegisterService(rcvr, "")

	r := mux.NewRouter()
	r.Handle("/rpc/v1", rpcSrv)
	r.Use(Logging)

	httpSrv := http.Server{Handler: r}

	idleConnsClosed := make(chan struct{})
	go func() {
		<-ctx.Done()
		log.Printf("Closing %s ...\n", sockpath)
		if err := httpSrv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("Listening %s\n", sockpath)

	if err := httpSrv.Serve(listener); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server Serve: %v", err)
	}

	<-idleConnsClosed

	return nil
}

func main() {
	rcvr := RPC{}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		cSignal := make(chan os.Signal, 1)
		signal.Notify(cSignal, syscall.SIGINT, syscall.SIGTERM)
		log.Println(<-cSignal)
		cancel()
	}()

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error { return serve(ctx, net.IPv4(127, 0, 0, 1), &rcvr) })
	group.Go(func() error { return serve(ctx, net.IPv4(0, 0, 0, 0), &rcvr) })
	group.Go(func() error { return serveUnixSocket(ctx, "@/tmp/server.sock", &rcvr) })

	if err := group.Wait(); err != nil {
		log.Fatal(err)
	}
}
