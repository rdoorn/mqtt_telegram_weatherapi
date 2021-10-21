package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/rdoorn/gohelper/mqtthelper"
	"github.com/rdoorn/gohelper/statsdhelper"
)

const mqttClientID = "mqtt_telegram_weatherapi"

type Handler struct {
	mqtt   *mqtthelper.Handler
	statsd *statsdhelper.Handler
	last   time.Time
}

type TelemetryMQTTStatus struct {
	Time    *int64  `json:"time"`
	TimeStr *string `json:"time_string"`
	Summary *string `json:"summary"`
	// Icon:clear-day
	SunriseTime  int64   `json:"sunrise_time"`
	SunsetTime   int64   `json:"sunset_time"`
	SunriseTimeH float64 `json:"sunrise_time_h"`
	SunsetTimeH  float64 `json:"sunset_time_h"`
	//SunsetTime:0
	PrecipIntensity *float64 `json:"rain_intensity"`
	//PrecipIntensityMax:0
	//PrecipIntensityMaxTime:0
	//PrecipProbability *float64 `json:"rain_posibility"`
	//PrecipType:  rain|
	//PrecipAccumulation:0
	Temperature *float64 `json:"temperature"`
	//TemperatureMin:0
	//TemperatureMinTime:0
	//TemperatureMax:0
	//TemperatureMaxTime:0
	ApparentTemperature *float64 `json:"apparent_temperature"`
	//ApparentTemperatureMin:0
	//ApparentTemperatureMinTime:0
	//ApparentTemperatureMax:0
	//ApparentTemperatureMaxTime:0
	//NearestStormBearing  *float64 `json:"nearest_storm_bearing"`
	//NearestStormDistance *float64 `json:"nearest_storm_distance"`

	//DewPoint         *float64 `json:"dew_point"`
	WindSpeed        *float64 `json:"wind_speed"`
	WindGust         *float64 `json:"wind_gust"`
	WindBearing      *int64   `json:"wind_bearing"`
	CloudCover       *int64   `json:"cloud_cover"`
	Humidity         *int64   `json:"humidity"`
	Pressure         *float64 `json:"pressure"`
	Visibility       *float64 `json:"visibility"`
	Ozone            *float64 `json:"ozone"`
	CarbonOxide      *float64 `json:"carbon_oxide"`
	NitrogenOxide    *float64 `json:"nitrogen_oxide"`
	SulphurDioxide   *float64 `json:"sulphur_dioxide"`
	PM2_5            *float64 `json:"pm2_5"`
	PM10             *float64 `json:"pm10"`
	MoonPhase        *string  `json:"moon_phase"`
	MoonIllumination *int     `json:"moon_illumination"`
	UVIndex          *float64 `json:"uv_index"`
	//UVIndexTime float64
}

func (h *Handler) mqttOut(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))

	switch msg.Topic() {
	case "weatherapi/out":

		update := &TelemetryMQTTStatus{}
		err := json.Unmarshal(msg.Payload(), update)
		if err != nil {
			log.Printf("error marshaling request: %s", err)
			return
		}
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.rain_intensity", update.TimeStr), fmt.Sprintf("%f", update.PrecipIntensity))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.temperature", update.TimeStr), fmt.Sprintf("%f", update.Temperature))
		log.Printf("sending value heerhugowaard2.%s.temperature=%f", update.TimeStr, update.Temperature)
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.apparent_temperature", update.TimeStr), fmt.Sprintf("%f", update.ApparentTemperature))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.wind_speed", update.TimeStr), fmt.Sprintf("%f", update.WindSpeed))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.wind_gust", update.TimeStr), fmt.Sprintf("%f", update.WindGust))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.wind_bearing", update.TimeStr), fmt.Sprintf("%f", update.WindBearing))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.cloud_cover", update.TimeStr), fmt.Sprintf("%f", update.CloudCover))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.humidity", update.TimeStr), fmt.Sprintf("%f", update.Humidity))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.pressure", update.TimeStr), fmt.Sprintf("%f", update.Pressure))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.visibility", update.TimeStr), fmt.Sprintf("%f", update.Visibility))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.ozone", update.TimeStr), fmt.Sprintf("%f", update.Ozone))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.moon_illimination", update.TimeStr), fmt.Sprintf("%f", update.MoonIllumination))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.uv_index", update.TimeStr), fmt.Sprintf("%d", update.UVIndex))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.carbon_oxide", update.TimeStr), fmt.Sprintf("%f", update.CarbonOxide))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.nitrogen_oxide", update.TimeStr), fmt.Sprintf("%f", update.NitrogenOxide))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.sulphur_dioxide", update.TimeStr), fmt.Sprintf("%f", update.SulphurDioxide))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.pm2_5", update.TimeStr), fmt.Sprintf("%f", update.CarbonOxide))
		h.statsd.Gauge(1.0, fmt.Sprintf("heerhugowaard2.%s.pm10", update.TimeStr), fmt.Sprintf("%f", update.CarbonOxide))
	}

	h.last = time.Now()

}

func main() {

	h := Handler{
		mqtt:   mqtthelper.New(),
		statsd: statsdhelper.New(),
	}

	// Setup MQTT Sub
	err := h.mqtt.Subscribe(mqttClientID, "weatherapi/out", 0, h.mqttOut)
	if err != nil {
		panic(err)
	}

	// loop till exit
	sigterm := make(chan os.Signal, 10)
	signal.Notify(sigterm, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigterm:
			log.Printf("Program killed by signal!")
			return
		}
	}
}
