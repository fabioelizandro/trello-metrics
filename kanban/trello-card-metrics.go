package kanban

import (
	"sort"
	"time"

	"github.com/adlio/trello"
)

type TrelloCardMetrics struct {
	readyColumnName string
}

func NewTrelloCardMetrics(readyColumnName string) *TrelloCardMetrics {
	return &TrelloCardMetrics{readyColumnName: readyColumnName}
}

func (d *TrelloCardMetrics) DurationInDays(actions trello.ActionCollection, columns []*trello.List) int {
	if len(actions) == 0 {
		return 0
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

	return int(firstEnteredDoneList.Sub(firstEnteredReadyList).Round(time.Hour*24).Hours() / 24)
}

func (d *TrelloCardMetrics) DoneAt(card *trello.Card, actions trello.ActionCollection) time.Time {
	if len(actions) == 0 {
		return card.CreatedAt()
	}

	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Date.After(actions[j].Date)
	})

	return actions[0].Date
}

func (d *TrelloCardMetrics) firstEnteredReadyList(readyColumnIndex int, columns []*trello.List, sortedActions trello.ActionCollection) time.Time {
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
