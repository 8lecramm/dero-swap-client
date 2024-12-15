package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"swap-client/cfg"
	"swap-client/coin"
	"swap-client/dero"
	"swap-client/monero"
	"time"
)

func CreateSwap(pair string, wallet string, dest string, amount float64, price float64, fee uint64) int64 {

	var entry coin.Swap_Entry
	var height uint64
	var payout float64 = amount

	// get current block height. Ignore transactions < height
	switch pair {
	case coin.BTCDERO, coin.LTCDERO, coin.ARRRDERO:
		height = coin.XTCCheckBlockHeight(pair)
	case coin.XMRDERO:
		height = monero.GetHeight()
	default:
		height = dero.CheckBlockHeight()
	}

	if height == 0 {
		return 0
	}

	entry.Coin = pair
	entry.Wallet = wallet
	entry.Destination = dest
	entry.Price = price
	entry.Amount = amount
	entry.Created = time.Now().UnixMilli()
	entry.Block = height
	entry.Status = 0

	if pair == coin.DEROBTC {
		entry.Fee = fee
	}

	// create an integrated address for all Dero -> X swaps
	if pair == coin.DEROLTC || pair == coin.DEROBTC || pair == coin.DEROARRR || pair == coin.DEROXMR {
		if entry.Wallet = dero.MakeIntegratedAddress(entry.Created); entry.Wallet == "" {
			return 0
		}
		payout = entry.Price
	}

	json_bytes, err := json.Marshal(&entry)
	if err != nil {
		return 0
	}

	err = os.WriteFile(fmt.Sprintf("swaps/active/%d", entry.Created), json_bytes, 0644)
	if err != nil {
		return 0
	}

	log.Printf("Swap request (%d) of %.8f (%s) successfully created\n", entry.Created, payout, entry.Coin)

	return entry.Created
}

// X to Dero Swaps
func XTCSwap(pair string, dst_addr string, amount float64, resp *coin.Swap_Response) (err error) {

	// check if Dero address is registered on chain
	if !dero.IsDeroAddressRegistered(dst_addr) {
		return fmt.Errorf("dero address is not registered")
	}

	// check balance and include locked swap balance
	balance, err := dero.CheckWalletBalance()
	if err != nil {
		return fmt.Errorf("couldn't check swap balance")
	}
	if coin.Locked.GetLockedBalance(pair)+amount+dero.TxFee > balance {
		return fmt.Errorf("insufficient swap balance")
	}

	coin.Locked.AddLockedBalance(pair, amount)

	// create/get a deposit address
	switch pair {
	case coin.BTCDERO:
		resp.Wallet = coin.BTC_address
	case coin.LTCDERO:
		resp.Wallet = coin.LTC_address
	case coin.XMRDERO:
		resp.Wallet = monero.MakeIntegratedAddress()
	case coin.ARRRDERO:
		resp.Wallet = coin.ARRR_address
	}
	if resp.Wallet == "" {
		return fmt.Errorf("no swap deposit address available")
	}

	var coin_value float64
	var atomicUnits uint = 8

	switch pair {
	case coin.BTCDERO:
		coin_value = mk.Pairs.BTC
	case coin.LTCDERO:
		coin_value = mk.Pairs.LTC
	case coin.ARRRDERO:
		coin_value = mk.Pairs.ARRR
	case coin.XMRDERO:
		coin_value = mk.Pairs.XMR
		atomicUnits = 12
	}

	deposit_value := coin_value * amount

	var loops int = 5
	var isAvailable bool

	// if there is a request with the same deposit amount, run in a loop and lower deposit value by 1 Sat
	for i := 0; i < loops; i++ {
		if coin.IsAmountFree(pair, deposit_value) {
			isAvailable = true
			break
		}
		deposit_value -= SATOSHI
	}
	if !isAvailable || deposit_value == 0 {
		return fmt.Errorf("Pre-Check: Couldn't create swap")
	}

	deposit_value = coin.RoundFloat(deposit_value, atomicUnits)

	resp.ID = CreateSwap(pair, resp.Wallet, dst_addr, amount, deposit_value, 0)
	if resp.ID == 0 {
		return fmt.Errorf("couldn't create swap")
	}
	resp.Deposit = deposit_value

	return nil
}

// Dero to X swaps
func DeroXTCSwap(pair string, dst_addr string, amount float64, resp *coin.Swap_Response) (err error) {

	var balance float64
	var atomicUnits uint = 8

	// validate destination wallet and check for sufficient swap balance
	if pair != coin.DEROXMR {
		if !coin.XTCValidateAddress(pair, dst_addr) {
			return fmt.Errorf("%s address is not valid", pair[5:])
		}
		balance = coin.XTCGetBalance(pair)
	} else {
		if !monero.ValidateAddress(dst_addr) {
			return fmt.Errorf("XMR address is not valid")
		}
		balance = monero.GetBalance()
		atomicUnits = 12
	}

	var coin_value float64
	var fees float64
	var btc_fees BTC_Fees

	// determine fees and current price
	switch pair {
	case coin.DEROLTC:
		coin_value = mk.Pairs.DEROLTC
		fees = cfg.SwapFees.Withdrawal.DeroLTC
	case coin.DEROARRR:
		coin_value = mk.Pairs.DEROARRR
		fees = cfg.SwapFees.Withdrawal.DeroARRR
	case coin.DEROBTC:
		coin_value = mk.Pairs.DEROBTC
		btc_fees = GetBTCFees()
		fees = float64((btc_fees.HalfHour*141)+(500-((btc_fees.HalfHour*141)%500))) / 100000000
	case coin.DEROXMR:
		coin_value = mk.Pairs.DEROXMR
		fees = cfg.SwapFees.Withdrawal.DeroXMR
	}

	payout_value := coin_value * amount
	if payout_value-fees < 0 {
		return fmt.Errorf("fees > payout value")
	}
	if payout_value == 0 || fees == 0 {
		return fmt.Errorf("couldn't create swap")
	}

	// check for reserved balance
	if coin.Locked.GetLockedBalance(pair)+payout_value+fees > balance {
		return fmt.Errorf("insufficient swap balance")
	}

	coin.Locked.AddLockedBalance(pair, payout_value)

	payout_value -= fees
	payout_value = coin.RoundFloat(payout_value, atomicUnits)
	resp.Swap = payout_value

	resp.ID = CreateSwap(pair, "", dst_addr, amount, payout_value, btc_fees.HalfHour)
	if resp.ID == 0 {
		return fmt.Errorf("couldn't create swap")
	}
	resp.Wallet = dero.MakeIntegratedAddress(resp.ID)

	return nil
}
