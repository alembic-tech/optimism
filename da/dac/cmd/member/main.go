package main

import (
	"encoding/hex"
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
  cli.StringFlag{
    Name: "private-key",
    Usage: "32 bytes BLS private key",
    EnvVar: "PRIVATE_KEY",
  },
  cli.StringFlag{
    Name: "directory",
    Usage: "path to directory where batches will be stored",
    EnvVar: "DIRECTORY",
  },
}

func main() {
	oplog.SetupDefaults()

	app := cli.NewApp()
	app.Flags = flags
	app.Version = fmt.Sprintf("%s-%s-%s", Version, GitCommit, GitDate)
	app.Name = "dac-member"
	app.Usage = "DAC Member"
	app.Description = "Service for storing batches of data and sign a proof of storage"
	app.Action = Member
	app.Commands = []cli.Command{}

	err := app.Run(os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}

func Member(ctx *cli.Context) error {
  port := ctx.Int("port")

  storage := newFileStorage(ctx.String("directory"))

  member, err := newMember(storage, ctx.String("private-key"))
  if err != nil {
    return err
  }

  r := mux.NewRouter()
  r.HandleFunc("/batch", member.handlePost).Methods("POST")
  r.HandleFunc("/batch/{dataHash}", member.handleGet).Methods("GET")

  srv := &http.Server{
    Addr:    fmt.Sprintf(":%v", port),
    Handler: r,
  }

  publicKey := member.signer.GetPublicKey()
  log.Info("HTTP server start", "port", port, "public_key", hex.EncodeToString(publicKey.ToBytes()))
  return srv.ListenAndServe()
}
