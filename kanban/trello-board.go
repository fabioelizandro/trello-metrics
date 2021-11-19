package kanban

import (
	"time"

	"github.com/adlio/trello"
)

type TrelloBoard struct {
	client *trello.Client
	boardID string
}

type cardFetchResult struct {
	card *Card
	err error
}

func NewTrelloBoard(key, token, boardID string) *TrelloBoard {
	return &TrelloBoard{
		client:  trello.NewClient(key, token),
		boardID: boardID,
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
		go b.trelloCardToCard(trelloCard, cardChannel)
	}

	cards := []*Card{}
	for range trelloCards {
		fetchResult := <-cardChannel
		if fetchResult.err != nil {
			return nil, fetchResult.err
		}

		cards = append(cards, fetchResult.card)
	}

	return cards, nil
}

func (b *TrelloBoard) trelloCardToCard(card *trello.Card, channel chan *cardFetchResult) {
	listDurations, err := card.GetListDurations()
	if err != nil {
		channel <- &cardFetchResult{
			card: nil,
			err:  err,
		}
		return
	}

	if len(listDurations) == 0 { // handle cards created in the done list
		channel <- &cardFetchResult{
			card: &Card{
				Name:           card.Name,
				DurationInDays: 0,
			},
			err: nil,
		}
		return
	}

	var firstEnteredReadyList time.Time
	var firstEnteredDoneList time.Time
	for _, listDuration := range listDurations {
		if listDuration.ListName == "Ready" {
			firstEnteredReadyList = listDuration.FirstEntered
		}

		if listDuration.ListName == "Done" {
			firstEnteredDoneList = listDuration.FirstEntered
		}
	}

	if firstEnteredReadyList.IsZero() { // handle cards that were created in a list down stream to Ready
		firstEnteredReadyList = listDurations[0].FirstEntered
	}

	channel <- &cardFetchResult{
		card: &Card{
			Name:           card.Name,
			DurationInDays: int(firstEnteredDoneList.Sub(firstEnteredReadyList).Round(time.Hour*24).Hours() / 24),
		},
		err: nil,
	}
}

