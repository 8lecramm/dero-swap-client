package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"swap-client/coin"
	"swap-client/dero"
	"swap-client/monero"
	"time"

	"github.com/deroproject/derohe/rpc"
)

const (
	SWAP_CREATED = iota
	SWAP_CONFIRMED
	SWAP_DONE
	SWAP_EXPIRED
	SWAP_TOO_OLD
	SWAP_UNDEFINED
)

func Swap_Controller() {

	var file_data []byte
	var swap_e coin.Swap_Entry
	var expired, fails, sent, active uint

	var txs []rpc.Transfer
	var xmr_txs []monero.RPC_XMR_Transfer_Params

	var dir_entries []fs.DirEntry
	var err error

	for {

		time.Sleep(time.Minute)
		dir_entries, err = os.ReadDir("swaps/active")
		if err != nil {
			log.Println("Can't list swap entries")
			continue
		}

		expired = 0
		fails = 0
		sent = 0
		active = 0
		txs = nil
		xmr_txs = nil

		for _, e := range dir_entries {

			active++
			file_data = nil
			err = nil

			file_data, err = os.ReadFile("swaps/active/" + e.Name())
			if err != nil {
				fails++
				continue
			}
			err = json.Unmarshal(file_data, &swap_e)
			if err != nil {
				fails++
				continue
			}
			creation_t := time.UnixMilli(swap_e.Created)

			// if there was no deposit, mark the request as expired
			if swap_e.Status == SWAP_CREATED && time.Since(creation_t) > time.Hour {
				os.WriteFile(fmt.Sprintf("swaps/expired/%d", swap_e.Created), file_data, 0644)
				os.Remove("swaps/active/" + e.Name())
				switch swap_e.Coin {
				case coin.LTCDERO, coin.BTCDERO, coin.ARRRDERO, coin.XMRDERO:
					coin.Locked.RemoveLockedBalance(swap_e.Coin, swap_e.Amount)
				default:
					coin.Locked.RemoveLockedBalance(swap_e.Coin, swap_e.Price)
				}

				expired++
				continue
			}

			var found_deposit, visible bool

			// check for deposits
			switch swap_e.Coin {
			case coin.BTCDERO, coin.LTCDERO, coin.ARRRDERO:
				found_deposit, visible, _, err = coin.XTCListReceivedByAddress(swap_e.Coin, swap_e.Wallet, swap_e.Price, swap_e.Block, false)
			case coin.XMRDERO:
				if payment_id := monero.SplitIntegratedAddress(swap_e.Wallet); payment_id != "" {
					found_deposit = monero.XMRGetTX(payment_id, swap_e.Block)
					visible = found_deposit
				} else {
					log.Println("Can't split integrated XMR address")
				}
			default:
				found_deposit = dero.CheckIncomingTransfers(uint64(swap_e.Created), swap_e.Block)
				visible = found_deposit
			}

			if err != nil {
				log.Printf("Error checking incoming %s transactions\n", swap_e.Coin)
				fails++
				continue
			}

			// mark request as done
			if swap_e.Status == SWAP_DONE {
				err = os.WriteFile(fmt.Sprintf("swaps/done/%d", swap_e.Created), file_data, 0644)
				if err != nil {
					log.Printf("Can't mark swap as done, swap %d, err %v\n", swap_e.Created, err)
				} else {
					os.Remove("swaps/active/" + e.Name())
				}
			}

			// start payout if there are at least 2 confirmations
			// requests won't be marked as expired, if there is already 1 confirmation
			if visible {
				log.Printf("Found TX for ID %d (%s) on chain\n", swap_e.Created, swap_e.Coin)
				if found_deposit && swap_e.Status <= SWAP_CONFIRMED {
					// create transaction
					log.Printf("Found deposit for ID %d (%s): %.8f coins; adding to payout TX\n", swap_e.Created, swap_e.Coin, swap_e.Amount)

					switch swap_e.Coin {
					case coin.DEROLTC, coin.DEROBTC:
						log.Println("Starting LTC/BTC payout")
						_, txid := coin.XTCSend(swap_e.Coin, swap_e.Destination, swap_e.Price, swap_e.Fee)
						log.Printf("LTC/BTC TXID: %s\n", txid)
						coin.Locked.RemoveLockedBalance(swap_e.Coin, swap_e.Price)
					case coin.DEROARRR:
						log.Println("Starting ARRR payout")
						ok, result := coin.ARRR_Send(swap_e.Destination, swap_e.Price)
						log.Printf("ARRR status: %v, %s\n", ok, result)
						coin.Locked.RemoveLockedBalance(swap_e.Coin, swap_e.Price)
					case coin.DEROXMR:
						xmr_txs = append(xmr_txs, monero.AddTX(swap_e.Destination, swap_e.Price))
					default:
						txs = append(txs, dero.AddTX(swap_e.Destination, swap_e.Amount))
					}

					swap_e.Status = SWAP_DONE
					sent++
					active--
				} else {
					// transaction was confirmed
					swap_e.Status = SWAP_CONFIRMED
				}

				json_data, _ := json.Marshal(&swap_e)
				os.WriteFile("swaps/active/"+e.Name(), json_data, 0644)

				if swap_e.Status == SWAP_DONE {
					err = os.WriteFile(fmt.Sprintf("swaps/done/%d", swap_e.Created), file_data, 0644)
					if err != nil {
						log.Printf("Can't mark swap as done, swap %d, err %v\n", swap_e.Created, err)
					} else {
						os.Remove("swaps/active/" + e.Name())
					}
				}
			}
		}

		// Dero and Monero payout process
		if len(txs) > 0 {
			log.Println("Starting DERO payout process")
			dero.Payout(txs)
		}
		// TODO: create function and TX verification
		if len(xmr_txs) > 0 {
			log.Println("Starting XMR payout process")
			if ok, txid := monero.XMRSend(xmr_txs); ok {
				log.Printf("XMR transaction (TXID %s) successfully sent\n", txid)
			} else {
				log.Println("Error sending XMR transaction")
			}
		}

		if sent+expired+fails > 0 {
			log.Printf("Swap processing: %d sent, %d expired, %d errors\n", sent, expired, fails)
		}
	}
}

func SwapTracking(session int64) uint64 {

	if time.Since(time.UnixMilli(int64(session))) > time.Hour*24*3 {
		return SWAP_TOO_OLD
	}
	if _, err := os.ReadFile("swaps/expired/" + fmt.Sprintf("%d", session)); err == nil {
		return SWAP_EXPIRED
	}
	if _, err := os.ReadFile("swaps/done/" + fmt.Sprintf("%d", session)); err == nil {
		return SWAP_DONE
	}

	var swap coin.Swap_Entry
	if fd, err := os.ReadFile("swaps/active/" + fmt.Sprintf("%d", session)); err == nil {
		if err = json.Unmarshal(fd, &swap); err == nil {
			return swap.Status
		} else {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}

	return SWAP_UNDEFINED
}
