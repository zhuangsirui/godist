package base

import "testing"

func TestGPMD(t *testing.T) {
	gpmd := GPMD{
		Host: "localhost",
		Port: 3312,
	}
	gpmd.Address()
}
