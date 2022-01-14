package config

type SystemInfo struct {
	AgentId      string   `json:"agent_id"`
	HostName     string   `json:"host_name"`
	UserName     string   `json:"user_name"`
	Platform     string   `json:"platform"`
	UserGID      string   `json:"user_gid"`
	IPS          []string `json:"i_ps"`
	Architecture string   `json:"architecture"`
	PID          int      `json:"pid"`
	Response     []byte   `json:"response"`
	ResponseURL  string   `json:"response_url"`
}
