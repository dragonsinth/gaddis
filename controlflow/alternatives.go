package controlflow

type alternatives struct {
	anyReturn   bool
	anyContinue bool
}

func (a *alternatives) Add(flow Flow) {
	switch flow {
	case CONTINUE:
		a.anyContinue = true
	case MAYBE:
		a.anyContinue = true
		a.anyReturn = true
	case HALT:
		a.anyReturn = true
	default:
		panic("here")
	}
}

func (a *alternatives) Result() Flow {
	if !a.anyReturn {
		return CONTINUE
	} else if !a.anyContinue {
		return HALT
	} else {
		return MAYBE
	}
}
