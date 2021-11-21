package kanban

import "time"

type Board interface {
	DoneCards() ([]*DoneCard, error)
	ReadyCards() ([]*ReadyCard, error)
}

type DoneCard struct {
	Name     string
	LeadTime int
	DoneAt   time.Time
}

type ReadyCard struct {
	Name string
}
