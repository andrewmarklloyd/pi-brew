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

func main() {
	conf, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	datadogClient := datadog.NewDatadogClient(conf.DatadogApiKey, conf.DatadogAppKey)
	err = datadogClient.PublishMetric(context.Background(), "pi-brew.start")
	if err != nil {
		panic(err)
	}

	fmt.Printf("configuration - desired temp: %d, temp variance: %d\n", conf.DesiredTemp, conf.TempVarianceDegrees)

	fmt.Println("setting up outlets")
	outletClient, err := outlet.SetupOutlets(conf.DesiredTemp, conf.TempVarianceDegrees)
	if err != nil {
		panic(err)
	}

	fmt.Println("getting primary tilt temp")
	temp, err := tilt.GetPrimaryTiltTemp()
	if err != nil {
		panic(err)
	}

	fmt.Println("current temp:", temp)
	err = outletClient.TriggerOutlets(temp)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Hour)
}
