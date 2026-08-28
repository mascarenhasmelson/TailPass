package utils

import "time"

type Service struct {
	ID           int        `json:"id"`
	Service_name string     `json:"service_name"`
	LocalIP      string     `json:"local_ip"`
	LocalPort    int        `json:"local_port"`
	RemoteIP     string     `json:"remote_ip"`
	RemotePort   int        `json:"remote_port"`
	Online       bool       `json:"online"`
	Lastseen     *time.Time `json:"last_seen"`
	PID          int        `json:"pid,omitempty"`
	// Running reflects the live in-process tunnel state (not persisted to
	// the DB) so the UI can tell apart "reachable" (Online, via health
	// check) from "tunnel actively forwarding" (Running).
	Running bool `json:"running,omitempty"`
}

// type Home struct {
// 	Status      string `json:"status"`
// 	PublicIP    string `json:"publicip"`
// 	ISPInfo     string `json:"ispinfo"`
// 	Interstatus string `json:"interstatus"`
// }
type IPInfoRaw struct {
	IP  string `json:"ip"`
	Org string `json:"org"`
}

type Error struct {
	Message string
}
