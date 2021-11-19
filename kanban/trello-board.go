package kanban

import (
	"sort"

	"github.com/adlio/trello"
)

type TrelloBoard struct {
	client             *trello.Client
	trelloCardDuration *TrelloCardDuration
	boardID            string
}

type cardFetchResult struct {
	card *Card
	err  error
}

func NewTrelloBoard(client *trello.Client, trelloCardDuration *TrelloCardDuration, boardID string) *TrelloBoard {
	return &TrelloBoard{
		client:             client,
		trelloCardDuration: trelloCardDuration,
		boardID:            boardID,
	}
}

func (b *TrelloBoard) DoneCards() ([]*Card, error) {
	trelloBoard, err := b.client.GetBoard(b.boardID, trello.Defaults())
	if err != nil {
		return nil, err
	}

	trelloColumns, err := trelloBoard.GetLists()
	if err != nil {
		return nil, err
	}

	trelloCards, err := trelloColumns[len(trelloColumns)-1].GetCards() // get all cards from the done list
	if err != nil {
		return nil, err
	}

	cardChannel := make(chan *cardFetchResult)
	for _, trelloCard := range trelloCards {
		go func(trelloCard *trello.Card) {
			days, err := b.trelloCardDuration.DurationInDays(trelloCard, trelloColumns)
			cardChannel <- &cardFetchResult{
				card: &Card{
					Name:           trelloCard.Name,
					DurationInDays: days,
				},
				err:  err,
			}
		}(trelloCard)
	}

	cards := []*Card{}
	for range trelloCards {
		fetchResult := <-cardChannel
		if fetchResult.err != nil {
			return nil, fetchResult.err
		}

		cards = append(cards, fetchResult.card)
	}

	sort.Slice(cards, func(i, j int) bool {
		return cards[i].DurationInDays < cards[j].DurationInDays
	})

	return cards, nil
}
