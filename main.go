package main

import (
	"context"
	"fmt"
	"time"

	"github.com/andrewmarklloyd/pi-brew/internal/pkg/config"
	"github.com/andrewmarklloyd/pi-brew/internal/pkg/datadog"
	"github.com/andrewmarklloyd/pi-brew/internal/pkg/outlet"
	"github.com/andrewmarklloyd/pi-brew/internal/pkg/tilt"
)

const tickerDuration = 5 * time.Minute

var datadogClient datadog.Client
var tiltClient tilt.Client
var outletClient outlet.Client

func main() {
	conf, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	datadogClient = datadog.NewDatadogClient(conf.DatadogApiKey, conf.DatadogAppKey)
	err = datadogClient.PublishMetric(context.Background(), "pi_brew.start")
	if err != nil {
		panic(err)
	}

	tiltClient, err = tilt.NewTiltClient()
	if err != nil {
		panic(err)
	}

	outletClient, err = outlet.SetupOutlets(conf.DesiredTemp, conf.TempVarianceDegrees)
	if err != nil {
		panic(err)
	}

	fmt.Printf("configuration - desired temp: %d, temp variance: %d\n", conf.DesiredTemp, conf.TempVarianceDegrees)

	ticker := time.NewTicker(tickerDuration)
	for range ticker.C {
		run(conf)
	}

}

func run(conf config.Config) {
	err := datadogClient.PublishMetric(context.Background(), "pi_brew.run_start")
	if err != nil {
		panic(err)
	}

	fmt.Println("getting primary tilt temp")
	temp, err := tiltClient.GetPrimaryTiltTemp()
	if err != nil {
		panic(err)
	}

	err = outletClient.TriggerOutlets(temp)
	if err != nil {
		panic(err)
	}

	err = datadogClient.PublishMetric(context.Background(), "pi_brew.run_end")
	if err != nil {
		panic(err)
	}
}
