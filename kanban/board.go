package kanban

type Board interface {
	DoneCards() ([]*DoneCard, error)
	ReadyCards() ([]*ReadyCard, error)
}

type DoneCard struct {
	Name           string
	DurationInDays int
}

type ReadyCard struct {
	Name string
}
