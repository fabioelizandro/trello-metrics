package main

import (
	"fmt"
	"io/ioutil"
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
		kanban.NewTrelloCardMetrics(
			cachedActions,
			env.MustRead("TRELLO_READY_COLUMN"),
		),
		env.MustRead("TRELLO_READY_COLUMN"),
		env.MustRead("TRELLO_BOARD_ID"),
	)

	doneCards, err := board.DoneCards()
	if err != nil {
		panic(err)
	}

	readyCards, err := board.ReadyCards()
	if err != nil {
		panic(err)
	}

	printMonteCarloSimulation(readyCards)
	renderListWithPercentage(doneCards)
	renderHistogram(doneCards)
}

func printMonteCarloSimulation(readyCards []*kanban.ReadyCard) {
	for _, card := range readyCards {
		println(card.Name)
	}
}

func renderListWithPercentage(doneCards []*kanban.DoneCard) {
	total := 0
	list := ""
	for _, card := range doneCards {
		total++
		list += fmt.Sprintf("<li>%d - %s - %f%%</li>", card.DurationInDays, card.Name, (float64(total)/float64(len(doneCards)))*100)
	}

	err := ioutil.WriteFile(
		"done-cards-list-with-percentage.html",
		[]byte(fmt.Sprintf(
			`
				<html>
					<body>
						<ul>%s</ul>
					</body>
				</html>
			`,
			list,
		)),
		0666,
	)
	if err != nil {
		panic(err)
	}
}

func renderHistogram(doneCards []*kanban.DoneCard) {
	histogram := map[int]int{}
	for _, card := range doneCards {
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
	f, err := os.Create("done-cards-histogram.html")
	if err != nil {
		panic(err)
	}

	err = bar.Render(f)
	if err != nil {
		panic(err)
	}
}
