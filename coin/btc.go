package coin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"net/http"
)

func XTCGetCookie(pair string) bool {

	var data []byte
	var err error

	switch pair {
	case BTCDERO, DEROBTC:
		if data, err = os.ReadFile(BTC_Dir + "/.cookie"); err != nil {
			log.Printf("Can't load BTC auth cookie: %v\n", err)
			return false
		}
	case LTCDERO, DEROLTC:
		if data, err = os.ReadFile(LTC_Dir + "/.cookie"); err != nil {
			log.Printf("Can't load LTC auth cookie: %v\n", err)
			return false
		}
	case ARRRDERO, DEROARRR:
		if data, err = os.ReadFile(ARRR_Dir + "/.cookie"); err != nil {
			log.Printf("Can't load ARRR auth cookie: %v\n", err)
			return false
		}
	}

	XTC_auth = base64.StdEncoding.EncodeToString(data)
	data = nil

	return true
}

func SetHeaders(request *http.Request, auth string) {

	request.Header.Add("Content-Type", "text/plain")
	request.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))
}

func XTCBuildRequest(pair string, method string, options []interface{}) (*http.Request, error) {

	json_object := &RPC_Request{Jsonrpc: "1.0", Id: "swap", Method: method, Params: options}
	json_bytes, err := json.Marshal(&json_object)
	if err != nil {
		log.Printf("Can't marshal %s request: %v\n", method, err)
		return nil, err
	}

	req, err := http.NewRequest("POST", XTC_GetURL(pair), bytes.NewBuffer(json_bytes))
	if err != nil {
		log.Printf("Can't create %s request: %v\n", method, err)
		return nil, err
	}

	XTCGetCookie(pair)
	SetHeaders(req, XTC_auth)

	return req, nil
}

func XTCNewWallet(pair string) (result bool, message string) {

	switch pair {
	case DEROARRR, ARRRDERO, DEROXMR, XMRDERO:
		return false, ""
	}
	req, err := XTCBuildRequest(pair, "createwallet", []interface{}{"swap_wallet"})
	if err != nil {
		log.Printf("Can't build createwallet request: %v\n", err)
		return false, ""
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send createwallet request: %v\n", err)
		return false, ""
	}
	defer resp.Body.Close()

	var json_resonse RPC_NewWallet_Response

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &json_resonse)
	if err != nil {
		log.Printf("Can't unmarshal createwallet response: %v\n", err)
		return false, ""
	}

	if json_resonse.Error != (RPC_Error{}) {
		message = json_resonse.Error.Message
	} else {
		result = true
	}

	return result, message
}

func XTCLoadWallet(pair string) (result bool, message string) {

	req, err := XTCBuildRequest(pair, "loadwallet", []interface{}{"swap_wallet"})
	if err != nil {
		log.Printf("Can't build loadwallet request: %v\n", err)
		return false, ""
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send loadwallet request: %v\n", err)
		return false, ""
	}
	defer resp.Body.Close()

	var json_resonse RPC_NewWallet_Response

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &json_resonse)
	if err != nil {
		log.Printf("Can't unmarshal createwallet response: %v\n", err)
		return false, ""
	}

	if json_resonse.Error != (RPC_Error{}) {
		message = json_resonse.Error.Message
	} else {
		result = true
	}

	return result, message
}

func XTCNewAddress(pair string) string {

	req, err := XTCBuildRequest(pair, "getnewaddress", []interface{}{})
	if err != nil {
		log.Printf("Can't build getnewaddress request: %v\n", err)
		return ""
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send getnewaddress request: %v\n", err)
		return ""
	}

	defer resp.Body.Close()

	var json_resonse RPC_NewAddress_Response

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &json_resonse)
	if err != nil {
		log.Printf("Can't unmarshal getnewaddress response: %v\n", err)
		return ""
	}
	log.Println("Successfully created new address")

	return json_resonse.Result
}

