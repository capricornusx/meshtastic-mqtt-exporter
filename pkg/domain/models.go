package domain

import "time"

var DeviceRoles = map[int]string{
	0:  "client",
	1:  "client_mute",
	2:  "router",
	3:  "router_client",
	4:  "repeater",
	5:  "tracker",
	6:  "sensor",
	7:  "tak",
	8:  "client_hidden",
	9:  "lost_and_found",
	10: "tak_tracker",
	11: "router_late",
}

func GetRoleName(role int) string {
	if name, exists := DeviceRoles[role]; exists {
		return name
	}
	return "unknown"
}

type MeshtasticMessage struct {
	From    uint32                 `json:"from"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
	RSSI    *float64               `json:"rssi,omitempty"`
	SNR     *float64               `json:"snr,omitempty"`
}

type TelemetryData struct {
	NodeID    string
	Type      string // device_metrics, environment_metrics, power_metrics
	RSSI      *float64
	SNR       *float64
	Timestamp time.Time

	// Device Metrics
	BatteryLevel       *float64
	Voltage            *float64
	ChannelUtilization *float64
	AirUtilTx          *float64
	UptimeSeconds      *float64

	// Environment Metrics
	Temperature        *float64
	RelativeHumidity   *float64
	BarometricPressure *float64
	GasResistance      *float64
	IAQ                *float64

	// Power Metrics
	Ch1Voltage *float64
	Ch1Current *float64
	Ch2Voltage *float64
	Ch2Current *float64
	Ch3Voltage *float64
	Ch3Current *float64
}

type NodeInfo struct {
	NodeID    string
	LongName  string
	ShortName string
	Hardware  string
	Role      string
	Timestamp time.Time
}

type TextMessage struct {
	NodeID    string
	Text      string
	Channel   int
	Timestamp time.Time
}

type Position struct {
	NodeID        string
	Latitude      *float64
	Longitude     *float64
	Altitude      *int32
	SatsInView    *int32
	PrecisionBits *int32
	Timestamp     time.Time
}

type Waypoint struct {
	NodeID      string
	WaypointID  int32
	Latitude    float64
	Longitude   float64
	Expire      int32
	LockedTo    int32
	Name        string
	Description string
	Icon        int32
	Timestamp   time.Time
}

type NeighborInfo struct {
	NodeID                    string
	NeighborID                string
	SNR                       float64
	LastRxTime                int32
	NodeBroadcastIntervalSecs int32
	Timestamp                 time.Time
}

// Alert for LoRa network.
type Alert struct {
	Severity    string
	Message     string
	Channel     string
	Mode        string
	TargetNodes []uint32
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
