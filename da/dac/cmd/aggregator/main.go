package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/urfave/cli"

	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/log"
)

var (
	Version   = "v0.0.1"
	GitCommit = ""
	GitDate   = ""
)

var flags = []cli.Flag{
  cli.IntFlag{
    Name: "port",
    Usage: "port of the HTTP server",
    EnvVar: "PORT",
    Value: 3000,
  },
}

func main() {
	oplog.SetupDefaults()

	app := cli.NewApp()
	app.Flags = flags
	app.Version = fmt.Sprintf("%s-%s-%s", Version, GitCommit, GitDate)
	app.Name = "dac-aggregator"
	app.Usage = "DAC Aggregator"
	app.Description = "Service that aggregates a list of DAC members"
	app.Action = Member
	app.Commands = []cli.Command{
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}

func Member(ctx *cli.Context) error {
  port := ctx.Int("port")

  aggregator := newAggregator()

  r := mux.NewRouter()
  r.HandleFunc("/batch", aggregator.handlePost).Methods("POST")
  r.HandleFunc("/batch/{dataHash}", aggregator.handleGet).Methods("GET")

  srv := &http.Server{
    Addr:    fmt.Sprintf(":%v", port),
    Handler: r,
  }

  log.Info("HTTP server start", "port", port)
  return srv.ListenAndServe()
}
