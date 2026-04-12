package sensor

type WaterQuality struct {
	PH          float64 `json:"ph" yaml:"ph"`
	Temperature float64 `json:"temperature" yaml:"temperature"`
	TDS         int     `json:"tds" yaml:"tds"`
	EC          int     `json:"ec" yaml:"ec"`
	ORP         int     `json:"orp" yaml:"orp"`
	Salinity    float64 `json:"salinity" yaml:"salinity"`
	SG          float64 `json:"sg" yaml:"sg"`
	Battery     int     `json:"battery" yaml:"battery"`
}

// Tuya DPS mapping for Kactoily 7-in-1 aquarium sensor:
//
//	dp1  = TDS (ppm, scale 0)
//	dp2  = Temperature (°C, scale 1 — divide by 10)
//	dp7  = Battery (%, scale 0)
//	dp10 = pH (scale 2 — divide by 100)
//	dp11 = EC (uS/cm, scale 0)
//	dp12 = ORP (mV, scale 0)
//	dp101 = Screen Light (bool)
//	dp102 = Salinity (%, scale 2 — divide by 100)
//	dp103 = SG (scale 3 — divide by 1000)
//	dp113 = Temperature unit ("f" or "c")
//	dp129 = Temperature in display unit (°F when dp113="f", scale 1 — divide by 10)

type ProbeResult struct {
	IP        string     `json:"ip" yaml:"ip"`
	OpenPorts []PortInfo `json:"open_ports" yaml:"open_ports"`
	Protocol  string     `json:"protocol" yaml:"protocol"`
	Details   string     `json:"details" yaml:"details"`
}

type PortInfo struct {
	Port     int    `json:"port" yaml:"port"`
	Open     bool   `json:"open" yaml:"open"`
	Service  string `json:"service" yaml:"service"`
	Response string `json:"response,omitempty" yaml:"response,omitempty"`
}
