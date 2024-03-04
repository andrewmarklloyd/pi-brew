package tilt

import (
	"fmt"
	"time"

	"github.com/alexhowarth/go-tilt"
)

const (
	primaryTiltColor = "Green"
	scanTimeout      = 10 * time.Second
)

func GetPrimaryTiltTemp() (uint16, error) {
	s := tilt.NewScanner()
	s.Scan(scanTimeout)

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
