package flight

import "fmt"

// WeatherConditions: condiciones meteorológicas para la decisión de vuelo.
type WeatherConditions struct {
	WindSpeedMS     float64 `json:"wind_speed_ms"`
	WindGustMS      float64 `json:"wind_gust_ms"`
	PrecipitationMM float64 `json:"precipitation_mm"`
	VisibilityM     float64 `json:"visibility_m"`
	TemperatureC    float64 `json:"temperature_c"`
	CloudBaseM      float64 `json:"cloud_base_m"`
}

// WeatherAssessment: el dictamen GO/NO-GO con su razón.
type WeatherAssessment struct {
	GoForFlight bool     `json:"go_for_flight"`
	Summary     string   `json:"summary"`
	Warnings    []string `json:"warnings"`
}

// AssessWeather decide si volar, como lo haría un piloto profesional conservador.
func AssessWeather(w WeatherConditions, env FlightEnvelope) WeatherAssessment {
	a := WeatherAssessment{GoForFlight: true}

	if w.WindSpeedMS > env.MaxWindSpeedMS {
		a.GoForFlight = false
		a.Warnings = append(a.Warnings, fmt.Sprintf("viento %.1f m/s > límite %.1f", w.WindSpeedMS, env.MaxWindSpeedMS))
	}
	if w.WindGustMS > env.MaxWindSpeedMS*1.3 {
		a.GoForFlight = false
		a.Warnings = append(a.Warnings, fmt.Sprintf("ráfagas %.1f m/s peligrosas", w.WindGustMS))
	}
	if w.PrecipitationMM > 0.2 {
		a.GoForFlight = false
		a.Warnings = append(a.Warnings, "precipitación: el equipo no es resistente al agua")
	}
	if w.VisibilityM < env.MinVisibilityM {
		a.GoForFlight = false
		a.Warnings = append(a.Warnings, fmt.Sprintf("visibilidad %.0fm insuficiente", w.VisibilityM))
	}
	// Las baterías LiPo pierden rendimiento en frío extremo y se degradan en calor
	if w.TemperatureC < 0 {
		a.GoForFlight = false
		a.Warnings = append(a.Warnings, "temperatura bajo cero: rendimiento de batería comprometido")
	}
	if w.TemperatureC > 40 {
		a.Warnings = append(a.Warnings, "calor extremo: vigilar temperatura de batería")
	}

	if a.GoForFlight {
		a.Summary = fmt.Sprintf("GO — viento %.1f m/s, visibilidad %.0fm, %.0f°C", w.WindSpeedMS, w.VisibilityM, w.TemperatureC)
	} else {
		a.Summary = "NO-GO — condiciones fuera de límites operacionales"
	}
	return a
}
