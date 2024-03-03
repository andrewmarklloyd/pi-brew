package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alexhowarth/go-tilt"
	"github.com/jaedle/golang-tplink-hs100/pkg/configuration"
	"github.com/jaedle/golang-tplink-hs100/pkg/hs100"
)

const (
	heaterOutlet               = "fermentor-heater"
	fridgeOutlet               = "fermentor-fridge"
	desiredTempDefault         = 68
	tempVarianceDegreesDefault = 2
)

var (
	desiredTemp         uint16
	tempVarianceDegrees uint16
)

func main() {
	desiredTempString, ok := os.LookupEnv("DESIRED_TEMP")
	if ok {
		ui64, err := strconv.ParseUint(desiredTempString, 10, 64)
		if err != nil {
			panic(err)
		}

		desiredTemp = uint16(ui64)
	} else {
		desiredTemp = desiredTempDefault
	}

	desiredTempVarianceString, ok := os.LookupEnv("TEMP_VARIANCE")
	if ok {
		ui64, err := strconv.ParseUint(desiredTempVarianceString, 10, 64)
		if err != nil {
			panic(err)
		}

		tempVarianceDegrees = uint16(ui64)
	} else {
		tempVarianceDegrees = tempVarianceDegreesDefault
	}

	fmt.Printf("configuration - desired temp: %d, temp variance: %d\n", desiredTemp, tempVarianceDegrees)

	fmt.Println("setting up outlets")
	outlets, err := setupOutlets()
	if err != nil {
		panic(err)
	}

	fmt.Println("getting primary tilt temp")
	temp, err := getPrimaryTiltTemp()
	if err != nil {
		panic(err)
	}

	fmt.Println("current temp:", temp)
	err = triggerOutlets(outlets, temp)
	if err != nil {
		panic(err)
	}
}

func triggerOutlets(outlets map[string]*hs100.Hs100, temp uint16) error {
	adjustedTempLow := temp - tempVarianceDegrees
	adjustedTempHigh := temp + tempVarianceDegrees

	if adjustedTempLow < desiredTemp {
		fmt.Println("temp is too low, turning on heat")
		err := outlets[heaterOutlet].TurnOn()
		if err != nil {
			return fmt.Errorf("turning heater on")
		}
		err = outlets[fridgeOutlet].TurnOff()
		if err != nil {
			return fmt.Errorf("turning fridge off")
		}
		return nil
	}

	if adjustedTempHigh > desiredTemp {
		fmt.Println("temp is too high, turning on cooling")
		err := outlets[fridgeOutlet].TurnOn()
		if err != nil {
			return fmt.Errorf("turning fridge on")
		}
		err = outlets[heaterOutlet].TurnOff()
		if err != nil {
			return fmt.Errorf("turning heater off")
		}
		return nil
	}

	fmt.Println("temp is in ok range")

	return nil
}

func getPrimaryTiltTemp() (uint16, error) {
	s := tilt.NewScanner()
	s.Scan(10 * time.Second)

	primaryTiltColor := "Green"
	if len(s.Tilts()) == 0 {
		return 0, fmt.Errorf("did not find any tilts")
	}

	for _, t := range s.Tilts() {
		if t.Colour() == tilt.Colour(primaryTiltColor) {
			return t.Fahrenheit(), nil
		}
	}

	return 0, fmt.Errorf("no tilts matched primary color %s", primaryTiltColor)
}

func setupOutlets() (map[string]*hs100.Hs100, error) {
	allDevices, err := hs100.Discover("192.168.1.1/24", configuration.Default().WithTimeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("error getting devices: %w", err)
	}

	outlets := make(map[string]*hs100.Hs100, 2)
	for _, d := range allDevices {
		name, err := d.GetName()
		if err != nil {
			return nil, fmt.Errorf("error getting device name: %w", err)
		}
		if name == heaterOutlet {
			outlets[heaterOutlet] = d
		}
		if name == fridgeOutlet {
			outlets[fridgeOutlet] = d
		}
	}

	if len(outlets) != 2 {
		return nil, fmt.Errorf("expected to find 2 outlets but found %d", len(outlets))
	}

	return outlets, nil
}
