package service

import ()

type SieveCalcRPC struct{}

type Request struct {
	Rng int
	calcType string
}

type Reply struct {
	Result []int
}

func (t *SieveCalcRPC) RpcBlockConcSieve(req Request, res *Reply) error {
	res.Result = SieveCalc{}.blockConcSieve(req.Rng)
	return nil
}
