package base

import(
)

type RoutineId uint64

type Routine struct {
	Id RoutineId
	Channel chan []byte
}

func (r *Routine) Cast(message []byte) {
	r.Channel <- message
}
