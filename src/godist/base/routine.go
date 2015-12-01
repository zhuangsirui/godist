package base

// 使用一个 uint64 保存每个 Goroutine 的 ID 。
type RoutineId uint64

// 持有每个 Routine 的信息。其中 Channel 字段必须不能是同步通道。否则 Cast 消息
// 会阻塞。
type Routine struct {
	id      RoutineId
	idLock  bool
	Channel chan []byte
}

// 设置 Goroutine 的 ID 。只能够被 godist 自己调用。如果调用了两次，则会抛出
// panic 。
func (r *Routine) SetId(id RoutineId) {
	if r.idLock {
		panic("godist.base: cannot set routine id twice!")
	}
	r.id = id
	r.idLock = true
}

func (r *Routine) GetId() RoutineId {
	if !r.idLock {
		panic("godist.base: id have not assigned")
	}
	return r.id
}

func (r *Routine) Cast(message []byte) {
	r.Channel <- message
}
