package queue

type Queue interface {
	Publish(message []byte) error
}
