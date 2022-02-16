package backend

import "net"

func Socket(f string) (*FakeTPM, error) {
	conn, err := net.Dial("unix", f)
	if err != nil {
		return nil, err
	}
	return &FakeTPM{
		ReadWriteCloser: conn,
	}, nil
}
