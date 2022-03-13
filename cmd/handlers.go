package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"wxigate/calcrain"

	"github.com/ebarkie/aprs"
)

func awpHandlerV1(opts options) http.Handler {
	wx := aprs.Wx{
		Lat:  float64(opts.latitude),
		Lon:  float64(opts.longitude),
		Type: opts.comment,
	}

	// SwName are SwVers are concatenated in the 'comment' field and then
	// immediately followed by the Wx.Type. This is performed in the upstream
	// aprs library and looks like:
	//
	//   fmt.Sprintf("%s%s%s", aprs.SwName, aprs.SwVers, wx.Type)
	//
	// On aprs.fi a lowercase 'v%d' gets dropped, so that is why this is 'V'.
	aprs.SwName = "wxigate-V"
	aprs.SwVers = Version

	f := aprs.Frame{
		Dst:  aprs.Addr{Call: "APRS"},
		Src:  opts.aprsSource,
		Path: aprs.Path{aprs.Addr{Call: "TCPIP", Repeated: true}},
	}

	raindata := calcrain.Data{}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query()

		// Zero to avoid sending inaccurate data.
		wx.Zero()

		for k, v := range query {
			switch k {
			case "dateutc":
				layout := "2006-01-02 15:04:05"
				dateutc, err := time.Parse(layout, v[0])
				if err != nil {
					msg := fmt.Sprintf("'dateutc' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.Timestamp = dateutc
			case "baromrelin":
				baromrelin, err := strconv.ParseFloat(v[0], 64)
				if err != nil {
					msg := fmt.Sprintf("'baromrelin' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.Altimeter = baromrelin
			case "tempf":
				temp, err := strconv.ParseFloat(v[0], 64)
				if err != nil {
					msg := fmt.Sprintf("'tempf' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.Temp = int(temp)
			case "humidity":
				humidity, err := strconv.Atoi(v[0])
				if err != nil {
					msg := fmt.Sprintf("'humidity' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.Humidity = humidity
			case "hourlyrainin":
				hourlyrainin, err := strconv.ParseFloat(v[0], 64)
				if err != nil {
					msg := fmt.Sprintf("'hourlyrainin' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.RainLastHour = hourlyrainin
			case "24hourrainin":
				hourlyrainin24, err := strconv.ParseFloat(v[0], 64)
				if err != nil {
					msg := fmt.Sprintf("'24hourrainin' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.RainLast24Hours = hourlyrainin24
			case "dailyrainin":
				dailyrainin, err := strconv.ParseFloat(v[0], 64)
				if err != nil {
					msg := fmt.Sprintf("'dailyrainin' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.RainToday = dailyrainin
			case "solarradiation":
				solarradiation, err := strconv.ParseFloat(v[0], 64)
				if err != nil {
					msg := fmt.Sprintf("'solarradiation' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.SolarRad = int(solarradiation)
			case "winddir":
				winddir, err := strconv.Atoi(v[0])
				if err != nil {
					msg := fmt.Sprintf("'winddir' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.WindDir = winddir
			case "windgustmph":
				windgustmph, err := strconv.ParseFloat(v[0], 64)
				if err != nil {
					msg := fmt.Sprintf("'windgustmph' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.WindGust = int(windgustmph)
			case "windspeedmph":
				windspeedmph, err := strconv.ParseFloat(v[0], 64)
				if err != nil {
					msg := fmt.Sprintf("'windspeedmph' error: %v", err)
					errorHandler(w, req, msg, http.StatusServiceUnavailable)
					return
				}
				wx.WindSpeed = int(windspeedmph)
			}
		}

		if opts.calcRainLast24Hours {
			raindata.Append(wx.RainToday, wx.Timestamp)

			rl, err := raindata.RainLast24Hours(wx.RainToday, wx.Timestamp, opts.calcRainLast24HoursThreshold)

			if err == nil {
				wx.RainLast24Hours = rl
			}
		}

		if opts.dial != nil {
			fmt.Printf("Sending, temp(%d): %v\n", wx.Temp, wx.String())

			f.Text = wx.String()

			err := f.SendIS(*opts.dial, opts.dialpass)
			if err != nil {
				msg := fmt.Sprintf("Upload error: %s", err)
				errorHandler(w, req, msg, http.StatusServiceUnavailable)
				return
			}
		} else {
			fmt.Printf("Received but not sending, temp(%d): %v\n", wx.Temp, wx.String())
		}

		w.WriteHeader(http.StatusOK)
	})
}

func errorHandler(w http.ResponseWriter, req *http.Request, msg string, status int) {

	// server output
	stream := os.Stderr
	if status == http.StatusNotFound {
		stream = os.Stdout
	}
	fmt.Fprintf(
		stream,
		"%v -- \"%v %v %v\" %d %s\n",
		req.RemoteAddr,
		req.Method,
		req.URL.Path,
		req.Proto,
		status,
		msg,
	)

	// response
	w.WriteHeader(status)
	fmt.Fprintf(w, "%d - %s", status, msg)
}

func defaultHandler(w http.ResponseWriter, req *http.Request) {
	errorHandler(w, req, "not found", http.StatusNotFound)
}
