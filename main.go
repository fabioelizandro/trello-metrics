package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"trello-metrics/kanban"

	"github.com/adlio/trello"
	"github.com/fabioelizandro/goenv"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func main() {
	env := goenv.NewEnv(goenv.MustParseDotfileFromFilepath(".env"))

	cachedActions, err := kanban.CreateTrelloCachedCardActions("trello-metrics")
	if err != nil {
		panic(err)
	}

	board := kanban.NewTrelloBoard(
		trello.NewClient(
			env.MustRead("TRELLO_API_KEY"),
			env.MustRead("TRELLO_USER_TOKEN"),
		),
		kanban.NewTrelloCardDuration(cachedActions),
		env.MustRead("TRELLO_BOARD_ID"),
	)

	cards, err := board.DoneCards()
	if err != nil {
		panic(err)
	}

	printListWithPercentage(cards)
	renderHistogram(cards)
}

func printListWithPercentage(cards []*kanban.Card) {
	total := 0
	for _, card := range cards {
		total++
		fmt.Printf("%d - %s - %f%%\n", card.DurationInDays, card.Name, (float64(total)/float64(len(cards)))*100)
	}
}

func renderHistogram(cards []*kanban.Card) {
	histogram := map[int]int{}
	for _, card := range cards {
		histogram[card.DurationInDays]++
	}

	xAxisInt := []int{}
	for durationInDays := range histogram {
		xAxisInt = append(xAxisInt, durationInDays)
	}
	sort.Ints(xAxisInt)
	xAxis := []string{}
	for _, i := range xAxisInt {
		xAxis = append(xAxis, strconv.Itoa(i))
	}
	yAxis := []opts.BarData{}
	for _, i := range xAxisInt {
		yAxis = append(yAxis, opts.BarData{Value: histogram[i]})
	}

	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Histogram",
		Subtitle: "Shows the duration distribution of done cards",
	}))
	bar.SetXAxis(xAxis)
	bar.AddSeries("Done", yAxis)
	f, err := os.Create("bar.html")
	if err != nil {
		panic(err)
	}

	err = bar.Render(f)
	if err != nil {
		panic(err)
	}
}
