package agent

type Agent interface {
	Answer(query string) (string, error)
	SetOnReActInitCallback(callback func(string, []Action))
	SetOnReActCompleteCallback(callback func(string, []ActionObservation))
}
