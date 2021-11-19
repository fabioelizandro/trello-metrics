package main

import (
	"fmt"
	"trello-metrics/kanban"

	"github.com/fabioelizandro/goenv"
)

func main() {
	env := goenv.NewEnv(goenv.MustParseDotfileFromFilepath(".env"))

	board, err := kanban.CreateCachedBoard(
		kanban.NewTrelloBoard(
			env.MustRead("TRELLO_API_KEY"),
			env.MustRead("TRELLO_USER_TOKEN"),
			env.MustRead("TRELLO_BOARD_ID"),
		),
		"trello-metrics",
	)
	if err != nil {
		panic(err)
	}

	cards, err := board.DoneCards()
	if err != nil {
		panic(err)
	}

	for _, card := range cards {
		fmt.Printf("The card %s took %d days to get done\n", card.Name, card.DurationInDays)
	}
}
