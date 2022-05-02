package server

type IChecker interface {
	Check() error
}
