package kanban

type Board interface {
	DoneCards() ([]*Card, error)
}

type Card struct {
	Name           string
	DurationInDays int
}
