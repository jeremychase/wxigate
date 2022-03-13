package cmd

import (
	"net"

	"github.com/ebarkie/aprs"
)

type options struct {
	aprsSource                   aprs.Addr
	callsign                     string
	comment                      string
	dial                         *string
	ssid                         string
	argAddress                   string
	requestlog                   *string
	address                      net.IP
	dialpass                     int
	longitude                    float64
	latitude                     float64
	port                         uint
	showVersion                  bool
	calcRainLast24Hours          bool
	calcRainLast24HoursThreshold uint
}
