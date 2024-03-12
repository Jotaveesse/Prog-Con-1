package service

import (
	"exercicio5/shared"
	"time"
)

type SieveCalcRPC struct{}

func (t *SieveCalcRPC) RpcBlockConcSieve(req shared.Request, res *shared.Reply) error {
	var startTime, endTime time.Time
	startTime = time.Now()
	res.Result = SieveCalc{}.blockConcSieve(req.Rng)
	endTime = time.Now()

	res.ProcessTime = endTime.Sub(startTime)
	return nil
}
