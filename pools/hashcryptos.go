package pools

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &HashCryptos{}

type HashCryptos struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewHashCryptos(addr string) *HashCryptos {
	return &HashCryptos{Address: addr}
}

func (p *HashCryptos) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://www.hashcryptos.com/api/wallet/?address=%s", p.Address), &jsonPayload)
	if err != nil {
		return 0
	}
	unpaid, ok := jsonPayload["unpaid"].(string)
	if !ok {
		return 0
	}
	vtc, _ := strconv.ParseFloat(unpaid, 64)
	vtc *= 100000000
	return uint64(vtc)
}

func (p *HashCryptos) GetStratumUrl() string {
	return "stratum+tcp://stratum3.hashcryptos.com:9991"
}

func (p *HashCryptos) GetUsername() string {
	return p.Address
}

func (p *HashCryptos) GetPassword() string {
	return "x"
}

func (p *HashCryptos) GetID() int {
	return 7
}

func (p *HashCryptos) GetName() string {
	return "HashCryptos.com"
}

func (p *HashCryptos) GetFee() float64 {
	return 0.00
}

func (p *HashCryptos) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://hashcryptos.com/?address=%s", addr))
}
