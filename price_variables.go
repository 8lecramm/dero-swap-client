package main

import "sync"

type (
	Price_Provider struct {
		// TradeOgre fields below
		Success bool   `json:"success"`
		Init    string `json:"initialprice"`
		Price   string `json:"price"`
		High    string `json:"high"`
		Low     string `json:"low"`
		Volume  string `json:"volume"`
		Bid     string `json:"bid"`
		Ask     string `json:"ask"`
		// Xeggex fields below
		BestAsk   string `json:"bestAsk"`
		BestBid   string `json:"bestBid"`
		LastPrice string `json:"lastPrice"`
	}
	Swap_Price struct {
		Ask    float64
		Bid    float64
		Median float64
	}
	BTC_Fees struct {
		Fastest  uint64 `json:"fastestFee"`
		Hour     uint64 `json:"hourFee"`
		HalfHour uint64 `json:"halfHourFee"`
	}
	Swap_Markets struct {
		BTC      float64  `json:"btcdero,omitempty"`
		LTC      float64  `json:"ltcdero,omitempty"`
		ARRR     float64  `json:"arrrdero,omitempty"`
		XMR      float64  `json:"xmrdero,omitempty"`
		DEROLTC  float64  `json:"deroltc,omitempty"`
		DEROBTC  float64  `json:"derobtc,omitempty"`
		DEROARRR float64  `json:"deroarrr,omitempty"`
		DEROXMR  float64  `json:"deroxmr,omitempty"`
		BTCFees  BTC_Fees `json:"btc_fees"`
	}
	Swap_Balance struct {
		Dero float64 `json:"dero"`
		LTC  float64 `json:"ltc,omitempty"`
		BTC  float64 `json:"btc,omitempty"`
		ARRR float64 `json:"arrr,omitempty"`
		XMR  float64 `json:"xmr,omitempty"`
	}
)

const (
	ASK = iota
	BID
	MEDIAN
)
const (
	TO = iota
	XEGGEX
)
const SATOSHI float64 = 1e-08

type MarketData struct {
	Pairs  Swap_Markets
	Update map[string]int64
	sync.RWMutex
}

var mk = &MarketData{Update: make(map[string]int64)}
var IsPairAvailable = make(map[string]bool)
var lock sync.Mutex

// Price URLs
var DERO_USDT = []string{"https://tradeogre.com/api/v1/ticker/DERO-USDT", "https://api.xeggex.com/api/v2/market/getbysymbol/DERO/USDT"}
var DERO_BTC = []string{"https://tradeogre.com/api/v1/ticker/DERO-BTC", "https://api.xeggex.com/api/v2/market/getbysymbol/DERO/BTC"}
var LTC_USDT = []string{"https://tradeogre.com/api/v1/ticker/LTC-USDT", "https://api.xeggex.com/api/v2/market/getbysymbol/LTC/USDT"}
var XMR_USDT = []string{"https://tradeogre.com/api/v1/ticker/XMR-USDT", "https://api.xeggex.com/api/v2/market/getbysymbol/XMR/USDT"}
var ARRR_USDT = "https://tradeogre.com/api/v1/ticker/ARRR-USDT"

type ORDER int

// BTC mempool fees
const BTC_MEMPOOL = "https://mempool.space/api/v1/fees/recommended"
