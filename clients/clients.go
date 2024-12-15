package clients

import (
	"log"
	"swap-client/coin"

	"github.com/lesismal/nbio/nbhttp/websocket"
)

func IsExternalSwapAvailable(pair string, amount float64) (ok bool, client *websocket.Conn) {

	Clients.Range(func(key any, value any) bool {
		c := value.(ClientInfo)
		for _, p := range c.PairInfo {
			if p.Pair == pair && p.Balance >= amount {
				ok = true
				client = key.(*websocket.Conn)
				return false
			}
		}
		return true
	})
	return
}

func PrepareExternalSwap(pair string, amount float64) (bool, *websocket.Conn) {

	// only XMR swaps
	if pair != coin.XMRDERO && pair != coin.DEROXMR {
		log.Println("Only 3rd party XMR swaps")
		return false, nil
	}

	ok, conn := IsExternalSwapAvailable(pair, amount)
	if !ok {
		log.Println("No 3rd party swaps available")
		return false, nil
	}

	return true, conn
}

func (c *SwapState) ChangeClientState(mode uint, conn *websocket.Conn) {

	c.Lock()
	defer c.Unlock()

	if mode == LOCK {
		c.Client[conn] = true
	} else {
		c.Client[conn] = false
	}
}

func (c *SwapState) CheckClientState(conn *websocket.Conn) bool {

	c.RLock()
	defer c.RUnlock()

	return c.Client[conn]
}

func (c *SwapState) AddOrigin(conn *websocket.Conn, target *websocket.Conn) {

	c.Lock()
	defer c.Unlock()

	c.Result[conn] = target
}

func (c *SwapState) GetOrigin(conn *websocket.Conn) *websocket.Conn {

	c.RLock()
	defer c.RUnlock()

	return c.Result[conn]
}
