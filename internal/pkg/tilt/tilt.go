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

type Client struct {
	scanner tilt.Scanner
	tilt    tilt.Tilt
}

func NewTiltClient() (Client, error) {
	s := tilt.NewScanner()
	s.Scan(scanTimeout)

	if len(s.Tilts()) == 0 {
		return Client{}, fmt.Errorf("did not find any tilts")
	}

	for _, t := range s.Tilts() {
		if t.Colour() == tilt.Colour(primaryTiltColor) {
			return Client{
				scanner: *s,
				tilt:    t,
			}, nil
		}
	}

	return Client{}, fmt.Errorf("no tilts matched primary color %s", primaryTiltColor)
}

func (t *Client) GetPrimaryTiltTemp() (uint16, error) {
	return t.tilt.Fahrenheit(), nil
}
