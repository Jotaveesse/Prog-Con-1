package main

import (
	"exercicio3/client"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	chart "github.com/wcharczuk/go-chart/v2"
)

type TestResult struct {
	Name    string
	Ns      []int
	Results []float64
}

func joinTests(tests1 []TestResult, tests2 []TestResult) []TestResult {
	var joinedTests []TestResult

	for _, test1 := range tests1 {
		for _, test2 := range tests2 {
			if test1.Name == test2.Name {
				var joinedTest TestResult
				joinedTest.Name = test1.Name

				joinedTest.Ns = append(test1.Ns, test2.Ns...)
				joinedTest.Results = append(test1.Results, test2.Results...)

				joinedTests = append(joinedTests, joinedTest)
			}
		}
	}

	return joinedTests
}

func main() {
	names := []string{"UDP", "TCP"}
	functions := []func(int, string) ([]int, time.Duration){client.SieveClientUDP, client.SieveClientTCP}
	testNs1 := []int{100, 100, 100}
	testNs2 := []int{100, 100, 100}

	tests1 := runTests(names, functions, testNs1)
	tests2 := runTests(names, functions, testNs2)
	testsJoined := joinTests(tests1, tests2)

	makeBarChart(tests1, "comp-tempo")
	makeBarChart(tests2, "comp-tempo2")
	makeDiffLineGraph(testsJoined, "diff")
	makeDiffPercLineGraph(testsJoined, "diff-perc")
}

func runTests(names []string, testFuncs []func(int, string) ([]int, time.Duration), nArr []int) []TestResult {
	avrgLoops := 1000

	var results []TestResult

	//para cada função
	for i, function := range testFuncs {
		var res TestResult
		res.Name = names[i]

		//para cada valor de N
		for _, n := range nArr {
			totalRtt := 0

			for k := 0; k < avrgLoops; k++ {
				_, rtt := function(n, "blk_conc")

				totalRtt += int(rtt.Microseconds())
			}

			avrgMicro := float64(totalRtt) / float64(avrgLoops)

			res.Ns = append(res.Ns, n)
			res.Results = append(res.Results, avrgMicro)
		}

		results = append(results, res)
	}

	return results
}

func makeBarChart(tests []TestResult, outputFile string) {
	var yTicks []chart.Tick
	maxDiff := math.Inf(-1)

	var bars []chart.Value

	for i := 0; i < len(tests[0].Results); i++ {

		for j, test := range tests {
			sty := chart.Style{
				FillColor:   chart.DefaultColors[j],
				StrokeColor: chart.DefaultColors[j],
				StrokeWidth: 0,
			}

			result := test.Results[i]

			barValue := chart.Value{Value: result, Label: fmt.Sprintf("%s", addSeparator(test.Ns[i], ".")), Style: sty}
			bars = append(bars, barValue)

			if result > maxDiff {
				maxDiff = result
			}
		}
	}

	title := "Média de Tempo"
	for _, test := range tests {
		title += " " + test.Name + ","
	}

	title = replaceLastOccurrence(title, ",", "")
	title = replaceLastOccurrence(title, ",", " e")

	for i := int64(0); i <= 10; i++ {
		val := float64(int(i * int64(roundToNextThousand(int(maxDiff))/10)))
		yTicks = append(yTicks, chart.Tick{Value: val, Label: fmt.Sprintf("%.0f", val)})
	}

	graph := chart.BarChart{
		Title:      fmt.Sprintf(title),
		TitleStyle: chart.Style{FontSize: 14},

		YAxis: chart.YAxis{
			Name:  "Microsegundos",
			Ticks: yTicks,
		},

		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Height:   512,
		BarWidth: 60,
		Bars:     bars,
	}

	file, _ := os.Create("graphs/" + outputFile + ".png")
	defer file.Close()
	graph.Render(chart.PNG, file)
}

