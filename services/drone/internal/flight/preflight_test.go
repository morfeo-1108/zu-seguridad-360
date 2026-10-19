package flight

import "testing"

func goodState() DroneState {
	return DroneState{
		BatteryPct: 87, BatteryHealthPct: 94, GPSSatellites: 16,
		CompassCalibrated: true, IMUCalibrated: true, FirmwareOK: true,
		GimbalOK: true, SDCardFreeMB: 8000, HomePointSet: true,
		SignalStrength: 92, MotorsOK: true,
	}
}

func TestRunPreflight_ClearsWhenAllGood(t *testing.T) {
	env := DefaultEnvelope()
	weather := WeatherAssessment{GoForFlight: true, Summary: "GO"}
	res := RunPreflight(goodState(), env, weather, 300)
	if !res.Cleared {
		t.Errorf("estado óptimo debería autorizar vuelo, razón: %s", res.Reason)
	}
}

func TestRunPreflight_BlocksOnCriticalFailures(t *testing.T) {
	env := DefaultEnvelope()
	goodWeather := WeatherAssessment{GoForFlight: true, Summary: "GO"}

	tests := []struct {
		name   string
		mutate func(*DroneState)
	}{
		{"batería baja", func(s *DroneState) { s.BatteryPct = 30 }},
		{"sin home point", func(s *DroneState) { s.HomePointSet = false }},
		{"GPS débil", func(s *DroneState) { s.GPSSatellites = 6 }},
		{"brújula descalibrada", func(s *DroneState) { s.CompassCalibrated = false }},
		{"firmware desactualizado", func(s *DroneState) { s.FirmwareOK = false }},
		{"señal débil", func(s *DroneState) { s.SignalStrength = 40 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := goodState()
			tt.mutate(&s)
			res := RunPreflight(s, env, goodWeather, 300)
			if res.Cleared {
				t.Errorf("'%s' debería BLOQUEAR el vuelo, pero autorizó", tt.name)
			}
		})
	}
}

func TestRunPreflight_BlocksOnBadWeather(t *testing.T) {
	env := DefaultEnvelope()
	badWeather := WeatherAssessment{GoForFlight: false, Summary: "NO-GO viento"}
	res := RunPreflight(goodState(), env, badWeather, 300)
	if res.Cleared {
		t.Error("clima NO-GO debe bloquear el vuelo aunque el dron esté perfecto")
	}
}
