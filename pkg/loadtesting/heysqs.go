package loadtesting

import (
	"context"
	"fmt"
	"os"

	"github.com/jyouturner/hey/input"
)

// FromCsvToSQS runs the load testing by reading lines from CSV file, and send messages to AWS SQS
func FromCsvToSQS(ConcurrentUsers int, filename string, ignoreFirstLine bool, queueName string, dryRun bool) (int, error) {
	// open file
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	d := input.NewCsvInputReader(f, ignoreFirstLine)

	tester := LoadTester{
		ConcurrentUsers: ConcurrentUsers,
	}
	if dryRun {
		//return tester.Run(d, &loadtesting.MockWorker{
		//	delayDuration: 20 * time.Millisecond,
		//})
		return 0, fmt.Errorf("no imp")
	} else {
		return tester.Run(d, NewSqsWorker(context.TODO(), queueName))
	}
}

////type MockWorker struct {
//	delayDuration time.Duration
//}

//func (s *MockWorker) process(ctx context.Context, input *string) (*string, error) {
//	//time.Sleep(s.delayDuration)
//	id := uuid.New().String()
//	return &id, nil
//}
