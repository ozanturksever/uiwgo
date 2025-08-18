// examples/reactivity_demo/main.go
// Demonstration of the new SolidJS-inspired reactivity system

package main

import (
	"fmt"

	"app/golid"
)

func main() {
	fmt.Println("🚀 Golid Reactivity System Demo")
	fmt.Println("================================")

	// Demo 1: Basic Signal Usage
	fmt.Println("\n📡 Demo 1: Basic Signals")
	basicSignalDemo()

	// Demo 2: Effects and Automatic Dependency Tracking
	fmt.Println("\n⚡ Demo 2: Effects and Dependency Tracking")
	effectDemo()

	// Demo 3: Memoized Computations
	fmt.Println("\n🧠 Demo 3: Memoized Computations")
	memoDemo()

	// Demo 4: Batched Updates
	fmt.Println("\n📦 Demo 4: Batched Updates")
	batchDemo()

	// Demo 5: Owner Context and Cleanup
	fmt.Println("\n🏠 Demo 5: Owner Context and Cleanup")
	ownerDemo()

	// Demo 6: Complex Reactive Graph
	fmt.Println("\n🕸️  Demo 6: Complex Reactive Graph")
	complexDemo()

	fmt.Println("\n✅ All demos completed successfully!")
}

func basicSignalDemo() {
	// Create a signal with initial value
	count, setCount := golid.CreateSignal(0)

	fmt.Printf("Initial count: %d\n", count())

	// Update the signal
	setCount(5)
	golid.FlushScheduler()
	fmt.Printf("Updated count: %d\n", count())

	// Update with a function
	name, setName := golid.CreateSignal("World")
	fmt.Printf("Initial greeting: Hello, %s!\n", name())

	setName("Golid")
	golid.FlushScheduler()
	fmt.Printf("Updated greeting: Hello, %s!\n", name())
}

func effectDemo() {
	count, setCount := golid.CreateSignal(0)
	var effectRuns int

	// Create an effect that automatically tracks dependencies
	golid.CreateEffect(func() {
		currentCount := count()
		effectRuns++
		fmt.Printf("Effect #%d: Count is now %d\n", effectRuns, currentCount)
	}, nil)

	golid.FlushScheduler()

	// Update the signal - effect will run automatically
	setCount(10)
	golid.FlushScheduler()

	setCount(20)
	golid.FlushScheduler()

	fmt.Printf("Total effect runs: %d\n", effectRuns)
}

func memoDemo() {
	number, setNumber := golid.CreateSignal(5)
	var computations int

	// Create a memoized computation
	doubled := golid.CreateMemo(func() int {
		computations++
		val := number()
		fmt.Printf("Computing double of %d (computation #%d)\n", val, computations)
		return val * 2
	}, nil)

	golid.FlushScheduler()

	// Access memo multiple times - should only compute once
	fmt.Printf("First access: %d\n", doubled())
	fmt.Printf("Second access: %d\n", doubled())
	fmt.Printf("Third access: %d\n", doubled())

	// Update source signal - memo will recompute
	setNumber(10)
	golid.FlushScheduler()
	fmt.Printf("After update: %d\n", doubled())

	fmt.Printf("Total computations: %d\n", computations)
}

func batchDemo() {
	x, setX := golid.CreateSignal(1)
	y, setY := golid.CreateSignal(2)
	var effectRuns int

	// Effect that depends on both signals
	golid.CreateEffect(func() {
		sum := x() + y()
		effectRuns++
		fmt.Printf("Effect run #%d: %d + %d = %d\n", effectRuns, x(), y(), sum)
	}, nil)

	golid.FlushScheduler()

	fmt.Println("Updating signals individually:")
	setX(10)
	golid.FlushScheduler()
	setY(20)
	golid.FlushScheduler()

	fmt.Println("Updating signals in batch:")
	golid.Batch(func() {
		setX(100)
		setY(200)
	})
	golid.FlushScheduler()

	fmt.Printf("Total effect runs: %d (should be 4: initial + 2 individual + 1 batch)\n", effectRuns)
}

func ownerDemo() {
	var cleanupCalled bool

	// Create a root context
	result, cleanup := golid.CreateRoot(func() string {
		count, setCount := golid.CreateSignal(0)

		// Register cleanup function
		golid.OnCleanup(func() {
			cleanupCalled = true
			fmt.Println("🧹 Cleanup function called!")
		})

		// Create effect within this context
		golid.CreateEffect(func() {
			fmt.Printf("Count in root context: %d\n", count())
		}, nil)

		setCount(42)
		return "Root context result"
	})

	golid.FlushScheduler()
	fmt.Printf("Root function returned: %s\n", result)
	fmt.Printf("Cleanup called before cleanup(): %t\n", cleanupCalled)

	// Call cleanup - should dispose all resources
	cleanup()
	fmt.Printf("Cleanup called after cleanup(): %t\n", cleanupCalled)
}

func complexDemo() {
	// Create a complex reactive graph
	firstName, setFirstName := golid.CreateSignal("John")
	lastName, setLastName := golid.CreateSignal("Doe")

	// Derived signal: full name
	fullName := golid.CreateMemo(func() string {
		return firstName() + " " + lastName()
	}, nil)

	// Another derived signal: initials
	initials := golid.CreateMemo(func() string {
		first := firstName()
		last := lastName()
		if len(first) > 0 && len(last) > 0 {
			return string(first[0]) + "." + string(last[0]) + "."
		}
		return ""
	}, nil)

	// Effect that displays the information
	golid.CreateEffect(func() {
		fmt.Printf("👤 %s (%s)\n", fullName(), initials())
	}, nil)

	golid.FlushScheduler()

	// Update names
	fmt.Println("Updating first name:")
	setFirstName("Jane")
	golid.FlushScheduler()

	fmt.Println("Updating last name:")
	setLastName("Smith")
	golid.FlushScheduler()

	fmt.Println("Batch updating both names:")
	golid.Batch(func() {
		setFirstName("Alice")
		setLastName("Johnson")
	})
	golid.FlushScheduler()
}
