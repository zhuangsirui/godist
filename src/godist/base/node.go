package base

import (
	"fmt"
)

type Node struct {
	Port uint16
	Host string
	Name string
}

func (n *Node) FullName() string {
	return fmt.Sprintf("%s@%s", n.Name, n.Host)
}
