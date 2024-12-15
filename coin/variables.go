package coin

import (
	"net/http"
	"sync"
)

type (
	Swap_Request struct {
		Pair     string  `json:"pair"`
		Amount   float64 `json:"amount"`
		DeroAddr string  `json:"dero_address"`
		Extern   bool    `json:"extern,omitempty"`
	}
	Swap_Response struct {
		ID      int64        `json:"id"`
		Wallet  string       `json:"wallet,omitempty"`
		Deposit float64      `json:"deposit,omitempty"`
		Swap    float64      `json:"swap,omitempty"`
		Error   string       `json:"error,omitempty"`
		Request Swap_Request `json:"request"`
	}
	Swap_Tracking struct {
		ID    int64  `json:"id"`
		State uint64 `json:"state"`
	}
	Swap_Entry struct {
		Coin        string  `json:"coin"`
		Wallet      string  `json:"wallet"`
		Destination string  `json:"destination"`
		Amount      float64 `json:"amount"`
		Price       float64 `json:"price"`
		Fee         uint64  `json:"fee"`
		Created     int64   `json:"created"`
		Block       uint64  `json:"block"`
		Balance     float64 `json:"balance"`
		Status      uint64  `json:"status"`
		Txid        string  `json:"txid"`
	}
	Swap struct {
		Dero_balance float64
		LTC_balance  float64
		BTC_balance  float64
		ARRR_balance float64
		XMR_balance  float64
		sync.RWMutex
	}
)

var Locked Swap
var Supported_pairs = []string{BTCDERO, LTCDERO, ARRRDERO, XMRDERO, DEROBTC, DEROLTC, DEROARRR, DEROXMR}
var Pairs = make(map[string]bool)
var SimplePairs = make(map[string]bool)
var XTC_URL = make(map[string]string)
var XTC_Daemon = &http.Client{}
var XTC_auth string

var BTC_address, LTC_address, ARRR_address, XMR_address string
var BTC_Dir, LTC_Dir, ARRR_Dir string
