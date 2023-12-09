package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

//----------------SEQUENCIAL----------------

func sieve(rng int) []int {
	rng++
	var composites = make([]bool, rng)

	var rngRoot = int(math.Sqrt(float64(rng)))

	//marca todos os compostos
	for i := 2; i <= rngRoot; i++ {
		if !composites[i] {
			for j := int(math.Pow(float64(i), 2)); j < rng; j += i {
				composites[j] = true
			}
		}
	}

	//extrai os primos do bit array
	var primes = make([]int, 0, 1)
	for i := 2; i < rng; i++ {
		if !composites[i] {
			primes = append(primes, i)
		}
	}

	return primes
}

//----------------CONCORRENTE----------------

func markDivided(idx int, rng_end int, composites []bool, wg *sync.WaitGroup) {
	defer wg.Done()

	if !composites[idx] {
		for j := idx * 2; j < rng_end; j += idx {
			composites[j] = true
		}
	}
}

func concSieve(rng int) []int {
	rng++
	var composites = make([]bool, rng)
	var wg sync.WaitGroup

	//cada thread marca multiplos de um numero
	var rngRoot = int(math.Sqrt(float64(rng)))
	for i := 2; i <= rngRoot; i++ {
		wg.Add(1)
		go markDivided(i, rng, composites, &wg)
	}
	wg.Wait()

	var primes = make([]int, 0, 1)

	//extrai primos do bit array
	for i := 2; i < rng; i++ {
		if !composites[i] {
			primes = append(primes, i)
		}
	}

	return primes
}

//----------------CONCORRENTE MELHORADO----------------

func markBlock(start int, end int, primes *[]int, wg *sync.WaitGroup, mutex *sync.Mutex, firstPrimes *[]int) []int {
	defer wg.Done()

	rng := end - start + 1
	endRoot := int(math.Sqrt(float64(end)))

	//pula todos os pares, logo tamanho da array precisa ser so metade to range
	var composites = make([]bool, rng/2)

	for _, i := range *firstPrimes {
		//todos os compostos  maiores q a raiz quadrada do limite ja estarao marcados
		if i > endRoot {
			break
		}

		//acha primeiro multiplo de i maior que start
		firstComposite := ((start + i - 1) / i) * i

		//se i^2 é maior q o primeiro multiplo é melhor so começar de i^2
		sqrdI := i * i
		if firstComposite < sqrdI {
			firstComposite = sqrdI
		}

		//se primeiro multiplo é par pega o proximo multiplo, que será impar
		if (firstComposite & 1) == 0 {
			firstComposite += i
		}

		//marca todos os multiplos de i dentro do intervalo
		doubleI := i * 2
		for j := firstComposite; j <= end; j += doubleI {
			composites[(j-start)/2] = true
		}

	}

	//extrai todos os primos do bit array
	var slicePrimes = make([]int, 0, 100)

	//extrai todos os primos do bit array
	for i := 0; i < rng/2; i++ {
		if !composites[i] {
			prime := (start+i)*2 - 1
			slicePrimes = append(slicePrimes, prime)
		}
	}

	mutex.Lock()
	*primes = append(*primes, slicePrimes...)
	mutex.Unlock()

	return *primes
}

func blockConcSieve(rng int) []int {
	primes := []int{2}
	var wg sync.WaitGroup
	var mutex sync.Mutex

	rngRoot := int(math.Sqrt(float64(rng)))

	var firstPrimes []int
	//calcula todos os primos que terão seus multiplos marcados inicialmente
	if rngRoot < 10000 {
		firstPrimes = sieve(rngRoot)
		firstPrimes = firstPrimes[1:]
	} else {
		firstPrimes = blockConcSieve(rngRoot)
		firstPrimes = firstPrimes[1:]
	}

	sliceSize := 128 * 1024 //128K * 8B (int tem 8 bytes) = 1MB por thread

	for start := 2; start <= rng; start += sliceSize {
		var end = start + sliceSize

		if end > rng {
			end = rng
		}

		wg.Add(1)
		go markBlock(start, end, &primes, &wg, &mutex, &firstPrimes)
	}

	wg.Wait()

	return primes
}

func mains() { //trocar nome pra main se quiser executar individualmente
	var primeRange int
	var printPrimes string
	var doConc string
	var doImprov string

	//runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Print("Choose range\n")
	fmt.Scan(&primeRange)

	fmt.Print("Run concurrently? (y or n)\n")
	fmt.Scan(&doConc)

	var startTime time.Time

	var primes []int
	if doConc == "y" {
		fmt.Print("Run improved version? (y or n)\n")
		fmt.Scan(&doImprov)

		startTime = time.Now()
		if doImprov == "y" {
			primes = blockConcSieve(primeRange)
		} else {
			primes = concSieve(primeRange)
		}
	} else {
		startTime = time.Now()
		primes = sieve(primeRange)
	}

	endTime := time.Now()

	fmt.Print("Found ", len(primes), " primes in ", endTime.Sub(startTime), "\n")

	fmt.Print("Print primes? (y or n)\n")
	fmt.Scan(&printPrimes)

	if printPrimes == "y" {
		fmt.Print(primes)
	}
}
