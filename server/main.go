package main

import (
	"fmt"
	"net"
	"os"
	"wa/db"
	"wa/def"
	"wa/rpc"
	"wa/rpc/pb"

	"aconfig"

	"github.com/fatih/color"
	"google.golang.org/grpc"

	"net/http"
	_ "net/http/pprof"
)

var BuildTime string

type Config struct {
	LogLevel int
	Pprof    bool
	Port     int
}

var cfg_fn = "server.toml"

func main() {
	fmt.Println("Build Time: " + color.HiBlueString(BuildTime))

	color.HiBlue(`loading config %s`, cfg_fn)
	cfg := &Config{
		LogLevel: -1,
		Pprof:    false,
		Port:     3423,
	}
	aconfig.Load(cfg_fn, cfg)
	aconfig.Save(cfg_fn, cfg)

	db.LogLevel = cfg.LogLevel
	color.HiBlue(`set LogLevel to %d`, db.LogLevel)

	if cfg.Pprof {
		go func() {
			color.HiYellow("Pprof : http://localhost:7788/debug/pprof")
			e := http.ListenAndServe("0.0.0.0:7788", nil)
			if e != nil {
				panic(e)
			}
		}()
	}

	def.RpcPort = cfg.Port
	listen()
}

func listen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", def.RpcPort))
	if err != nil {
		color.HiRed("failed to listen: %v", err)
		os.Exit(1)
	}
	defer lis.Close()
	sv := &rpc.WaServer{}
	s := grpc.NewServer()
	pb.RegisterRpcServer(s, sv)

	color.White("Server version: %s + %s",
		def.VERSION(false), def.VERSION((true)))
	color.HiGreen("Server running at %d", def.RpcPort)

	if err := s.Serve(lis); err != nil {
		color.HiRed("failed to serve: %v", err)
	}
}
