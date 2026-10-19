package flight

import "testing"

func TestAssessWeather_GoNoGo(t *testing.T) {
	env := DefaultEnvelope() // MaxWindSpeedMS = 10, MinVisibilityM = 500
	tests := []struct {
		name string
		w    WeatherConditions
		want bool
	}{
		{"condiciones ideales", WeatherConditions{WindSpeedMS: 4, WindGustMS: 6, PrecipitationMM: 0, VisibilityM: 5000, TemperatureC: 22}, true},
		{"viento excesivo", WeatherConditions{WindSpeedMS: 12, VisibilityM: 5000, TemperatureC: 22}, false},
		{"ráfagas peligrosas", WeatherConditions{WindSpeedMS: 8, WindGustMS: 14, VisibilityM: 5000, TemperatureC: 22}, false},
		{"lluvia", WeatherConditions{WindSpeedMS: 4, PrecipitationMM: 2, VisibilityM: 5000, TemperatureC: 22}, false},
		{"visibilidad baja", WeatherConditions{WindSpeedMS: 4, VisibilityM: 200, TemperatureC: 22}, false},
		{"bajo cero", WeatherConditions{WindSpeedMS: 4, VisibilityM: 5000, TemperatureC: -2}, false},
		{"calor extremo pero vuela", WeatherConditions{WindSpeedMS: 4, VisibilityM: 5000, TemperatureC: 42}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AssessWeather(tt.w, env)
			if got.GoForFlight != tt.want {
				t.Errorf("GoForFlight = %v, quería %v (%s)", got.GoForFlight, tt.want, got.Summary)
			}
		})
	}
}
