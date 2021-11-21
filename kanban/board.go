package kanban

type Board interface {
	DoneCards() ([]*DoneCard, error)
}

type DoneCard struct {
	Name           string
	DurationInDays int
}