func XTCGetAddress(pair string) string {

	switch pair {
	case DEROARRR, ARRRDERO, DEROXMR, XMRDERO:
		return ""
	}

	_, _, address, _ := XTCListReceivedByAddress(pair, "", 0, 0, true)

	return address
}

func XTCCheckBlockHeight(pair string) uint64 {

	req, err := XTCBuildRequest(pair, "getblockcount", []interface{}{})
	if err != nil {
		log.Printf("Can't build getblockcount request: %v\n", err)
		return 0
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send getblockcount request: %v\n", err)
		return 0
	}

	defer resp.Body.Close()

	var response RPC_GetBlockCount_Response

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Can't unmarshal getblockcount response: %v\n", err)
		return 0
	} else {
		return response.Result
	}
}

func XTCReceivedByAddress(pair string, wallet string) (float64, error) {

	req, err := XTCBuildRequest(pair, "getreceivedbyaddress", []interface{}{wallet, 2})
	if err != nil {
		log.Printf("Can't build getreceivedbyaddress request: %v\n", err)
		return 0, err
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send getreceivedbyaddress request: %v\n", err)
		return 0, err
	}

	defer resp.Body.Close()

	var response RPC_ReceivedByAddress_Response

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Can't unmarshal getreceivedbyaddress response: %v\n", err)
		return 0, err
	} else {
		return response.Result, nil
	}
}

func XTCListReceivedByAddress(pair string, wallet string, amount float64, height uint64, get_address bool) (bool, bool, string, error) {

	var method string
	var options []interface{}

	if !get_address {
		if pair == ARRRDERO {
			method = "zs_listreceivedbyaddress"
			options = append(options, wallet, 1, 3, height)
		} else {
			method = "listreceivedbyaddress"
			options = append(options, 1, false, false, wallet)
		}
	} else {
		method = "listreceivedbyaddress"
		options = append(options, 0, true)
	}

	req, err := XTCBuildRequest(pair, method, options)
	if err != nil {
		log.Printf("Can't build listreceivedbyaddress request: %v\n", err)
		return false, false, "", err
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send listreceivedbyaddress request: %v\n", err)
		return false, false, "", err
	}

	defer resp.Body.Close()

	if pair != ARRRDERO {
		var response RPC_ListReceivedByAddress_Response

		body, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(body, &response)
		if err != nil {
			log.Printf("Can't unmarshal listreceivedbyaddress response: %v\n", err)
			return false, false, "", err
		}

		if !get_address {
			if len(response.Result) > 0 {
				for _, tx := range response.Result[0].Txids {
					tx_data, err := XTCGetTransaction(pair, tx)
					if err != nil {
						log.Printf("Error checking TX: %v\n", err)
						continue
					}
					if tx_data.Result.Amount == amount && tx_data.Result.Blockheight >= height {
						if tx_data.Result.Confirmations > 1 {
							return true, true, "", nil
						} else {
							return false, true, "", nil
						}
					}
				}
			}
		} else {
			if len(response.Result) > 0 {
				return false, false, response.Result[0].Address, nil
			} else {
				return false, false, "", nil
			}
		}
	} else {
		var response RPC_ARRR_ListReceivedByAddress_Response

		body, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(body, &response)
		if err != nil {
			log.Printf("Can't unmarshal zs_listreceivedbyaddress response: %v\n", err)
			return false, false, "", err
		}

		if len(response.Result) > 0 {
			for e := range response.Result {
				if response.Result[e].BlockHeight >= height {
					for _, tx := range response.Result[e].Received {
						if tx.Value == amount {
							if response.Result[e].Confirmations > 1 {
								return true, true, "", nil
							} else {
								return false, true, "", nil
							}
						}
					}
				}
			}
		}
	}

	return false, false, "", nil
}

func XTCGetTransaction(pair string, txid string) (result RPC_GetTransaction_Response, err error) {

	req, err := XTCBuildRequest(pair, "gettransaction", []interface{}{txid, false})
	if err != nil {
		log.Printf("Can't build gettransaction request: %v\n", err)
		return result, err
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send gettransaction request: %v\n", err)
		return result, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Can't unmarshal gettransaction response: %v\n", err)
		return result, err
	}

	return result, nil
}

