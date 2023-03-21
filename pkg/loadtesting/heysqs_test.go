package loadtesting

import (
	"testing"
)

func TestLoadTestFromCsvIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	type args struct {
		concurrentUsers int
		filename        string
		ignoreFirstLine bool
		queueName       string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "test with real csv file",
			args: args{
				concurrentUsers: 5,
				filename:        "mandrill100.csv",
				ignoreFirstLine: false,
				queueName:       "development-jerryyou-webhook-mandrill-app2",
			},
			want:    100,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromCsvToSQS(tt.args.concurrentUsers, tt.args.filename, tt.args.ignoreFirstLine, tt.args.queueName, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTestFromCsv() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("LoadTestFromCsv() = %v, want %v", got, tt.want)
			}
		})
	}
}
