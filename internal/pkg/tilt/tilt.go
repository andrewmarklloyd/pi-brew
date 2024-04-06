package tilt

import (
	"fmt"
	"time"

	"github.com/alexhowarth/go-tilt"
)

const (
	PrimaryTiltColor = "Blue"
	scanTimeout      = 10 * time.Second
)

func GetStats() (uint16, float64, error) {
	s := tilt.NewScanner()
	s.Scan(scanTimeout)

	if len(s.Tilts()) == 0 {
		return 0, 0, fmt.Errorf("did not find any tilts")
	}

	var temp uint16
	var gravity float64

	for _, t := range s.Tilts() {
		if t.Colour() == tilt.Colour(PrimaryTiltColor) {
			temp = t.Fahrenheit()
			gravity = t.Gravity()
		}
	}

	if temp == 0 {
		return 0, 0, fmt.Errorf("temp value not set")
	}

	if gravity == 0 {
		return 0, 0, fmt.Errorf("gravity value not set")
	}

	return temp, gravity, nil
}
