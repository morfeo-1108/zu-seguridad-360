package flight

import "fmt"

// PreflightCheck es un ítem del checklist que un piloto profesional verifica
// antes de cada despegue. Si alguno crítico falla, NO se vuela.
type PreflightCheck struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Critical bool   `json:"critical"`
	Detail   string `json:"detail"`
}

// DroneState es la telemetría de estado para el pre-vuelo.
type DroneState struct {
	BatteryPct        int  `json:"battery_pct"`
	BatteryHealthPct  int  `json:"battery_health_pct"`
	GPSSatellites     int  `json:"gps_satellites"`
	CompassCalibrated bool `json:"compass_calibrated"`
	IMUCalibrated     bool `json:"imu_calibrated"`
	FirmwareOK        bool `json:"firmware_ok"`
	GimbalOK          bool `json:"gimbal_ok"`
	SDCardFreeMB      int  `json:"sd_card_free_mb"`
	HomePointSet      bool `json:"home_point_set"`
	SignalStrength    int  `json:"signal_strength"` // 0-100
	MotorsOK          bool `json:"motors_ok"`
}

// PreflightResult: ¿se autoriza el despegue?
type PreflightResult struct {
	Cleared bool             `json:"cleared"`
	Checks  []PreflightCheck `json:"checks"`
	Reason  string           `json:"reason"`
}

// RunPreflight ejecuta el checklist completo, como la lista de un piloto.
func RunPreflight(s DroneState, env FlightEnvelope, weather WeatherAssessment, missionDistance float64) PreflightResult {
	checks := []PreflightCheck{
		{Name: "Batería suficiente", Passed: s.BatteryPct >= 50, Critical: true,
			Detail: fmt.Sprintf("%d%% (mínimo 50%% para iniciar)", s.BatteryPct)},
		{Name: "Salud de batería", Passed: s.BatteryHealthPct >= 80, Critical: true,
			Detail: fmt.Sprintf("%d%% de capacidad", s.BatteryHealthPct)},
		{Name: "Fix GPS robusto", Passed: s.GPSSatellites >= env.MinGPSSatellites, Critical: true,
			Detail: fmt.Sprintf("%d satélites", s.GPSSatellites)},
		{Name: "Brújula calibrada", Passed: s.CompassCalibrated, Critical: true, Detail: ""},
		{Name: "IMU calibrada", Passed: s.IMUCalibrated, Critical: true, Detail: ""},
		{Name: "Firmware actualizado", Passed: s.FirmwareOK, Critical: true, Detail: ""},
		{Name: "Gimbal operativo", Passed: s.GimbalOK, Critical: true, Detail: ""},
		{Name: "Motores OK", Passed: s.MotorsOK, Critical: true, Detail: ""},
		{Name: "Home point fijado", Passed: s.HomePointSet, Critical: true, Detail: "RTH requiere home point"},
		{Name: "Señal de control", Passed: s.SignalStrength >= 70, Critical: true,
			Detail: fmt.Sprintf("%d%%", s.SignalStrength)},
		{Name: "Almacenamiento", Passed: s.SDCardFreeMB >= 500, Critical: false,
			Detail: fmt.Sprintf("%d MB libres", s.SDCardFreeMB)},
		{Name: "Clima apto (GO/NO-GO)", Passed: weather.GoForFlight, Critical: true,
			Detail: weather.Summary},
		{Name: "Reserva para misión completa", Passed: BatteryReserveOK(s.BatteryPct, missionDistance, env),
			Critical: true, Detail: "energía para misión + regreso + margen"},
	}

	result := PreflightResult{Cleared: true, Checks: checks}
	for _, c := range checks {
		if c.Critical && !c.Passed {
			result.Cleared = false
			if result.Reason == "" {
				result.Reason = "Falló verificación crítica: " + c.Name
			}
		}
	}
	return result
}
