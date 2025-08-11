package main

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"time"
	"tz1/internal/subscription"
	sdb "tz1/internal/subscription/db"
	"tz1/pkg/client/postgresql"
	"tz1/pkg/config"
	"tz1/pkg/logging"
)

func main() {
	logger := logging.GetLogger()
	logger.Info("create router")
	router := httprouter.New()

	cfg := config.GetConfig()

	postgreSQLClient, err := postgresql.NewClient(context.TODO(), 6, cfg.Storage)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	logger.Info("register subscription handler")
	sRep := sdb.NewRepository(postgreSQLClient, logger)
	sHandler := subscription.NewHandler(sRep, logger)
	sHandler.Register(router)

	start(router, cfg)
}

func start(router *httprouter.Router, cfg *config.Config) {
	logger := logging.GetLogger()
	logger.Info("start application")

	var listener net.Listener
	var listenErr error

	logger.Info("listen tcp")
	listener, listenErr = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIp, cfg.Listen.Port))
	logger.Infof("server is listening port %s:%s", cfg.Listen.BindIp, cfg.Listen.Port)

	if listenErr != nil {
		logger.Fatal(listenErr)
	}

	server := &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(server.Serve(listener))
}
