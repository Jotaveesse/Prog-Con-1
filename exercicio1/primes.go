package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

func mark_divided(idx int, rng_end int, not_primes []bool, wg *sync.WaitGroup) {
	defer wg.Done()

	if !not_primes[idx] {
		for j := idx * 2; j < rng_end; j += idx {
			not_primes[j] = true
		}
	}
}

func conc_sieve(rng int) []int {
	rng++
	var not_primes = make([]bool, rng)
	var wg sync.WaitGroup

	/*
		0 -> is prime
		1 -> not prime
	*/

	var rng_root = int(math.Sqrt(float64(rng)))
	for i := 2; i <= rng_root; i++ {
		wg.Add(1)
		go mark_divided(i, rng, not_primes, &wg)
	}
	wg.Wait()

	var primes = make([]int, 0, 1)

	for i := 2; i < rng; i++ {
		if !not_primes[i] {
			primes = append(primes, i)
		}
	}

	return primes
}

func sieve(rng int) []int {
	rng++
	var not_primes = make([]bool, rng)

	var rng_root = int(math.Sqrt(float64(rng)))

	for i := 2; i <= rng_root; i++ {
		if !not_primes[i] {
			for j := int(math.Pow(float64(i), 2)); j < rng; j += i {
				not_primes[j] = true
			}
		}
	}

	var primes = make([]int, 0, 1)

	for i := 2; i < rng; i++ {
		if !not_primes[i] {
			primes = append(primes, i)
		}
	}

	return primes
}

func exec() { //trocar nome pra main se quiser executar individualmente
	var prime_range int
	var print_primes string
	var do_conc string

	fmt.Print("Choose range\n")
	fmt.Scan(&prime_range)

	fmt.Print("Run concurrently? (y or n)\n")
	fmt.Scan(&do_conc)

	start_time := time.Now()

	var primes []int
	if do_conc == "y" {
		primes = conc_sieve(prime_range)
	} else {
		primes = sieve(prime_range)
	}

	end_time := time.Now()

	fmt.Print("Found ", len(primes), " primes in ", end_time.Sub(start_time), "\n")

	fmt.Print("Print primes? (y or n)\n")
	fmt.Scan(&print_primes)

	if print_primes == "y" {
		fmt.Print(primes)
	}
}
