package tilt

import (
	"fmt"
	"time"

	"github.com/alexhowarth/go-tilt"
)

const (
	PrimaryTiltColor = "Green"
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
		if t.Colour() == tilt.Colour(PrimaryTiltColor) {
			return Client{
				scanner: *s,
				tilt:    t,
			}, nil
		}
	}

	return Client{}, fmt.Errorf("no tilts matched primary color %s", PrimaryTiltColor)
}

func (t *Client) GetPrimaryTiltTemp() uint16 {
	return t.tilt.Fahrenheit()
}

func (t *Client) GetPrimaryTiltGravity() float64 {
	return t.tilt.Gravity()
}
