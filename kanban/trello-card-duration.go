package kanban

import (
	"sort"
	"time"

	"github.com/adlio/trello"
)

type TrelloCardDuration struct {
	cachedActions *TrelloCachedCardActions
}

func NewTrelloCardDuration(cachedActions *TrelloCachedCardActions) *TrelloCardDuration {
	return &TrelloCardDuration{cachedActions: cachedActions}
}

func (d *TrelloCardDuration) DurationInDays(card *trello.Card, columns []*trello.List) (int,error) {
	listChangeAction, err := d.cachedActions.Actions(card)
	if err != nil {
		return 0, err
	}

	sortedActions := listChangeAction.FilterToListChangeActions()
	sort.Slice(sortedActions, func(i, j int) bool {
		return sortedActions[i].Date.Before(sortedActions[j].Date)
	})

	if len(sortedActions) == 0 {
		return 0, nil
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

	return int(firstEnteredDoneList.Sub(firstEnteredReadyList).Round(time.Hour*24).Hours() / 24), nil
}


