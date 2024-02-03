package shared

const SievePort = 4040

type Request struct {
	Type string
	Rng  int
}

type Reply struct {
	Result []int
}
