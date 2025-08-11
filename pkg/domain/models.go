package domain

import "time"

type MeshtasticMessage struct {
	From    uint32                 `json:"from"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type TelemetryData struct {
	NodeID             string
	BatteryLevel       *float64
	Voltage            *float64
	Temperature        *float64
	RelativeHumidity   *float64
	BarometricPressure *float64
	ChannelUtilization *float64
	AirUtilTx          *float64
	UptimeSeconds      *float64
	RSSI               *float64
	SNR                *float64
	Timestamp          time.Time
}

type NodeInfo struct {
	NodeID    string
	LongName  string
	ShortName string
	Hardware  string
	Role      string
	Timestamp time.Time
}

// Alert for LoRa network.
type Alert struct {
	Severity    string
	Message     string
	Channel     string
	Mode        string
	TargetNodes []string
	Timestamp   time.Time
}

type MetricState struct {
	NodeID    string             `json:"node_id"`
	Timestamp int64              `json:"timestamp"`
	Metrics   map[string]float64 `json:"metrics"`
	Labels    map[string]string  `json:"labels"`
}

type StateSnapshot struct {
	Version   string        `json:"version"`
	Timestamp int64         `json:"timestamp"`
	Nodes     []MetricState `json:"nodes"`
}
