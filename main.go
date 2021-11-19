package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"trello-metrics/kanban"

	"github.com/fabioelizandro/goenv"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
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

	sort.Slice(cards, func(i, j int) bool {
		return cards[i].DurationInDays < cards[j].DurationInDays
	})

	total := 0
	for _, card := range cards {
		total++
		fmt.Printf("%d - %s - %f%%\n", card.DurationInDays, card.Name, (float64(total)/float64(len(cards)))*100)
	}

	histogram := map[int]int{}
	for _, card := range cards {
		if card.DurationInDays >= 0 {
			histogram[card.DurationInDays]++
		}
	}
	xAxisInt := []int{}
	for durationInDays := range histogram {
		xAxisInt = append(xAxisInt, durationInDays)
	}
	sort.Ints(xAxisInt)

	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Histogram",
		Subtitle: "Shows the duration distribution of done cards",
	}))

	xAxis := []string{}
	for _, i := range xAxisInt {
		xAxis = append(xAxis, strconv.Itoa(i))
	}

	bar.SetXAxis(xAxis)
	yAxis := []opts.BarData{}
	for _, i := range xAxisInt {
		yAxis = append(yAxis, opts.BarData{Value: histogram[i]})
	}
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
