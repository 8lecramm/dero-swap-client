package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"swap-client/cfg"
	"swap-client/clients"
	"swap-client/coin"
	"swap-client/dero"
	"swap-client/monero"
	"time"
)

func GetMarket(pair string, provider int) (prices Swap_Price) {

	resp, err := http.Get(pair)
	if err != nil {
		log.Printf("Market: HTTP Get: %v\n", err)
		return Swap_Price{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Market: Read Body: %v\n", err)
		return Swap_Price{}
	}

	var market Price_Provider

	if err := json.Unmarshal(body, &market); err != nil {
		log.Printf("Market: Cannot unmarshal data: %v\n", err)
		return Swap_Price{}
	}

	switch provider {
	case TO:
		prices.Ask, err = strconv.ParseFloat(market.Ask, 64)
		prices.Bid, err = strconv.ParseFloat(market.Bid, 64)
		prices.Median, err = strconv.ParseFloat(market.Price, 64)
	case XEGGEX:
		prices.Ask, err = strconv.ParseFloat(market.BestAsk, 64)
		prices.Bid, err = strconv.ParseFloat(market.BestBid, 64)
		prices.Median, err = strconv.ParseFloat(market.LastPrice, 64)
	}
	if err != nil {
		log.Printf("Market: Cannot convert string: %v\n", err)
		return
	}

	return
}

// Get Pair values
// TODO: configurable fees; clean-up
func GetPrice(pair string) (bid float64, ask float64) {

	var base, base_usd Swap_Price
	var atomicUnits uint = 8
	var simple bool

	switch pair {
	case coin.BTCDERO, coin.DEROBTC:
		simple = true
		for i := range DERO_BTC {
			if base = GetMarket(DERO_BTC[i], i); base.Ask > 0 && base.Bid > 0 {
				break
			}
		}
	case coin.LTCDERO, coin.DEROLTC:
		for i := range LTC_USDT {
			if base = GetMarket(LTC_USDT[i], i); base.Ask > 0 && base.Bid > 0 {
				break
			}
		}
	case coin.XMRDERO, coin.DEROXMR:
		atomicUnits = 12
		for i := range XMR_USDT {
			if base = GetMarket(XMR_USDT[i], i); base.Ask > 0 && base.Bid > 0 {
				break
			}
		}
	case coin.ARRRDERO, coin.DEROARRR:
		base = GetMarket(ARRR_USDT, TO)
	}

	if base.Ask == 0 || base.Bid == 0 {
		return 0, 0
	}

	if simple {
		bid = base.Bid - (base.Bid * cfg.SwapFees.Swap.Bid / 100)
		bid = coin.RoundFloat(bid, atomicUnits)
		ask = base.Ask + (base.Ask * cfg.SwapFees.Swap.Ask / 100)
		ask = coin.RoundFloat(ask, atomicUnits)
		return
	}

	for i := range DERO_USDT {
		if base_usd = GetMarket(DERO_USDT[i], i); base_usd.Ask > 0 && base_usd.Bid > 0 {
			break
		}
	}
	if base_usd.Ask == 0 || base_usd.Bid == 0 {
		return 0, 0
	}

	bid = base_usd.Bid - (base_usd.Bid * cfg.SwapFees.Swap.Bid / 100)
	ask = base_usd.Ask + (base_usd.Ask * cfg.SwapFees.Swap.Ask / 100)

	bid = coin.RoundFloat(bid/base.Bid, atomicUnits)
	ask = coin.RoundFloat(ask/base.Ask, atomicUnits)

	return
}

// TODO: simplify
func UpdateMarkets() {

	var xmr, ltc, btc, arrr float64
	var deroxmr, deroltc, derobtc, deroarrr float64

	for p := range coin.SimplePairs {
		switch p {
		case coin.XMRDERO, coin.DEROXMR:
			deroxmr, xmr = GetPrice(p)
		case coin.LTCDERO, coin.DEROLTC:
			deroltc, ltc = GetPrice(p)
		case coin.ARRRDERO, coin.DEROARRR:
			deroarrr, arrr = GetPrice(p)
		case coin.BTCDERO, coin.DEROBTC:
			derobtc, btc = GetPrice(p)
		}

		// sometimes TradeOgre's BID/ASK values are swapped
		if deroxmr > 0 && xmr > 0 && deroxmr > xmr {
			swap := xmr
			xmr = deroxmr
			deroxmr = swap
		}
		if deroltc > 0 && ltc > 0 && deroltc > ltc {
			swap := ltc
			ltc = deroltc
			deroltc = swap
		}
		if derobtc > 0 && btc > 0 && derobtc > btc {
			swap := btc
			btc = derobtc
			derobtc = swap
		}
		if deroarrr > 0 && arrr > 0 && deroarrr > arrr {
			swap := arrr
			arrr = deroarrr
			deroarrr = swap
		}
	}

	mk.Lock()
	defer mk.Unlock()

	// TODO: simplify
	if xmr > 0 {
		mk.Pairs.XMR = xmr
		mk.Update[coin.XMRDERO] = time.Now().UnixMilli()
		IsPairAvailable[coin.XMRDERO] = true
	} else {
		t := time.UnixMilli(mk.Update[coin.XMRDERO])
		if time.Since(t) > time.Minute*2 && coin.Pairs[coin.XMRDERO] {
			IsPairAvailable[coin.XMRDERO] = false
			log.Println("XMR->DERO disabled")
		}
	}
	if deroxmr > 0 {
		mk.Pairs.DEROXMR = deroxmr
		mk.Update[coin.DEROXMR] = time.Now().UnixMilli()
		IsPairAvailable[coin.DEROXMR] = true
	} else {
		t := time.UnixMilli(mk.Update[coin.DEROXMR])
		if time.Since(t) > time.Minute*2 && coin.Pairs[coin.DEROXMR] {
			IsPairAvailable[coin.DEROXMR] = false
			log.Println("DERO->XMR disabled")
		}
	}
	if ltc > 0 {
		mk.Pairs.LTC = ltc
		mk.Update[coin.LTCDERO] = time.Now().UnixMilli()
		IsPairAvailable[coin.LTCDERO] = true
	} else {
		t := time.UnixMilli(mk.Update[coin.LTCDERO])
		if time.Since(t) > time.Minute*2 && coin.Pairs[coin.LTCDERO] {
			IsPairAvailable[coin.LTCDERO] = false
			log.Println("LTC->DERO disabled")
		}
	}
	if deroltc > 0 {
		mk.Pairs.DEROLTC = deroltc
		mk.Update[coin.DEROLTC] = time.Now().UnixMilli()
		IsPairAvailable[coin.DEROLTC] = true
	} else {
		t := time.UnixMilli(mk.Update[coin.DEROLTC])
		if time.Since(t) > time.Minute*2 && coin.Pairs[coin.DEROLTC] {
			IsPairAvailable[coin.DEROLTC] = false
			log.Println("DERO->LTC disabled")
		}
	}
	if btc > 0 {
		mk.Pairs.BTC = btc
		mk.Update[coin.BTCDERO] = time.Now().UnixMilli()
		IsPairAvailable[coin.BTCDERO] = true
	} else {
		t := time.UnixMilli(mk.Update[coin.BTCDERO])
		if time.Since(t) > time.Minute*2 && coin.Pairs[coin.BTCDERO] {
			IsPairAvailable[coin.BTCDERO] = false
			log.Println("BTC->DERO disabled")
		}
	}
	if derobtc > 0 {
		mk.Pairs.DEROBTC = derobtc
		mk.Update[coin.DEROBTC] = time.Now().UnixMilli()
		IsPairAvailable[coin.DEROBTC] = true
	} else {
		t := time.UnixMilli(mk.Update[coin.DEROBTC])
		if time.Since(t) > time.Minute*2 && coin.Pairs[coin.DEROBTC] {
			IsPairAvailable[coin.DEROBTC] = false
			log.Println("DERO->BTC disabled")
		}
	}
	if arrr > 0 {
		mk.Pairs.ARRR = arrr
		mk.Update[coin.ARRRDERO] = time.Now().UnixMilli()
		IsPairAvailable[coin.ARRRDERO] = true
	} else {
		t := time.UnixMilli(mk.Update[coin.ARRRDERO])
		if time.Since(t) > time.Minute*2 && coin.Pairs[coin.ARRRDERO] {
			IsPairAvailable[coin.ARRRDERO] = false
			log.Println("ARRR->DERO disabled")
		}
	}
	if deroarrr > 0 {
		mk.Pairs.DEROARRR = deroarrr
		mk.Update[coin.DEROARRR] = time.Now().UnixMilli()
		IsPairAvailable[coin.DEROARRR] = true
	} else {
		t := time.UnixMilli(mk.Update[coin.DEROARRR])
		if time.Since(t) > time.Minute*2 && coin.Pairs[coin.DEROARRR] {
			IsPairAvailable[coin.DEROARRR] = false
			log.Println("DERO->ARRR disabled")
		}
	}

	mk.Pairs.BTCFees = GetBTCFees()
	balance := UpdatePool()

	var out WS_Message

	out.Method = "client"
	out.Params = balance

	if Connection != nil {
		Connection.WriteJSON(out)
	} else {
		log.Println("<nil> server connection")
	}
}

func UpdatePool() clients.ClientInfo {

	lock.Lock()
	defer lock.Unlock()

	var info clients.ClientInfo
	var pair clients.PairInfo

	info.Nickname = cfg.Settings.Nickname

	for p := range coin.Pairs {
		switch p {
		case coin.DEROLTC, coin.DEROBTC, coin.DEROARRR:
			pair.Balance = coin.XTCGetBalance(p)
			pair.Pair = p
		case coin.DEROXMR:
			pair.Balance = monero.GetBalance()
			pair.Pair = p
		case coin.XMRDERO, coin.LTCDERO, coin.BTCDERO, coin.ARRRDERO:
			pair.Balance = dero.GetBalance()
			pair.Pair = p
		default:
			continue
		}
		info.PairInfo = append(info.PairInfo, pair)
	}

	return info
}

func GetBTCFees() (fees BTC_Fees) {

	resp, err := http.Get(BTC_MEMPOOL)
	if err != nil {
		log.Printf("HTTP Get: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read Body: %v\n", err)
		return
	}

	if err := json.Unmarshal(body, &fees); err != nil {
		log.Printf("Cannot unmarshal data: %v\n", err)
		return
	}

	return
}
