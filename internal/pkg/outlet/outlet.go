package outlet

import (
	"context"
	"fmt"
	"time"

	"github.com/andrewmarklloyd/pi-brew/internal/pkg/datadog"
	"github.com/jaedle/golang-tplink-hs100/pkg/configuration"
	"github.com/jaedle/golang-tplink-hs100/pkg/hs100"
)

const (
	heaterOutlet = "fermentor-heater"
	fridgeOutlet = "fermentor-fridge"
)

type Client struct {
	desiredTemp, tempVarianceDegrees uint16
	outlets                          map[string]*hs100.Hs100
	datadogClient                    datadog.Client
}

func SetupOutlets(desiredTemp, tempVarianceDegrees uint16, datadogClient datadog.Client) (Client, error) {
	allDevices, err := hs100.Discover("192.168.1.1/24", configuration.Default().WithTimeout(5*time.Second))
	if err != nil {
		return Client{}, fmt.Errorf("error getting devices: %w", err)
	}

	outlets := make(map[string]*hs100.Hs100, 2)
	for _, d := range allDevices {
		name, err := d.GetName()
		if err != nil {
			return Client{}, fmt.Errorf("error getting device name: %w", err)
		}
		if name == heaterOutlet {
			outlets[heaterOutlet] = d
		}
		if name == fridgeOutlet {
			outlets[fridgeOutlet] = d
		}
	}

	if len(outlets) != 2 {
		return Client{}, fmt.Errorf("expected to find 2 outlets but found %d", len(outlets))
	}

	return Client{
		desiredTemp:         desiredTemp,
		tempVarianceDegrees: tempVarianceDegrees,
		outlets:             outlets,
		datadogClient:       datadogClient,
	}, nil
}

func (o *Client) TriggerOutlets(temp uint16) error {
	tempLow := o.desiredTemp - o.tempVarianceDegrees
	tempHigh := o.desiredTemp + o.tempVarianceDegrees

	if temp < tempLow {
		fmt.Printf("current temp %d is lower than %d, turning on heat\n", temp, tempLow)
		err := o.outlets[heaterOutlet].TurnOn()
		if err != nil {
			return fmt.Errorf("turning heater on")
		}
		err = o.outlets[fridgeOutlet].TurnOff()
		if err != nil {
			return fmt.Errorf("turning fridge off")
		}
		if err := o.publishMetrics("on", "off"); err != nil {
			return err
		}
		return nil
	}

	if temp > tempHigh {
		fmt.Printf("current temp %d is higher than %d, turning on cooling\n", temp, tempHigh)
		err := o.outlets[fridgeOutlet].TurnOn()
		if err != nil {
			return fmt.Errorf("turning fridge on")
		}
		err = o.outlets[heaterOutlet].TurnOff()
		if err != nil {
			return fmt.Errorf("turning heater off")
		}
		if err := o.publishMetrics("off", "on"); err != nil {
			return err
		}
		return nil
	}

	fmt.Println("temp is in ok range, ensuring fridge and heater are off")
	err := o.outlets[fridgeOutlet].TurnOff()
	if err != nil {
		return fmt.Errorf("turning fridge off")
	}
	err = o.outlets[heaterOutlet].TurnOff()
	if err != nil {
		return fmt.Errorf("turning heater off")
	}
	if err := o.publishMetrics("off", "off"); err != nil {
		return err
	}

	return nil
}

func (o *Client) publishMetrics(heaterMode, fridgeMode string) error {
	metricName := "outlet_trigger"
	err := o.datadogClient.PublishMetric(context.Background(), metricName, map[string]string{
		"outlet": heaterOutlet,
		"mode":   heaterMode,
	})
	if err != nil {
		return fmt.Errorf("publishing metric %s with values (outlet: %s, mode: %s) %w", metricName, heaterOutlet, heaterMode, err)
	}

	err = o.datadogClient.PublishMetric(context.Background(), metricName, map[string]string{
		"outlet": fridgeOutlet,
		"mode":   fridgeMode,
	})
	if err != nil {
		return fmt.Errorf("publishing metric %s with values (outlet: %s, mode: %s) %w", metricName, fridgeOutlet, fridgeMode, err)
	}

	return nil
}
