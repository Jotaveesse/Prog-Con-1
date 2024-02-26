package service

import (
	"exercicio4/shared"
)

type SieveCalcRPC struct{}

func (t *SieveCalcRPC) RpcBlockConcSieve(req shared.Request, res *shared.Reply) error {
	res.Result = SieveCalc{}.blockConcSieve(req.Rng)
	return nil
}
