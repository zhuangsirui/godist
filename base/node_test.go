package base

import "testing"

func TestNode(t *testing.T) {
	n := Node{
		Port: 3312,
		Host: "localhost",
		Name: "master_01",
	}
	n.FullName()
}
