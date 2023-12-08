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

func markBlock(start int, end int) []int {
	//pula todos os pares, logo tamanho da array rpecisa ser so metade to range
	var composites = make([]bool, (end-start+1)/2)

	var endRoot = int(math.Sqrt(float64(end)))

	//marca multiplos de todos os numeros, não so dos primos, no final da no mesmo, mas seria melhor se fossem so primos
	//pula de 2 em 2 porque pares nao sao considerados
	for i := 3; i <= endRoot; i += 2 {

		//pula os multiplos de 3,5,7,11,13 pra ficar mais rapido
		if i >= 3*3 && i%3 == 0 {
			continue
		}
		if i >= 5*5 && i%5 == 0 {
			continue
		}
		if i >= 7*7 && i%7 == 0 {
			continue
		}
		if i >= 11*11 && i%11 == 0 {
			continue
		}
		if i >= 13*13 && i%13 == 0 {
			continue
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
		for j := firstComposite; j <= end; j += i * 2 {
			composites[(j-start)/2] = true
		}

	}

	var primes = make([]int, 0, 1)

	//extrai todos os primos do bit array
	for i := 0; i < (end-start+1)/2; i++ {
		if !composites[i] {
			primes = append(primes, (start+i)*2-1)
		}
	}

	return primes
}

func blockConcSieve(rng int) []int {
	primes := []int{2}
	var wg sync.WaitGroup
	var mutex sync.Mutex

	sliceSize := 128 * 1024 //128K * 8B (int tem 8 bytes) = 1MB por thread

	for end := 2; end <= rng; end += sliceSize {
		wg.Add(1)
		var start = end + sliceSize

		//criação das threads
		go func(start int, end int) {
			defer wg.Done()

			if end > rng {
				end = rng
			}

			slicePrimes := markBlock(start, end)

			//lock para escrita na array final
			mutex.Lock()
			//pode dar erro caso tenham muitos primos ( >64 milhoes ) pq pode ficar sem memoria
			primes = append(primes, slicePrimes...)
			mutex.Unlock()

		}(end, start)
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
