package main

/*
#cgo CFLAGS: -fno-inline-functions

void foo() {
}

void emptyStack(int size, void* data) {
	int i;
	for(i = 0; i < size; i++) {
		// Grab function pointer.
		void* fptr = data+i;

		// Grab number of arguments.
		long nargs = (long)data+i+1;

		// Skip over the arguments (in real practice we would unpack the memory
		// when making the call).
		i += nargs;

		foo();
	}
}
*/
import "C"

import (
	"flag"
	"log"
	"time"
	"unsafe"
)

var (
	// The number of functions that can exist in a batch before it must be
	// executed in C-land.
	//
	// In actual practice there is no limit because any function that returns
	// something (either directly or via a return-value pointer argument) means
	// that the pending batch MUST be executed now as the callee expects the
	// return value to be made immedietly available.
	//
	// I chose 25 here because that is, in my experience, the average number of
	// OpenGL calls you can make before invoking a function that returns
	// something.
	BatchSize = flag.Int("bs", 25, "number of function calls per batch")

	// The number of arguments to (fakely) push onto the stack to measure the
	// overhead of pushing arguments onto the stack.
	//
	// Five was chosen as that is a good average for OpenGL functions.
	NArgs = flag.Int("args", 5, "number of fake arguments (to test stack overhead)")

	// 350k C calls is what we will measure the performance of. This is what I
	// would expect would be needed for an AAA game.
	//
	// The higher the number of calls, the more CGO overhead there is.
	NCalls = flag.Int("calls", 350000, "number of C function calls to perform")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	log.Println("Benchmarking:")
	log.Println("	Batch Size:", *BatchSize)
	log.Println("	Number Of Calls:", *NCalls)
	log.Println("	Number Of Args:", *NArgs)
	log.Println("")

	// Measure the time it takes to simply perform CGO calls directly.

	start := time.Now()
	for i := 0; i < *NCalls; i++ {
		C.foo()
	}
	cgoTime := time.Since(start)
	log.Println("CGO", cgoTime)

	// Measure the time it takes to perform CGO calls through batching.

	// Initialize the stack. In practice we just let the stack grow to any size
	// and we re-use the space (by slicing to zero).
	stack := make([]unsafe.Pointer, 0, (*BatchSize)*(*NArgs))

	// Reset the timer.
	start = time.Now()
	for i := 0; i < *NCalls; i++ {
		// Push fake function pointer onto the stack (would be used by a jump
		// switch in actual practice).
		stack = append(stack, unsafe.Pointer(uintptr(0)))

		// Push fake arguments onto the stack
		for a := 0; a < *NArgs; a++ {
			stack = append(stack, unsafe.Pointer(uintptr(i)))
		}

		// At every batchSize, we execute the pending batch using C.emptyStack.
		if (i % *BatchSize) == 0 {
			C.emptyStack(C.int(len(stack)), unsafe.Pointer(&stack[0]))

			// Slice the stack back to zero (we reuse the space later).
			stack = stack[:0]
		}
	}
	batchingTime := time.Since(start)
	log.Println("Batching", batchingTime)

	// Print GitHub syntax for data:
	log.Printf("%d | %d | %d | %v | %v", *BatchSize, *NCalls, *NArgs, cgoTime, batchingTime)
}
