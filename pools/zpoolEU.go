package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &ZpoolEU{}

type ZpoolEU struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewZpoolEU(addr string) *ZpoolEU {
	return &ZpoolEU{Address: addr}
}

func (p *ZpoolEU) GetPendingPayout() uint64 {
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

func (p *ZpoolEU) GetStratumUrl() string {
	return "stratum+tcp://verthash.eu.mine.zpool.ca:6144"
}

func (p *ZpoolEU) GetUsername() string {
	return p.Address
}

func (p *ZpoolEU) GetPassword() string {
	return "c=VTC,zap=VTC"
}

func (p *ZpoolEU) GetID() int {
	return 7
}

func (p *ZpoolEU) GetName() string {
	return "Zpool.ca - Europe"
}

func (p *ZpoolEU) GetFee() float64 {
	return 0.50
}

func (p *ZpoolEU) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://www.zpool.ca/wallet/%s", addr))
}
