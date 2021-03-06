package trellometrics

import (
	"errors"
	"sort"
	"time"

	"github.com/adlio/trello"
)

type Board struct {
	client            *trello.Client
	trelloCardMetrics *CardMetrics
	cachedActions     *CachedCardActions
	readyColumnName   string
	boardID           string
}

type DoneCard struct {
	Name     string
	LeadTime int
	DoneAt   time.Time
}

type ReadyCard struct {
	Name string
}

type cardFetchResult struct {
	card *DoneCard
	err  error
}

func NewBoard(client *trello.Client, cardMetrics *CardMetrics, cachedActions *CachedCardActions, readyColumnName string, boardID string) *Board {
	return &Board{
		client:            client,
		trelloCardMetrics: cardMetrics,
		cachedActions:     cachedActions,
		readyColumnName:   readyColumnName,
		boardID:           boardID,
	}
}

func (b *Board) DoneCards() ([]*DoneCard, error) {
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
			actions, err := b.cachedActions.ListChangeActions(trelloCard)
			if err != nil {
				cardChannel <- &cardFetchResult{
					card: nil,
					err:  err,
				}
				return
			}

			cardChannel <- &cardFetchResult{
				card: &DoneCard{
					Name:     trelloCard.Name,
					LeadTime: b.trelloCardMetrics.LeadTime(actions, trelloColumns),
					DoneAt:   b.trelloCardMetrics.DoneAt(trelloCard, actions),
				},
				err: nil,
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
		return cards[i].LeadTime < cards[j].LeadTime
	})

	return cards, nil
}

func (b *Board) ReadyCards() ([]*ReadyCard, error) {
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
