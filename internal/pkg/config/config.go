package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	desiredTempDefault         = 68
	tempVarianceDegreesDefault = 2
)

type Config struct {
	DesiredTemp         uint16
	TempVarianceDegrees uint16
	DatadogApiKey       string
	DatadogAppKey       string
}

func GetConfig() (Config, error) {
	var desiredTemp, tempVarianceDegrees uint16

	desiredTempString, ok := os.LookupEnv("DESIRED_TEMP")
	if ok {
		ui64, err := strconv.ParseUint(desiredTempString, 10, 64)
		if err != nil {
			return Config{}, err
		}

		desiredTemp = uint16(ui64)
	} else {
		desiredTemp = desiredTempDefault
	}

	desiredTempVarianceString, ok := os.LookupEnv("TEMP_VARIANCE")
	if ok {
		ui64, err := strconv.ParseUint(desiredTempVarianceString, 10, 64)
		if err != nil {
			return Config{}, err
		}

		tempVarianceDegrees = uint16(ui64)
	} else {
		tempVarianceDegrees = tempVarianceDegreesDefault
	}

	ddApiKeyBytes, err := os.ReadFile("/home/pi/.dd-api-key")
	if err != nil {
		return Config{}, fmt.Errorf("reading dd api key file: %w", err)
	}

	ddAppKeyBytes, err := os.ReadFile("/home/pi/.dd-app-key")
	if err != nil {
		return Config{}, fmt.Errorf("reading dd app key file: %w", err)
	}

	return Config{
		DesiredTemp:         desiredTemp,
		TempVarianceDegrees: tempVarianceDegrees,
		DatadogApiKey:       string(ddApiKeyBytes),
		DatadogAppKey:       string(ddAppKeyBytes),
	}, nil
}
