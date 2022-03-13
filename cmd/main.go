package cmd

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

const DEFAULT_PORT uint = 44244
const DEFAULT_ADDRESS_IPV4 string = "0.0.0.0"

var Version = "-dev"

func Body() int {
	opts, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if opts.showVersion {
		fmt.Printf("v%s\n", Version)
		return 0
	}

	opts, err = validate(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n\n", err)
		flag.Usage()
		return 1
	}

	err = server(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	return 0
}

func parseArgs() (options, error) {
	opts := options{}

	flag.BoolVar(&opts.showVersion, "version", false, "show version")
	flag.Float64Var(&opts.longitude, "longitude", 0.0, "longitude (decimal)")
	flag.Float64Var(&opts.latitude, "latitude", 0.0, "latitude (decimal)")
	flag.StringVar(&opts.callsign, "callsign", "", "callsign")
	flag.StringVar(&opts.comment, "comment", "", "comment")

	var dial string
	flag.StringVar(&dial, "dial", "tcp://cwop.aprs.net:14580", "dial address (disabled if empty)")
	flag.IntVar(&opts.dialpass, "dialpass", -1, "dial pass")

	var rl string
	flag.StringVar(&rl, "requestlog", "", "requestlog file path (disabled if empty)")

	flag.StringVar(&opts.ssid, "ssid", "15", "ssid")
	flag.UintVar(&opts.port, "port", DEFAULT_PORT, "tcp port (automatic 0)")
	flag.StringVar(&opts.argAddress, "address", DEFAULT_ADDRESS_IPV4, "IP address")

	flag.BoolVar(&opts.calcRainLast24Hours, "calcrain", false, "calculate trailing 24hr rainfall using daily rain")
	flag.UintVar(&opts.calcRainLast24HoursThreshold, "calcrainmins", 15, "minutes beyond 24hr allowed for calcrain")

	flag.Parse()

	if len(dial) > 0 {
		opts.dial = &dial
	}

	if len(rl) > 0 {
		opts.requestlog = &rl
	}

	return opts, nil
}

func validate(opts options) (options, error) {
	// longitude and latitude validation
	if opts.longitude == 0.0 {
		return opts, fmt.Errorf("invalid longitude")
	}
	if opts.latitude == 0.0 {
		return opts, fmt.Errorf("invalid latitude")
	}

	// address validation
	opts.address = net.ParseIP(opts.argAddress)
	if opts.address == nil {
		return opts, fmt.Errorf("invalid address")
	}

	// port validation
	const max_port = 65535
	if opts.port > max_port {
		return opts, fmt.Errorf("max port (%d)", max_port)
	}

	// ssid validation
	if len(opts.ssid) > 2 {
		return opts, fmt.Errorf("ssid too long")
	} else if len(opts.ssid) < 1 {
		return opts, fmt.Errorf("ssid empty")
	}

	// callsign validation
	if len(opts.callsign) == 0 {
		return opts, fmt.Errorf("missing callsign")
	}

	// APRS source validation
	err := opts.aprsSource.FromString(fmt.Sprintf("%s-%s", opts.callsign, opts.ssid))
	if err != nil {
		return opts, err
	}

	opts.callsign = strings.Trim(opts.callsign, "-")
	opts.ssid = strings.Trim(opts.ssid, "-")

	return opts, nil
}
