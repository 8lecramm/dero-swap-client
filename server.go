package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/deroproject/derohe/globals"
	"github.com/gorilla/websocket"

	"swap-client/cfg"
	"swap-client/coin"
	"swap-client/dero"
)

type WS_Message struct {
	ID     uint64 `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
	Result any    `json:"result"`
}

var Connection *websocket.Conn

// swap other coins to Dero
func Dero_Swap(request coin.Swap_Request) (response coin.Swap_Response) {

	var err error

	// check if destination wallet is valid. Registered usernames can also be used.
	if strings.HasPrefix(request.DeroAddr, "dero1") || strings.HasPrefix(request.DeroAddr, "deroi") {
		_, err = globals.ParseValidateAddress(request.DeroAddr)
	} else {
		if addr := dero.CheckAddress(request.DeroAddr); addr != "" {
			request.DeroAddr = addr
		} else {
			err = fmt.Errorf("invalid address")
		}
	}

	// basic checks
	if request.Amount == 0 || err != nil {
		response.Error = "invalid request"
		return
	}

	// prevent users from creating too many swap requests
	if Delay.CheckUser(request.DeroAddr) {
		response.Error = "2 minutes wait time triggered"
		return
	}

	// check if pair is enabled and available
	pair := request.Pair
	if !coin.IsPairEnabled(pair) || !IsPairAvailable[pair] {
		response.Error = fmt.Sprintf("%s swap currently not possible", pair)
		return
	}

	// create swap
	err = XTCSwap(pair, request.DeroAddr, coin.RoundFloat(request.Amount, 5), &response)

	if err != nil {
		response.Error = err.Error()
		log.Println(err)
	} else {
		Delay.AddUser(request.DeroAddr)
	}
	response.Request = request

	return response
}

// swap Dero to other coins
func Reverse_Swap(request coin.Swap_Request) (response coin.Swap_Response) {

	var err error

	// prevent users from creating too many swap requests
	if Delay.CheckUser(request.DeroAddr) {
		response.Error = "2 minutes wait time triggered"
		return
	}

	// check if pair is enabled and available
	pair := request.Pair
	if !coin.IsPairEnabled(pair) || !IsPairAvailable[pair] {
		response.Error = fmt.Sprintf("%s swap currently not possible", pair)
		return
	}

	response.Deposit = coin.RoundFloat(request.Amount, 5)

	// create swap
	err = DeroXTCSwap(pair, request.DeroAddr, response.Deposit, &response)

	if err != nil {
		response.Error = err.Error()
		log.Println(err)
	} else {
		Delay.AddUser(request.DeroAddr)
	}
	response.Request = request

	return response
}

func StartClient(server url.URL) {

	var err error

	for {

		var in WS_Message

		dialer := websocket.DefaultDialer
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		Connection, _, err = websocket.DefaultDialer.Dial(server.String(), nil)
		if err != nil {
			log.Println("Websocket error, re-connect in 10 seconds")
			time.Sleep(time.Second * 10)
			continue
		}

		log.Printf("Connected to server %s\n", cfg.Settings.ServerAddress)
		UpdateMarkets()

		for {
			if err := Connection.ReadJSON(&in); err != nil {
				break
			}

			var out WS_Message

			out.ID = in.ID

			switch in.Method {
			case "swap":

				var request coin.Swap_Request
				var response coin.Swap_Response

				p := in.Params.(map[string]any)
				out.Method = "client_ok"

				if d, ok := p["pair"]; ok {
					request.Pair = d.(string)
				}
				if d, ok := p["amount"]; ok {
					request.Amount = d.(float64)
				}
				if d, ok := p["dero_address"]; ok {
					request.DeroAddr = d.(string)
				}

				switch request.Pair {
				case coin.XMRDERO, coin.LTCDERO, coin.BTCDERO, coin.ARRRDERO:
					response = Dero_Swap(request)
				case coin.DEROXMR, coin.DEROLTC, coin.DEROBTC, coin.DEROARRR:
					response = Reverse_Swap(request)
				default:
					return
				}

				out.Result = response

				Connection.WriteJSON(out)
				UpdateMarkets()
			}
		}
		log.Println("Websocket error, re-connect in 10 seconds")
		time.Sleep(time.Second * 10)
	}
}
