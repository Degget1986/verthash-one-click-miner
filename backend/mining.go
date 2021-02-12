package backend

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/miners"
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

func (m *Backend) GetArgs() miners.BinaryArguments {
	tracking.Track(tracking.TrackingRequest{
		Category: "Mining",
		Action:   "Switch Pool",
		Name:     fmt.Sprintf("%v", m.pool.GetName()),
	})

    // TODO: Alert user through UI that payment dropdown
    // only applies when Zergpool is selected
    var password string
	if m.pool.GetID() == 5 {
	    // Use the payout dropdown option for Zergpool only
	    password = m.payout.GetPassword()
	} else {
	    password = m.pool.GetPassword()
	}

	return miners.BinaryArguments{
		StratumUrl:      m.pool.GetStratumUrl(),
		StratumUsername: m.pool.GetUsername(),
		StratumPassword: password,
	}
}

func (m *Backend) GetPoolFee() string {
	return fmt.Sprintf("%0.1f%%", m.pool.GetFee())
}

func (m *Backend) GetPoolName() string {
	return m.pool.GetName()
}

func (m *Backend) StartMining() bool {
	logging.Infof("Starting mining process...")

	tracking.Track(tracking.TrackingRequest{
		Category: "Mining",
		Action:   "Start",
	})

	args := m.GetArgs()

	startProcessMonitoring := make(chan bool)

	go func() {
		<-startProcessMonitoring
		continueLoop := true
		for continueLoop {
			newMinerBinaries := make([]*miners.BinaryRunner, 0)
			for _, br := range m.minerBinaries {
				if br.CheckRunning() == miners.RunningStateRapidFail {
					m.rapidFailures = append(m.rapidFailures, br)
					m.runtime.Events.Emit("minerRapidFail", br.MinerBinary.MainExecutableName)

				} else {
					newMinerBinaries = append(newMinerBinaries, br)
				}
			}

			m.minerBinaries = newMinerBinaries

			select {
			case <-m.stopMonitoring:
				continueLoop = false
			case <-time.After(time.Second):
			}
		}
		logging.Infof("Stopped monitoring thread")
	}()

	go func() {
		cycles := 0
		nhr := util.GetNetHash()
		bh := util.GetBlockHeight()
		continueLoop := true
		for continueLoop {
			cycles++
			if cycles > 150 {
				// Don't refresh this every time since we refresh it every second
				// and this pulls from Insight. Every 150s is fine (every block)
				nhr = util.GetNetHash()

				bh = util.GetBlockHeight()

				cycles = 0
			}
			hr := uint64(0)
			for _, br := range m.minerBinaries {
				hr += br.HashRate()
			}
			hashrate := float64(hr) / float64(1000)
			hashrateUnit := "kH/s"
			if hashrate > 1000 {
				hashrate /= 1000
				hashrateUnit = "MH/s"
			}
			if hashrate > 1000 {
				hashrate /= 1000
				hashrateUnit = "GH/s"
			}
			if hashrate > 1000 {
				hashrate /= 1000
				hashrateUnit = "TH/s"
			}
			m.runtime.Events.Emit("hashRate", fmt.Sprintf("%0.2f %s", hashrate, hashrateUnit))

			netHash := float64(nhr) / float64(1000000000)

			m.runtime.Events.Emit("networkHashRate", fmt.Sprintf("%0.2f %s", netHash, hashrateUnit))

			avgEarning := float64(hr) / float64(nhr) * float64(14400) // 14400 = Emission per day. Need to adjust for halving

			m.runtime.Events.Emit("avgEarnings", fmt.Sprintf("%0.2f VTC", avgEarning))

			m.runtime.Events.Emit("blockHeight", fmt.Sprintf("%d", bh))

			select {
			case <-m.stopHash:
				continueLoop = false
			case <-m.refreshHashChan:
			case <-time.After(time.Second):
			}
		}
	}()

	go func() {
		continueLoop := true
		for continueLoop {

			logging.Infof("Updating balance...")
			m.wal.Update()
			b, bi := m.wal.GetBalance()
			pb := m.pool.GetPendingPayout()
			m.runtime.Events.Emit("balance", fmt.Sprintf("%0.8f", float64(b)/float64(100000000)))
			m.runtime.Events.Emit("balanceImmature", fmt.Sprintf("%0.8f", float64(bi)/float64(100000000)))
			m.runtime.Events.Emit("balancePendingPool", fmt.Sprintf("%0.8f", float64(pb)/float64(100000000)))
			select {
			case <-m.stopBalance:
				continueLoop = false
			case <-m.refreshBalanceChan:
			case <-time.After(time.Minute * 5):
			}
		}
	}()

	go func() {
		continueLoop := true
		for continueLoop {

			runningProcesses := 0
			for _, br := range m.minerBinaries {
				if br.IsRunning() {
					runningProcesses++
				}
			}

			m.runtime.Events.Emit("runningMiners", runningProcesses)

			timeout := time.Second * 1
			if runningProcesses > 0 {
				timeout = time.Second * 10
			}
			select {
			case <-m.stopRunningState:
				continueLoop = false
			case <-m.refreshRunningState:
			case <-time.After(timeout):
			}
		}
	}()

	for _, br := range m.minerBinaries {
		err := br.Start(args)
		if err != nil {
			m.StopMining()
			logging.Errorf("Failure to start %s: %s\n", br.MinerBinary.MainExecutableName, err.Error())
			return false
		}
	}

	startProcessMonitoring <- true

	return true
}

func (m *Backend) RefreshBalance() {

	m.refreshBalanceChan <- true
}

func (m *Backend) RefreshHashrate() {

	m.refreshHashChan <- true
}

func (m *Backend) RefreshRunningState() {

	m.refreshRunningState <- true
}

func (m *Backend) StopMining() bool {
	tracking.Track(tracking.TrackingRequest{
		Category: "Mining",
		Action:   "Stop",
	})
	select {
	case m.stopMonitoring <- true:
	default:
	}
	logging.Infof("Stopping mining process...")
	for _, br := range m.minerBinaries {
		br.Stop()
	}
	select {
	case m.stopBalance <- true:
	default:
	}
	select {
	case m.stopHash <- true:
	default:
	}
	select {
	case m.stopRunningState <- true:
	default:
	}
	return true
}
