package kanban

import (
	"sort"
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

func NewTrelloBoard(client *trello.Client, boardID string) *TrelloBoard {
	return &TrelloBoard{
		client:  client,
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
		go b.trelloCardToCard(trelloCard, trelloColumns, cardChannel)
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

func (b *TrelloBoard) trelloCardToCard(card *trello.Card, columns []*trello.List, channel chan *cardFetchResult) {
	listChangeAction, err := card.GetListChangeActions()
	if err != nil {
		channel <- &cardFetchResult{
			card: nil,
			err:  err,
		}
		return
	}

	sortedActions := listChangeAction.FilterToListChangeActions()
	sort.Slice(sortedActions, func(i, j int) bool {
		return sortedActions[i].Date.Before(sortedActions[j].Date)
	})

	if len(sortedActions) == 0 {
		channel <- &cardFetchResult{
			card: &Card{
				Name:           card.Name,
				DurationInDays: 0,
			},
			err:  nil,
		}
		return
	}

	var firstEnteredReadyList time.Time
	var firstEnteredDoneList time.Time
	var firstEnteredInProgressList time.Time
	for _, action := range sortedActions {
		if trello.ListAfterAction(action) == nil {
			continue
		}

		if trello.ListAfterAction(action).ID == columns[2].ID { // READY
			firstEnteredReadyList = action.Date
		}

		if trello.ListAfterAction(action).ID == columns[3].ID { // IN PROGRESS
			firstEnteredInProgressList = action.Date
		}

		if trello.ListAfterAction(action).ID == columns[len(columns)-1].ID { // DONE
			firstEnteredDoneList = action.Date
		}
	}

	if firstEnteredReadyList.IsZero() { // handle cards that were created in the in progress list
		firstEnteredReadyList = firstEnteredInProgressList
	}

	if firstEnteredReadyList.IsZero() { // handle cards that were created down stream to in progress
		firstEnteredReadyList = sortedActions[0].Date
	}

	channel <- &cardFetchResult{
		card: &Card{
			Name:           card.Name,
			DurationInDays: int(firstEnteredDoneList.Sub(firstEnteredReadyList).Round(time.Hour*24).Hours() / 24),
		},
		err: nil,
	}
}

