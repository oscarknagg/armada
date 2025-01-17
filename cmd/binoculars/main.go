package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/G-Research/armada/internal/binoculars"
	"github.com/G-Research/armada/internal/binoculars/configuration"
	"github.com/G-Research/armada/internal/common"
	"github.com/G-Research/armada/internal/common/grpc"
	api "github.com/G-Research/armada/pkg/api/binoculars"
)

const CustomConfigLocation string = "config"

func init() {
	pflag.String(CustomConfigLocation, "", "Fully qualified path to application configuration file")
	pflag.Parse()
}

func main() {
	common.ConfigureLogging()
	common.BindCommandlineArguments()

	var config configuration.BinocularsConfig
	userSpecifiedConfig := viper.GetString(CustomConfigLocation)
	common.LoadConfig(&config, "./config/binoculars", userSpecifiedConfig)

	log.Info("Starting...")

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	shutdownMetricServer := common.ServeMetrics(config.MetricsPort)
	defer shutdownMetricServer()

	shutdownGateway := grpc.ServeGateway(
		config.HttpPort,
		config.GrpcPort,
		config.CorsAllowedOrigins,
		api.SwaggerJsonTemplate(),
		api.RegisterBinocularsHandler,
	)
	defer shutdownGateway()

	shutdown, wg := binoculars.StartUp(&config)
	go func() {
		<-stopSignal
		shutdown()
	}()
	wg.Wait()
}
