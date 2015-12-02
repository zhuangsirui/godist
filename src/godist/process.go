package godist

import "godist/base"

type Process struct {
	Channel chan []byte
	routine *base.Routine
}

func (agent *Agent) NewProcess() *Process {
	c := make(chan []byte, 10)
	routine := &base.Routine{
		Channel: c,
	}
	agent.RegisterRoutine(routine)
	return &Process{
		Channel: c,
		routine: routine,
	}
}

func NewProcess() *Process {
	c := make(chan []byte, 10)
	routine := &base.Routine{
		Channel: c,
	}
	RegisterRoutine(routine)
	return &Process{
		Channel: c,
		routine: routine,
	}
}

func (p *Process) GetId() base.RoutineId {
	return p.routine.GetId()
}

func (p *Process) Run(handler func([]byte)) {
	go p.run(handler)
}

func (p *Process) run(handler func([]byte)) {
	defer func() {
		err := recover()
		if err != nil {
			p.run(handler)
		}
	}()
	for {
		handler(<-p.Channel)
	}
}
