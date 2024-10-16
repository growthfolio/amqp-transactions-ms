package service

import (
	"testing"

	"github.com/felipemacedo1/transaction-producer-ms/internal/queue"
)

func TestCSVService_ProcessCSVFile(t *testing.T) {
	type fields struct {
		queue queue.Queue
	}
	type args struct {
		filePath string
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
			s := &CSVService{
				queue: tt.fields.queue,
			}
			if err := s.ProcessCSVFile(tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("CSVService.ProcessCSVFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
