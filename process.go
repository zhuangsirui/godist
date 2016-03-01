package godist

import (
	"godist/base"
	"log"
	"runtime/debug"
)

type Process struct {
	Channel chan []byte
	routine *base.Routine
}

func (agent *Agent) NewProcess() *Process {
	c := make(chan []byte, 100)
	routine := &base.Routine{
		Channel: c,
	}
	agent.RegisterRoutine(routine)
	return &Process{
		Channel: c,
		routine: routine,
	}
}

func (p *Process) GetId() base.RoutineId {
	return p.routine.GetId()
}

func (p *Process) Run(handler func([]byte) error) {
	p.run(handler)
}

func (p *Process) run(handler func([]byte) error) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("godist: process restart for reason: %s\n%s", err, debug.Stack())
			p.run(handler)
		}
	}()
	for {
		if err := handler(<-p.Channel); err != nil {
			log.Printf("godist.process: Process %d exit. reason: %s", p.GetId(), err)
			break
		}
	}
}
