package cfg

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"
	"swap-client/coin"
	"swap-client/dero"
	"swap-client/monero"

	"github.com/ybbus/jsonrpc/v3"
)

// load config file
func LoadConfig() {

	fd, err := os.ReadFile("config.json")
	if err != nil {
		log.Printf("Error loading config file: %v\n", err)
		return
	}
	err = json.Unmarshal(fd, &Settings)
	if err != nil {
		log.Printf("Error parsing config file: %v\n", err)
		return
	}

	dero.Dero_Daemon = jsonrpc.NewClient("http://" + Settings.Dero_Daemon + "/json_rpc")
	if Settings.Dero_Login != "" {
		dero.RPC_Login = base64.StdEncoding.EncodeToString([]byte(Settings.Dero_Login))
		dero.Dero_Wallet = jsonrpc.NewClientWithOpts("http://"+Settings.Dero_Wallet+"/json_rpc", &jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + dero.RPC_Login,
			},
		})
	} else {
		dero.Dero_Wallet = jsonrpc.NewClient("http://" + Settings.Dero_Wallet + "/json_rpc")
		log.Println("Dero Wallet: No RPC authorization specified")
	}

	Server_URL = url.URL{Scheme: "wss", Host: Settings.ServerAddress, Path: "/ws"}

	monero.Monero_Wallet = jsonrpc.NewClient("http://" + Settings.Monero_Wallet + "/json_rpc")

	coin.XTC_URL[coin.BTCDERO] = "http://" + Settings.BTC_Daemon
	coin.XTC_URL[coin.LTCDERO] = "http://" + Settings.LTC_Daemon
	coin.XTC_URL[coin.ARRRDERO] = "http://" + Settings.ARRR_Daemon

	// check if pair is "supported"
	for _, p := range Settings.Pairs {
		supported := false
		for i := range coin.Supported_pairs {
			if p == coin.Supported_pairs[i] {
				supported = true
				break
			}
		}
		if supported {
			coin.Pairs[p] = true
		} else {
			log.Printf("%s is not a supported pair\n", p)
		}
	}

	log.Printf("Config successfully loaded\n")

	LoadFees()
}

// load fees file
func LoadFees() {

	fd, err := os.ReadFile("fees.json")
	if err != nil {
		log.Printf("Error loading fees file: %v\n", err)
		return
	}
	err = json.Unmarshal(fd, &SwapFees)
	if err != nil {
		log.Printf("Error parsing fees file: %v\n", err)
		return
	}

	log.Printf("%-14s: Buy: %.2f%% / Sell: %.2f%%\n", "Fees", SwapFees.Swap.Ask, SwapFees.Swap.Bid)
}

// basic config check
func CheckConfig() bool {

	if Settings.Nickname == "" {
		log.Println("Please specify a nickname")
		return false
	}

	if Settings.Dero_Daemon == "" || Settings.Dero_Wallet == "" {
		log.Println("Dero Daemon or Dero Wallet is not set")
		return false
	}

	for p := range coin.Pairs {
		switch p {
		case coin.XMRDERO, coin.DEROXMR:
			if Settings.Monero_Wallet == "" {
				log.Printf("%s pair is set, but wallet is not set\n", p)
				return false
			}
		case coin.LTCDERO, coin.DEROLTC:
			if Settings.LTC_Daemon == "" || Settings.LTC_Dir == "" {
				log.Printf("%s pair is set, but daemon or directory is not set\n", p)
				return false
			} else {
				coin.LTC_Dir = Settings.LTC_Dir
			}
		case coin.BTCDERO, coin.DEROBTC:
			if Settings.BTC_Daemon == "" || Settings.BTC_Dir == "" {
				log.Printf("%s pair is set, but daemon or directory is not set\n", p)
				return false
			} else {
				coin.BTC_Dir = Settings.BTC_Dir
			}
		case coin.ARRRDERO, coin.DEROARRR:
			if Settings.ARRR_Daemon == "" || Settings.ARRR_Dir == "" {
				log.Printf("%s pair is set, but daemon or directory is not set\n", p)
				return false
			} else {
				coin.ARRR_Dir = Settings.ARRR_Dir
			}
		}
	}

	if dero.RPC_Login != "" {
		log.Printf("%-14s: %s\n", "Dero Wallet", "Using RPC authorization")
	}
	dero_wallet := dero.GetAddress()

	if dero.GetHeight() == 0 || dero.CheckBlockHeight() == 0 || dero_wallet == "" {
		log.Println("Dero daemon or wallet is not available")
		return false
	}
	log.Printf("%-14s: %s\n", "Dero Wallet", dero_wallet)

	return true
}

// TODO: simplify
func LoadWallets() {

	for p := range coin.Pairs {
		switch p {
		case coin.ARRRDERO, coin.DEROARRR:
			addr := coin.ARRR_GetAddress()
			if !coin.XTCValidateAddress(p, addr) {
				log.Printf("Disable pair \"%s\": wallet not available or other error\n", p)
				delete(coin.Pairs, p)
			} else {
				if coin.ARRR_address == "" {
					coin.ARRR_address = addr
					coin.SimplePairs[p] = true
					log.Printf("%-14s: %s\n", "ARRR Wallet", addr)
				}
			}
		case coin.XMRDERO, coin.DEROXMR:
			addr := monero.GetAddress()
			if !monero.ValidateAddress(addr) {
				log.Printf("Disable pair \"%s\": wallet not available or other error\n", p)
				delete(coin.Pairs, p)
			} else {
				if coin.XMR_address == "" {
					coin.XMR_address = addr
					coin.SimplePairs[p] = true
					log.Printf("%-14s: %s\n", "XMR Wallet", addr)
				}
			}
		case coin.LTCDERO, coin.DEROLTC:
			ok, err := coin.XTCLoadWallet(p)
			if !ok && !strings.Contains(err, "is already loaded") {
				ok, err := coin.XTCNewWallet(p)
				if !ok {
					log.Printf("Disable pair \"%s\": %s\n", p, err)
					delete(coin.Pairs, p)
				} else {
					addr := coin.XTCGetAddress(p)
					if coin.LTC_address == "" {
						coin.LTC_address = addr
						coin.SimplePairs[p] = true
						log.Printf("%-14s: %s\n", "LTC Wallet", addr)
					}
				}
			} else {
				addr := coin.XTCGetAddress(p)
				if coin.LTC_address == "" {
					coin.LTC_address = addr
					coin.SimplePairs[p] = true
					log.Printf("%-14s: %s\n", "LTC Wallet", addr)
				}
			}
		case coin.BTCDERO, coin.DEROBTC:
			ok, err := coin.XTCLoadWallet(p)
			if !ok && !strings.Contains(err, "is already loaded") {
				ok, err := coin.XTCNewWallet(p)
				if !ok {
					log.Printf("Disable pair \"%s\": %s\n", p, err)
					delete(coin.Pairs, p)
				} else {
					addr := coin.XTCGetAddress(p)
					if coin.BTC_address == "" {
						coin.BTC_address = addr
						coin.SimplePairs[p] = true
						log.Printf("%-14s: %s\n", "BTC Wallet", addr)
					}
				}
			} else {
				addr := coin.XTCGetAddress(p)
				if coin.BTC_address == "" {
					coin.BTC_address = addr
					coin.SimplePairs[p] = true
					log.Printf("%-14s: %s\n", "BTC Wallet", addr)
				}
			}
		default:
			continue
		}
	}
}
