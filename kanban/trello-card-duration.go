package kanban

import (
	"sort"
	"time"

	"github.com/adlio/trello"
)

type TrelloCardDuration struct {
	cachedActions   *TrelloCachedCardActions
	readyColumnName string
}

func NewTrelloCardDuration(cachedActions *TrelloCachedCardActions, readyColumnName string) *TrelloCardDuration {
	return &TrelloCardDuration{cachedActions: cachedActions, readyColumnName: readyColumnName}
}

func (d *TrelloCardDuration) DurationInDays(card *trello.Card, columns []*trello.List) (int, error) {
	actions, err := d.cachedActions.ListChangeActions(card)
	if err != nil {
		return 0, err
	}

	if len(actions) == 0 {
		return 0, nil
	}

	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Date.After(actions[j].Date)
	})

	readyColumnIndex := 0
	for index, column := range columns {
		if column.Name == d.readyColumnName {
			readyColumnIndex = index
		}
	}

	firstEnteredDoneList := actions[0].Date
	firstEnteredReadyList := d.firstEnteredReadyList(readyColumnIndex, columns, actions)

	return int(firstEnteredDoneList.Sub(firstEnteredReadyList).Round(time.Hour*24).Hours() / 24), nil
}

func (d *TrelloCardDuration) firstEnteredReadyList(readyColumnIndex int, columns []*trello.List, sortedActions trello.ActionCollection) time.Time {
	for _, action := range sortedActions {
		if trello.ListAfterAction(action) == nil {
			continue
		}

		if trello.ListAfterAction(action).ID == columns[readyColumnIndex].ID {
			return action.Date
		}
	}

	return d.firstEnteredReadyList(readyColumnIndex+1, columns, sortedActions)
}
