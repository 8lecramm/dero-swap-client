package cfg

import "net/url"

type Config struct {
	ServerAddress string   `json:"server"`
	Nickname      string   `json:"nickname"`
	Dero_Daemon   string   `json:"Dero_Daemon"`
	Dero_Wallet   string   `json:"dero_wallet"`
	Dero_Login    string   `json:"dero_login"`
	Monero_Daemon string   `json:"monero_daemon"`
	Monero_Wallet string   `json:"Monero_Wallet"`
	LTC_Daemon    string   `json:"LTC_Daemon"`
	LTC_Dir       string   `json:"LTC_Dir"`
	BTC_Daemon    string   `json:"BTC_Daemon"`
	BTC_Dir       string   `json:"BTC_Dir"`
	ARRR_Daemon   string   `json:"ARRR_Daemon"`
	ARRR_Dir      string   `json:"ARRR_Dir"`
	Pairs         []string `json:"pairs"`
	//SwapFees      float64 `json:"swap_fees"`
}

type (
	Fees struct {
		Swap       Swap_Fees       `json:"swap"`
		Withdrawal Withdrawal_Fees `json:"withdrawal"`
	}
	/*
		Swap_Fees struct {
			DeroLTC  float64 `json:"dero-ltc"`
			DeroBTC  float64 `json:"dero-btc"`
			DeroARRR float64 `json:"dero-arrr"`
			DeroXMR  float64 `json:"dero-xmr"`
			LTCDero  float64 `json:"ltc-dero"`
			BTCDero  float64 `json:"btc-dero"`
			ARRRDero float64 `json:"arrr-dero"`
			XMRDero  float64 `json:"xmr-dero"`
		}
	*/
	Swap_Fees struct {
		Bid float64 `json:"bid"`
		Ask float64 `json:"ask"`
	}
	Withdrawal_Fees struct {
		DeroLTC  float64 `json:"dero-ltc"`
		DeroBTC  float64 `json:"dero-btc"`
		DeroARRR float64 `json:"dero-arrr"`
		DeroXMR  float64 `json:"dero-xmr"`
	}
)

var Settings Config
var SwapFees Fees

var Server_URL url.URL
