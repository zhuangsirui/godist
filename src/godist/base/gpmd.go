package base

import(
	"fmt"
)

type GPMD struct {
	Host string
	Port uint16
}

func (g GPMD) Address() string {
	return fmt.Sprintf("%s:%d", g.Host, g.Port)
}
