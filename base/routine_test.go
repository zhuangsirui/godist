package base

import "testing"

func TestRoutine(t *testing.T) {
	c := make(chan []byte)
	r := Routine{
		Channel: c,
	}
	id := RoutineId(3312)
	r.SetId(id)
	if r.GetId() != id {
		t.Error("ID is wrong")
	}
	go r.Cast([]byte{'p', 'i', 'n', 'g'})
	<-c
}
