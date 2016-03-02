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

func (n *Node) String() string {
	return fmt.Sprintf("%s@%s:%d", n.Name, n.Host, n.Port)
}
