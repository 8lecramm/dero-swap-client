package coin

import (
	"encoding/json"
	"log"
	"math"
	"os"
)

func IsPairEnabled(coin string) bool {
	for i := range Pairs {
		if coin == i {
			return true
		}
	}
	return false
}

func IsAmountFree(coin string, amount float64) bool {

	dir_entries, err := os.ReadDir("swaps/active")
	if err != nil {
		return false
	}

	var swap_e Swap_Entry

	for _, e := range dir_entries {
		file_data, err := os.ReadFile("swaps/active/" + e.Name())
		if err != nil {
			return false
		}
		err = json.Unmarshal(file_data, &swap_e)
		if err != nil || (swap_e.Price == amount && swap_e.Coin == coin) {
			return false
		}
	}

	return true
}

func (r *Swap) GetLockedBalance(coin string) float64 {
	r.RLock()
	defer r.RUnlock()

	switch coin {
	case BTCDERO, LTCDERO, ARRRDERO, XMRDERO:
		return r.Dero_balance
	case DEROLTC:
		return r.LTC_balance
	case DEROBTC:
		return r.BTC_balance
	case DEROARRR:
		return r.ARRR_balance
	case DEROXMR:
		return r.XMR_balance
	default:
		return 0
	}
}

func (r *Swap) AddLockedBalance(coin string, amount float64) {
	r.Lock()
	defer r.Unlock()

	switch coin {
	case BTCDERO, LTCDERO, ARRRDERO, XMRDERO:
		r.Dero_balance += amount
	case DEROLTC:
		r.LTC_balance += amount
	case DEROBTC:
		r.BTC_balance += amount
	case DEROARRR:
		r.ARRR_balance += amount
	case DEROXMR:
		r.XMR_balance += amount
	}
}

func (r *Swap) RemoveLockedBalance(coin string, amount float64) {
	r.Lock()
	defer r.Unlock()

	switch coin {
	case BTCDERO, LTCDERO, ARRRDERO, XMRDERO:
		r.Dero_balance -= amount
	case DEROLTC:
		r.LTC_balance -= amount
	case DEROBTC:
		r.BTC_balance -= amount
	case DEROARRR:
		r.ARRR_balance -= amount
	case DEROXMR:
		r.XMR_balance -= amount
	}
}

func (r *Swap) LoadLockedBalance() {

	dir_entries, err := os.ReadDir("swaps/active")
	if err != nil {
		ErrorCheckingOpenSwaps()
	}

	var swap_e Swap_Entry

	for _, e := range dir_entries {
		file_data, err := os.ReadFile("swaps/active/" + e.Name())
		if err != nil {
			ErrorCheckingOpenSwaps()
		}
		err = json.Unmarshal(file_data, &swap_e)
		if err != nil {
			ErrorCheckingOpenSwaps()
		}
		switch swap_e.Coin {
		case LTCDERO, BTCDERO, ARRRDERO, XMRDERO:
			r.AddLockedBalance(swap_e.Coin, swap_e.Amount)
		default:
			r.AddLockedBalance(swap_e.Coin, swap_e.Price)
		}
	}
}

func ErrorCheckingOpenSwaps() {
	log.Println("Can't check reserved amounts")
	os.Exit(1)
}

// round value to X decimal places
func RoundFloat(value float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(value*ratio) / ratio
}
