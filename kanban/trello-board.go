package kanban

import (
	"errors"
	"sort"

	"github.com/adlio/trello"
)

type TrelloBoard struct {
	client            *trello.Client
	trelloCardMetrics *TrelloCardMetrics
	readyColumnName   string
	boardID           string
}

type cardFetchResult struct {
	card *DoneCard
	err  error
}

func NewTrelloBoard(client *trello.Client, trelloCardMetrics *TrelloCardMetrics, readyColumnName string, boardID string) *TrelloBoard {
	return &TrelloBoard{
		client:            client,
		trelloCardMetrics: trelloCardMetrics,
		readyColumnName:   readyColumnName,
		boardID:           boardID,
	}
}

func (b *TrelloBoard) DoneCards() ([]*DoneCard, error) {
	trelloBoard, err := b.client.GetBoard(b.boardID, trello.Defaults())
	if err != nil {
		return nil, err
	}

	trelloColumns, err := trelloBoard.GetLists()
	if err != nil {
		return nil, err
	}

	trelloCards, err := trelloColumns[len(trelloColumns)-1].GetCards()
	if err != nil {
		return nil, err
	}

	cardChannel := make(chan *cardFetchResult)
	for _, trelloCard := range trelloCards {
		go func(trelloCard *trello.Card) {
			days, err := b.trelloCardMetrics.DurationInDays(trelloCard, trelloColumns)
			cardChannel <- &cardFetchResult{
				card: &DoneCard{
					Name:           trelloCard.Name,
					DurationInDays: days,
				},
				err: err,
			}
		}(trelloCard)
	}

	cards := []*DoneCard{}
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

func (b *TrelloBoard) ReadyCards() ([]*ReadyCard, error) {
	trelloBoard, err := b.client.GetBoard(b.boardID, trello.Defaults())
	if err != nil {
		return nil, err
	}

	trelloColumns, err := trelloBoard.GetLists()
	if err != nil {
		return nil, err
	}

	var readyColumn *trello.List
	for _, column := range trelloColumns {
		if column.Name == b.readyColumnName {
			readyColumn = column
		}
	}

	if readyColumn == nil {
		return nil, errors.New("ready column not found")
	}

	trelloCards, err := readyColumn.GetCards()
	if err != nil {
		return nil, err
	}

	readyCards := []*ReadyCard{}
	for _, card := range trelloCards {
		readyCards = append(readyCards, &ReadyCard{Name: card.Name})
	}

	return readyCards, nil
}
