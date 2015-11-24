package godist

import(
)

type routineId uint64

type routine struct {
	id routineId
	channel chan []byte
}

func (r *routine) cast(message []byte) {
	r.channel <- message
}
