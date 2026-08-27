package rpc

type sessionState int

const (
	stateClosed = iota
	stateEstablished
)

type mosSession struct {
	state sessionState
}
