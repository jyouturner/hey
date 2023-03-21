package loadtesting

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"syscall"
	"time"
)

type LoadTester struct {
	ConcurrentUsers int
}

type Worker interface {
	Process(ctx context.Context, input []string) (*string, error)
}

type InputReader interface {
	Read() ([]string, error)
}

// Run is the main process to run the testing, through the given reader and worker
func (s *LoadTester) Run(reader InputReader, w Worker) (int, error) {

	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()
	//use buffered channels to unblock
	inputsChannels := make([]chan []string, s.ConcurrentUsers)
	successChannels := make(chan int, s.ConcurrentUsers)
	bufferSize := 100
	for i := 0; i < s.ConcurrentUsers; i++ {
		//start a go routine with a worker, to listen to the inputs channel
		inputs := make(chan []string, bufferSize)
		inputsChannels[i] = inputs
	}

	wg, ctxErrGrp := errgroup.WithContext(ctxWithCancel)
	// goroutine to check for signals to gracefully finish all functions. note it is not part of the errgroup therefore g.wait() will not wait for it.
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		select {
		case sig := <-signalChannel:
			fmt.Printf("Received cancel signal: %s\n", sig)
			cancel()
		case <-ctxErrGrp.Done():
		}
	}()
	//start the worker go routines
	s.StartWorkers(ctxErrGrp, wg, inputsChannels, successChannels, w)
	//start the dispatcher
	s.StartInputFanout(ctxErrGrp, wg, inputsChannels, reader)

	// Check whether any of the goroutines failed. Since g is accumulating the
	// errors, we don't need to send them (or check for them) in the individual
	// results sent on the channel.
	err := wg.Wait()
	//close the results channel
	//close(successChannels)
	//print results
	sum := 0
	for r := range successChannels {
		sum = sum + r
	}
	return sum, err

}

// StartWorkers starts goroutines to run the workers. Each worker routine get a input channel (of string value). After the worker is done
// results (the sum of success operations) of the specific worker will be sent to the results channel.
func (s *LoadTester) StartWorkers(ctx context.Context, wg *errgroup.Group, inputs []chan []string, successChannels chan int, w Worker) {
	fmt.Println("starting workers")
	workers := int32(s.ConcurrentUsers)
	for i := 0; i < s.ConcurrentUsers; i++ {
		id := i
		wg.Go(func() error {
			defer func() {
				// Last one out closes shop
				if atomic.AddInt32(&workers, -1) == 0 {
					close(successChannels)
					fmt.Println("worker", id, "closed successChannels")
				}
			}()
			fmt.Println("worker", id, "started")

			successes, err := s.RunWorker(ctx, inputs[id], w)

			successChannels <- successes
			if err != nil {
				fmt.Println("unfortunately, exited due to error", err)
				return fmt.Errorf("worker %v %v", id, err)
			} else {
				fmt.Println("worker", id, "is done", "success count", successes)
				return nil
			}

		})
	}
	fmt.Println("workers started")
}

// RunWorker will listen to the input channel and run the process function. It will loop forever and exit when below happens
// 1. context is canceled
// 2. error while processing, which will cause the context being cancelled since we are using the errgroup package.
// 3. input channel is closed
// it will return the total number of success operations and potentialy the error
func (s *LoadTester) RunWorker(ctx context.Context, inputs chan []string, w Worker) (int, error) {
	//successes := make([]*string, len(inputs))
	//errs := make([]error, len(inputs))
	successCount := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Println("received signal to exit")
			return successCount, nil
		case input, ok := <-inputs:
			if ok {
				_, err := w.Process(ctx, input)

				if err != nil {
					if errors.Is(err, context.Canceled) {
						//error due to cancelled request or context. don't want to log it
						return successCount, nil
					} else {
						// exit on any failure
						return successCount, err
					}
				} else {
					//success
					successCount = successCount + 1
				}
			} else {
				//sender closed the channel
				fmt.Println("there are no more values to receive and the channel is closed.")
				return successCount, nil
			}
		}

	}
}

func SleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		fmt.Println("sleep end due to context done")
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// StartInputFanout start a goroutine to read from the data source and send to the input channels for worker to process. This go routine will exit
// 1. when context is closed
// 2. when there is error. This will cause the context being cancelled since we are using the errgroup package.
// 3. when the reader read a NIL from the channel
func (s *LoadTester) StartInputFanout(ctx context.Context, wg *errgroup.Group, inputsChannels []chan []string, reader InputReader) {
	fmt.Println("starting to read input line by line and send to workers ...")
	wg.Go(func() error {
		defer func() {
			fmt.Println("close the inputs channels")
			for _, channel := range inputsChannels {
				close(channel)
			}

		}()
		i := 0
		for {
			line, err := reader.Read()
			if line == nil {
				//end
				return nil
			}
			if err == io.EOF {
				//end
				return nil
			}
			if err != nil {
				fmt.Println("found error", err)
				return fmt.Errorf("input data dispatcher failed %v", err)
			}
			select {
			case <-ctx.Done():
				fmt.Println("dispatcher received signal to exit")
				return nil
			case inputsChannels[i] <- line:
				//Sends to a buffered channel block only when the buffer is full. Receives block when the buffer is empty.
				i = i + 1
				if i >= len(inputsChannels) {
					i = 0
				}
			}
		}
	})

}
