package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &Zpool{}

type Zpool struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewZpool(addr string) *Zpool {
	return &Zpool{Address: addr}
}

func (p *Zpool) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://www.zpool.ca/api/wallet?address=%s", p.Address), &jsonPayload)
	if err != nil {
		return 0
	}
	vtc, ok := jsonPayload["unpaid"].(float64)
	if !ok {
		return 0
	}
	vtc *= 100000000
	return uint64(vtc)
}

func (p *Zpool) GetStratumUrl() string {
	return "stratum+tcp://verthash.na.mine.zpool.ca:6144"
}

func (p *Zpool) GetUsername() string {
	return p.Address
}

func (p *Zpool) GetPassword() string {
	return "c=VTC,zap=VTC"
}

func (p *Zpool) GetID() int {
	return 6
}

func (p *Zpool) GetName() string {
	return "Zpool.ca - North America"
}

func (p *Zpool) GetFee() float64 {
	return 0.50
}

func (p *Zpool) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://www.zpool.ca/wallet/%s", addr))
}
