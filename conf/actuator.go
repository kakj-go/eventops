package conf

type Actuator struct {
	Dialer Dialer `json:"dialer"`
}

type Dialer struct {
	PeerID          string `json:"peerId"`
	PeerToken       string `json:"peerToken"`
	Peers           string `json:"peers"`
	PrintTunnelData bool   `json:"printTunnelData"`
}
