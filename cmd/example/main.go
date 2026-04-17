package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/akhenakh/numbat-cgo"
)

func main() {
	// Initialize Context
	ctx := numbat.NewContext()
	defer ctx.Free()

	// Perform a Calculation and Extract Raw Value
	res, err := ctx.Interpret("120 km/h -> m/s")
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	fmt.Printf("String Output: %s\n", res.StringOutput)
	if res.IsQuantity {
		// Notice how we get exactly 33.333333333333336 as a raw float64!
		fmt.Printf("Raw Float64 Value: %f\n", res.Value)
		fmt.Printf("Unit: %s\n\n", res.Unit)
	}

	// Inject Go Variables Safely
	fmt.Println("Setting variable 'flight_time' from Go...")
	err = ctx.SetVariable("flight_time", 2.5, "hours")
	if err != nil {
		log.Fatalf("Error setting variable: %v\n", err)
	}

	res, err = ctx.Interpret("flight_time -> minutes")
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	fmt.Printf("Flight time string: %s\n", res.StringOutput)
	fmt.Printf("Flight time raw float: %f\n", res.Value)
	fmt.Printf("Flight time unit: %s\n\n", res.Unit)

	// Demonstrate Thread Safety
	fmt.Println("Running concurrent evaluations...")
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine evaluates an expression at the same time
			expr := fmt.Sprintf("%d meters * 2", id)
			r, e := ctx.Interpret(expr)

			if e != nil {
				log.Printf("Goroutine %d error: %v\n", id, e)
				return
			}
			fmt.Printf("Goroutine %d result: %s (Raw: %f, Unit: %s)\n", id, r.StringOutput, r.Value, r.Unit)
		}(i)
	}
	wg.Wait()
	fmt.Println("\nConcurrent evaluations finished successfully!")
}
