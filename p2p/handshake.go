package p2p

type HandshakeFunc func(any) error

func NopHAndshakeFunc(any) error { return nil }