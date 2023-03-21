package loadtesting

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type MockWorker struct {
	delayDuration time.Duration
}

func (s *MockWorker) Process(ctx context.Context, input []string) (*string, error) {
	//time.Sleep(s.delayDuration)
	id := uuid.New().String()
	return &id, nil
}

type MockFailureWorker struct {
}

func (s *MockFailureWorker) Process(ctx context.Context, input []string) (*string, error) {
	//randomly return stuff
	r := rand.Float64()
	id := uuid.New().String()

	if r < 0.01 {
		fmt.Println("make it fail!")
		return nil, fmt.Errorf("Failure!")
	} else {
		return &id, nil
	}
}

type MockFailureReader struct {
}

func (s *MockFailureReader) Read() ([]string, error) {
	//randomly return stuff
	r := rand.Float64()
	t := fmt.Sprintf("%f", r)
	if r < 0.01 {
		return []string{t}, fmt.Errorf("Bad")
	} else {

		return []string{t}, nil
	}
}

type MockReader struct {
	index int
	max   int
}

func (s *MockReader) Read() ([]string, error) {
	s.index = s.index + 1
	if s.index > s.max {
		// EOF
		return nil, nil
	}
	r := rand.Float64()
	t := fmt.Sprintf("%f", r)
	return []string{t}, nil
}

func TestLoadTester_Run_Mock(t *testing.T) {

	type fields struct {
		ConcurrentUsers int
	}
	type args struct {
		reader InputReader
		w      Worker
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "test process success",
			fields: fields{
				ConcurrentUsers: 5,
			},
			args: args{
				reader: &MockReader{
					max: 1000,
				},
				w: &MockWorker{},
			},
			want:    1000,
			wantErr: false,
		},

		{
			name: "test process failure",
			fields: fields{
				ConcurrentUsers: 5,
			},
			args: args{
				reader: &MockReader{
					max: 1000,
				},
				w: &MockFailureWorker{},
			},
			//want:    1,
			wantErr: true,
		},
		{
			name: "test read faiure",
			fields: fields{
				ConcurrentUsers: 5,
			},
			args: args{
				reader: &MockFailureReader{},
				w:      &MockWorker{},
			},
			//want:    1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LoadTester{
				ConcurrentUsers: tt.fields.ConcurrentUsers,
			}
			got, err := s.Run(tt.args.reader, tt.args.w)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTester.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want > 0 && got != tt.want {
				t.Errorf("LoadTester.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadTester_StartWorkers(t *testing.T) {
	type fields struct {
		ConcurrentUsers int
	}
	type args struct {
		ctx             context.Context
		wg              *errgroup.Group
		inputs          []chan []string
		successChannels chan int
		w               Worker
	}
	ctxWithCancel, _ := context.WithCancel(context.Background())
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test start workers success",
			fields: fields{
				ConcurrentUsers: 5,
			},
			args: args{
				ctx: ctxWithCancel,
				wg:  &errgroup.Group{},
				inputs: []chan []string{
					make(chan []string),
					make(chan []string),
					make(chan []string),
					make(chan []string),
					make(chan []string),
				},
				successChannels: make(chan int, 5),
				w:               &MockWorker{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := &LoadTester{
				ConcurrentUsers: tt.fields.ConcurrentUsers,
			}
			go func() {
				time.Sleep(2 * time.Second)
				//close the channels
				for i := 0; i < len(tt.args.inputs); i++ {
					fmt.Println("closing", i)
					close(tt.args.inputs[i])
				}
			}()
			s.StartWorkers(tt.args.ctx, tt.args.wg, tt.args.inputs, tt.args.successChannels, tt.args.w)
			tt.args.wg.Wait()

		})
	}
}

func TestLoadTester_RunWorker(t *testing.T) {
	type fields struct {
		ConcurrentUsers int
	}
	type args struct {
		ctx    context.Context
		inputs chan []string
		w      Worker
	}
	ctxWithCancel, _ := context.WithCancel(context.Background())
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "test start worker success",
			fields: fields{
				ConcurrentUsers: 5,
			},
			args: args{
				ctx:    ctxWithCancel,
				inputs: make(chan []string),
				w:      &MockWorker{},
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LoadTester{
				ConcurrentUsers: tt.fields.ConcurrentUsers,
			}
			go func() {
				//sleep for 2 seconds then close the input channel to terminate the worker routine
				time.Sleep(2 * time.Second)
				close(tt.args.inputs)
			}()
			got, err := s.RunWorker(tt.args.ctx, tt.args.inputs, tt.args.w)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTester.RunWorker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LoadTester.RunWorker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadTester_SleepWithContext(t *testing.T) {
	type fields struct {
		ConcurrentUsers int
	}
	type args struct {
		ctx context.Context
		d   time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SleepWithContext(tt.args.ctx, tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("LoadTester.SleepWithContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadTester_StartInputFanout(t *testing.T) {
	type fields struct {
		ConcurrentUsers int
	}
	type args struct {
		ctx            context.Context
		wg             *errgroup.Group
		inputsChannels []chan []string
		reader         InputReader
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LoadTester{
				ConcurrentUsers: tt.fields.ConcurrentUsers,
			}
			s.StartInputFanout(tt.args.ctx, tt.args.wg, tt.args.inputsChannels, tt.args.reader)
		})
	}
}
