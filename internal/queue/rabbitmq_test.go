package queue

import (
	"reflect"
	"testing"
)

func TestNewRabbitMQ(t *testing.T) {
	type args struct {
		url       string
		queueName string
	}
	tests := []struct {
		name    string
		args    args
		want    *RabbitMQ
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRabbitMQ(tt.args.url, tt.args.queueName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRabbitMQ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRabbitMQ() = %v, want %v", got, tt.want)
			}
		})
	}
}
