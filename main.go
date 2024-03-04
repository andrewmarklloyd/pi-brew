package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/andrewmarklloyd/pi-brew/internal/pkg/outlet"
	"github.com/andrewmarklloyd/pi-brew/internal/pkg/tilt"
)

const (
	desiredTempDefault         = 68
	tempVarianceDegreesDefault = 2
)

var (
	desiredTemp         uint16
	tempVarianceDegrees uint16
)

func main() {
	desiredTemp, tempVarianceDegrees, err := getConfig()
	if err != nil {
		panic(fmt.Sprintf("getting config: %s", err.Error()))
	}

	fmt.Printf("configuration - desired temp: %d, temp variance: %d\n", desiredTemp, tempVarianceDegrees)

	fmt.Println("setting up outlets")
	outletClient, err := outlet.SetupOutlets(desiredTemp, tempVarianceDegrees)
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
}

func getConfig() (uint16, uint16, error) {
	desiredTempString, ok := os.LookupEnv("DESIRED_TEMP")
	if ok {
		ui64, err := strconv.ParseUint(desiredTempString, 10, 64)
		if err != nil {
			return 0, 0, err
		}

		desiredTemp = uint16(ui64)
	} else {
		desiredTemp = desiredTempDefault
	}

	desiredTempVarianceString, ok := os.LookupEnv("TEMP_VARIANCE")
	if ok {
		ui64, err := strconv.ParseUint(desiredTempVarianceString, 10, 64)
		if err != nil {
			return 0, 0, err
		}

		tempVarianceDegrees = uint16(ui64)
	} else {
		tempVarianceDegrees = tempVarianceDegreesDefault
	}

	return desiredTemp, tempVarianceDegrees, nil
}
