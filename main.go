package main

import (
	"os"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/route"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func main() {
	core.Logger.Info("Starting server...")
	bsSvc := buzzscreen.New()
	server := core.NewServer(bsSvc)
	configPath := os.Getenv("GOPATH") + "/src/github.com/Buzzvil/buzzscreen-api/config/"
	tracer.Start()
	defer tracer.Stop()
	server.Init(configPath, &env.Config)
	route.InitLegacyRoute(server.Engine())
	server.Start()
}
