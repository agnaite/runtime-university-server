package main

import (
	"fmt"
	"log"
	"net"

	"github.com/heroku/dogwood/pkg/proxyproto"
	"github.com/heroku/runtime-university-server/server"
	"github.com/heroku/runtime-university-server/spec"
	"github.com/joeshaw/envdecode"
	"google.golang.org/grpc"
)

type config struct {
	HTTPPort        int `env:"HEROKU_ROUTER_HTTP_PORT,required"`
	HealthCheckPort int `env:"HEROKU_ROUTER_HEALTHCHECK_PORT,required"`
}

func runGRPCServer(name string, port int) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	proxyprotoLn := &proxyproto.Listener{Listener: ln}
	defer proxyprotoLn.Close()

	srv := grpc.NewServer()
	spec.RegisterRouteGuideServer(srv, server.NewRouteGuideServer())
	log.Printf("at=binding service=%s port=%d", name, port)
	err = srv.Serve(proxyprotoLn)
	if err != nil {
		log.Fatal(err)
	}
}

func runHealthCheckServer(name string, port int) {
	log.Printf("at=binding service=%s port=%d", name, port)
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		log.Printf("at=handle-request service=%s", name)
		if err != nil {
			log.Fatal(err)
		}
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	cfg := &config{}
	envdecode.MustDecode(cfg)
	go runGRPCServer("grpc-server", cfg.HTTPPort)
	runHealthCheckServer("health-check", cfg.HealthCheckPort)
}