func makeDiffPercLineGraph(tests []TestResult, outputFile string) {
	var xTicks []chart.Tick
	var yTicks []chart.Tick
	var xValues []float64

	for i := range tests[0].Results {
		xValues = append(xValues, float64(i))
	}

	for _, test := range tests {
		for j, n := range test.Ns {
			xTicks = append(xTicks, chart.Tick{Value: float64(j), Label: fmt.Sprintf("%s", addSeparator(n, "."))})
		}
	}

	ratios := ratioArrays(tests[1].Results, tests[0].Results)

	maxDiff := math.Inf(-1)
	minDiff := math.Inf(1)

	for _, diff := range ratios {
		if diff > maxDiff {
			maxDiff = diff
		}
		if diff < minDiff {
			minDiff = diff
		}
	}

	for i := 0; i <= 10; i++ {
		val := minDiff + float64(i)*((maxDiff-minDiff)/10)
		yTicks = append(yTicks, chart.Tick{Value: val, Label: fmt.Sprintf("%.2f%%", val)})
	}

	title := "Diferença Relativa Entre"
	for _, test := range tests {
		title += " " + test.Name + ","
	}

	title = replaceLastOccurrence(title, ",", "")
	title = replaceLastOccurrence(title, ",", " e")

	graph := chart.Chart{
		Title:      title,
		TitleStyle: chart.Style{FontSize: 14},
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		XAxis: chart.XAxis{
			Name:  "N",
			Ticks: xTicks,
		},
		YAxis: chart.YAxis{
			Name:  "Porcentagem",
			Ticks: yTicks,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: chart.DefaultColors[0],
					StrokeWidth: 2,
				},
				XValues: xValues,
				YValues: ratios,
			},
		},
	}

	file, _ := os.Create("graphs/" + outputFile + ".png")
	defer file.Close()
	graph.Render(chart.PNG, file)
}

func makeDiffLineGraph(tests []TestResult, outputFile string) {
	var xTicks []chart.Tick
	var yTicks []chart.Tick
	var xValues []float64

	for i := range tests[0].Results {
		xValues = append(xValues, float64(i))
	}

	for _, test := range tests {
		for j, n := range test.Ns {
			xTicks = append(xTicks, chart.Tick{Value: float64(j), Label: fmt.Sprintf("%s", addSeparator(n, "."))})
		}
	}

	ratios := subtractArrays(tests[0].Results, tests[1].Results)

	maxDiff := math.Inf(-1)
	minDiff := math.Inf(1)

	for _, diff := range ratios {
		if diff > maxDiff {
			maxDiff = diff
		}
		if diff < minDiff {
			minDiff = diff
		}
	}

	for i := 0; i <= 10; i++ {
		val := float64(i) * (float64(roundToNextThousand(int(maxDiff-minDiff))) / 10)
		yTicks = append(yTicks, chart.Tick{Value: val, Label: fmt.Sprintf("%.0f", val)})
	}

	title := "Diferença Entre"
	for _, test := range tests {
		title += " " + test.Name + ","
	}

	title = replaceLastOccurrence(title, ",", "")
	title = replaceLastOccurrence(title, ",", " e")

	graph := chart.Chart{
		Title:      title,
		TitleStyle: chart.Style{FontSize: 14},
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		XAxis: chart.XAxis{
			Name:  "N",
			Ticks: xTicks,
		},
		YAxis: chart.YAxis{
			Name:  "Microsegundos",
			Ticks: yTicks,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: chart.DefaultColors[0],
					StrokeWidth: 2,
				},
				XValues: xValues,
				YValues: ratios,
			},
		},
	}

	file, _ := os.Create("graphs/" + outputFile + ".png")
	defer file.Close()
	graph.Render(chart.PNG, file)
}

func roundToNextThousand(number int) int {
	rounded := (number + 999) / 1000 * 1000

	return rounded
}

func subtractArrays(array1, array2 []float64) []float64 {
	result := make([]float64, len(array1))

	if len(array1) != len(array2) {
		return result
	}

	for i := range array1 {
		result[i] = (array1[i] - array2[i])
	}

	return result
}

func ratioArrays(array1, array2 []float64) []float64 {
	result := make([]float64, len(array1))

	if len(array1) != len(array2) {
		return result
	}

	for i := range array1 {
		result[i] = 100 * (array1[i] - array2[i]) / array2[i]
	}

	return result
}

func replaceLastOccurrence(input, oldChar, newChar string) string {
	lastIndex := strings.LastIndex(input, oldChar)

	if lastIndex == -1 {
		return input
	}

	result := input[:lastIndex] + newChar + input[lastIndex+len(oldChar):]

	return result
}

func addSeparator(number int, separator string) string {
	strNumber := strconv.Itoa(number)

	length := len(strNumber)

	result := make([]byte, 0, length+(length-1)/3)

	for i := length - 1; i >= 0; i-- {
		if (length-i-1)%3 == 0 && i != length-1 {
			result = append(result, separator[0])
		}
		result = append(result, strNumber[i])
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}