func XTCGetBalance(pair string) float64 {

	var method string
	var options []interface{}
	if pair == DEROARRR || pair == ARRRDERO {
		method = "z_getbalance"
		options = append(options, ARRR_GetAddress())
	} else {
		method = "getbalance"
	}

	req, err := XTCBuildRequest(pair, method, options)
	if err != nil {
		log.Printf("Can't build getbalance request: %v\n", err)
		return 0
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send getbalance request: %v\n", err)
		return 0
	}

	defer resp.Body.Close()

	var result RPC_GetBalance_Result

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Can't unmarshal getbalance response: %v\n", err)
		return 0
	}

	if strings.HasPrefix(pair, "dero") {
		result.Result -= Locked.GetLockedBalance(pair)
	}

	return RoundFloat(result.Result, 8)
}

func XTCSend(pair string, wallet string, amount float64, fee uint64) (bool, string) {

	var options []interface{}

	switch pair {
	case DEROLTC:
		options = append(options, wallet, amount)
	case DEROBTC:
		options = append(options, wallet, amount, "", "", false, true, nil, "unset", nil, fee)
	}

	req, err := XTCBuildRequest(pair, "sendtoaddress", options)
	if err != nil {
		log.Printf("Can't build sendtoaddress request: %v\n", err)
		return false, ""
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send sendtoaddress request: %v\n", err)
		return false, ""
	}

	defer resp.Body.Close()

	var result RPC_Send_Result

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Can't unmarshal sendtoaddress response: %v\n", err)
		return false, ""
	}

	return true, result.Result
}

func XTCValidateAddress(pair string, address string) bool {

	var method string
	if pair == DEROARRR || pair == ARRRDERO {
		method = "z_validateaddress"
	} else {
		method = "validateaddress"
	}

	req, err := XTCBuildRequest(pair, method, []interface{}{address})
	if err != nil {
		log.Printf("Can't build validateaddress request: %v\n", err)
		return false
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send validateaddress request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	var result RPC_Validate_Address_Result

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Can't unmarshal validateaddress response: %v\n", err)
		return false
	}

	return result.Result.IsValid
}

func XTC_GetURL(pair string) string {

	switch pair {
	case BTCDERO, DEROBTC:
		return XTC_URL[BTCDERO]
	case LTCDERO, DEROLTC:
		return XTC_URL[LTCDERO]
	case ARRRDERO, DEROARRR:
		return XTC_URL[ARRRDERO]
	default:
		return ""
	}
}

func ARRR_Send(wallet string, amount float64) (bool, string) {

	var params []RPC_ARRR_SendMany_Params

	params = append(params, RPC_ARRR_SendMany_Params{Address: wallet, Amount: amount})
	options := []interface{}{ARRR_GetAddress(), params}

	req, err := XTCBuildRequest(DEROARRR, "z_sendmany", options)
	if err != nil {
		log.Printf("Can't build z_sendmany request: %v\n", err)
		return false, ""
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send z_sendmany request: %v\n", err)
		return false, ""
	}

	defer resp.Body.Close()

	var result RPC_Send_Result

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Can't unmarshal z_sendmany response: %v\n", err)
		return false, ""
	}

	return true, result.Result
}

func ARRR_GetAddress() string {

	req, err := XTCBuildRequest(DEROARRR, "z_listaddresses", nil)
	if err != nil {
		log.Printf("Can't build z_listaddresses request: %v\n", err)
		return ""
	}

	resp, err := XTC_Daemon.Do(req)
	if err != nil {
		log.Printf("Can't send z_listaddresses request: %v\n", err)
		return ""
	}

	defer resp.Body.Close()

	var result RPC_ARRR_ListAddresses

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Can't unmarshal z_listaddresses response: %v\n", err)
		return ""
	}

	if len(result.Result) == 0 {
		return ""
	} else {
		return result.Result[0]
	}
}
