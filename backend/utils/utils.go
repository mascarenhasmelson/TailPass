package utils

type Service struct {
	ID           int    `json:"id"`
	Service_name string `json:"service_name"`
	LocalIP      string `json:"local_ip"`
	LocalPort    int    `json:"local_port"`
	RemoteIP     string `json:"remote_ip"`
	RemotePort   int    `json:"remote_port"`
	PID          int    `json:"pid"`
}
