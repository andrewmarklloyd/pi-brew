package main

import (
	"context"
	"time"

	"github.com/andrewmarklloyd/pi-brew/internal/pkg/config"
	"github.com/andrewmarklloyd/pi-brew/internal/pkg/datadog"
	"github.com/andrewmarklloyd/pi-brew/internal/pkg/outlet"
	"github.com/andrewmarklloyd/pi-brew/internal/pkg/tilt"
	"go.uber.org/zap"
)

const tickerDuration = 10 * time.Minute

var (
	datadogClient datadog.Client
	tiltClient    tilt.Client
	outletClient  outlet.Client
	logger        *zap.SugaredLogger
)

func main() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger = l.Sugar().Named("pi-brew")
	defer logger.Sync()

	conf, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("getting config: %s", err.Error())
	}

	logger.Infof("using configuration: (temp: %d, variance: %d)", conf.DesiredTemp, conf.TempVarianceDegrees)

	datadogClient = datadog.NewDatadogClient(conf.DatadogApiKey, conf.DatadogAppKey)
	err = datadogClient.PublishMetric(context.Background(), "pi_brew.start", nil)
	if err != nil {
		logger.Fatalf("publishing start metric: %s", err.Error())
	}

	tiltClient, err = tilt.NewTiltClient()
	if err != nil {
		logger.Fatalf("creating tilt client: %s", err.Error())
	}

	outletClient, err = outlet.SetupOutlets(conf.DesiredTemp, conf.TempVarianceDegrees, datadogClient)
	if err != nil {
		logger.Fatalf("setting up outlets: %s", err.Error())
	}

	run(conf)

	ticker := time.NewTicker(tickerDuration)
	for range ticker.C {
		run(conf)
	}

}

func run(conf config.Config) {
	err := datadogClient.PublishMetric(context.Background(), "pi_brew.run_start", nil)
	if err != nil {
		logger.Fatalf("publishing run start metric: %s", err.Error())
	}

	temp := tiltClient.GetPrimaryTiltTemp()
	gravity := tiltClient.GetPrimaryTiltGravity()
	logger.Infof("gravity measure: %f", gravity)

	err = datadogClient.PublishMetricWithValue(context.Background(), "pi_brew.temp", float64(temp), map[string]string{
		"color": tilt.PrimaryTiltColor,
	})
	if err != nil {
		logger.Fatalf("publishing temp metric: %s", err.Error())
	}

	err = datadogClient.PublishMetricWithValue(context.Background(), "pi_brew.gravity", float64(gravity), map[string]string{
		"color": tilt.PrimaryTiltColor,
	})
	if err != nil {
		logger.Fatalf("publishing gravity metric: %s", err.Error())
	}

	err = outletClient.TriggerOutlets(temp, logger)
	if err != nil {
		logger.Fatalf("triggering outlets: %s", err.Error())
	}

	err = datadogClient.PublishMetric(context.Background(), "pi_brew.run_end", nil)
	if err != nil {
		logger.Fatalf("publishing run end metric: ", err.Error())
	}
}
