package agent

type Config struct {
	Address        string `json:"address"`
	ReportInterval int    `json:"report_interval"`
	PollInterval   int    `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}
