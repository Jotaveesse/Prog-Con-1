package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	chart "github.com/wcharczuk/go-chart/v2"
)

type TestResult struct {
	Name    string
	Ns      []int
	Results []float64
	Memory  []uint64
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
				joinedTest.Memory = append(test1.Memory, test2.Memory...)

				joinedTests = append(joinedTests, joinedTest)
			}
		}
	}

	return joinedTests
}

func main() {
	names := []string{"Sequencial", "Concorrente", "Concorrente Melhorado"}
	functions := []func(int) []int{sieve, concSieve, blockConcSieve}
	testNs1 := []int{500000, 1000000, 5000000}
	testNs2 := []int{10000000, 50000000, 100000000}

	tests1 := runTests(names, functions, testNs1)
	tests2 := runTests(names, functions, testNs2)
	testsJoined := joinTests(tests1, tests2)

	makeBarChart(tests1, "comp-tempo")
	makeBarChart(tests2, "comp-tempo2")
	makeBarMemChart(tests1, "comp-mem")
	makeBarMemChart(tests2, "comp-mem2")
	makeDiffLineGraph(testsJoined[1:3], "diff")
	makeDiffPercLineGraph(testsJoined[1:3], "diff-perc")
}

func runTests(names []string, testFuncs []func(int) []int, nArr []int) []TestResult {
	avrgLoops := 30

	var results []TestResult
	var m runtime.MemStats

	//para cada função
	for i, function := range testFuncs {
		var res TestResult
		res.Name = names[i]

		//para valor de N
		for _, n := range nArr {
			highestMem := uint64(0)
			start_time := time.Now()

			//tira a media de tempo e memoria de todos os loops
			for k := 0; k < avrgLoops; k++ {
				function(n)

				runtime.ReadMemStats(&m)
				highestMem += m.Alloc
			}

			end_time := time.Now()

			//chama garbage cleaner so por preucaução
			runtime.GC()

			micro := float64(end_time.Sub(start_time).Microseconds()) / float64(avrgLoops)

			res.Ns = append(res.Ns, n)
			res.Results = append(res.Results, micro)
			res.Memory = append(res.Memory, highestMem/uint64(avrgLoops))
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

func makeBarMemChart(tests []TestResult, outputFile string) {
	var yTicks []chart.Tick
	maxDiff := math.Inf(-1)

	var bars []chart.Value

	for i := 0; i < len(tests[0].Memory); i++ {

		for j, test := range tests {
			sty := chart.Style{
				FillColor:   chart.DefaultColors[j],
				StrokeColor: chart.DefaultColors[j],
				StrokeWidth: 0,
			}

			result := float64(test.Memory[i]) / 1024

			barValue := chart.Value{Value: result, Label: fmt.Sprintf("%s", addSeparator(test.Ns[i], ".")), Style: sty}
			bars = append(bars, barValue)

			if result > maxDiff {
				maxDiff = result
			}
		}
	}

	title := "Média de Uso de Memória"
	for _, test := range tests {
		title += " " + test.Name + ","
	}

	title = replaceLastOccurrence(title, ",", "")
	title = replaceLastOccurrence(title, ",", " e")

	for i := 0; i <= 10; i++ {
		val := float64(i * roundToNextThousand(int(maxDiff)/10))
		yTicks = append(yTicks, chart.Tick{Value: val, Label: fmt.Sprintf(addSeparator(int(val), "."))})
	}

	graph := chart.BarChart{
		Title:      fmt.Sprintf(title),
		TitleStyle: chart.Style{FontSize: 14},

		YAxis: chart.YAxis{
			Name:  "KB",
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

// func makeLineGraph(n int, name string) {
// 	loops := 8
// 	avrgLoops := 6
// 	var xValues []float64
// 	var yValuesSeq []float64
// 	var xTicks []chart.Tick
// 	var yTicks []chart.Tick

// 	var maxVal float64

// 	for i := 0; i < loops; i++ {

// 		start_time := time.Now()
// 		for j := 0; j < avrgLoops; j++ {
// 			sieve(n)
// 		}
// 		end_time := time.Now()

// 		micro := end_time.Sub(start_time).Microseconds() / int64(avrgLoops)

// 		xValues = append(xValues, float64(i))
// 		xTicks = append(xTicks, chart.Tick{Value: float64(i), Label: fmt.Sprintf("%v", i)})

// 		yValuesSeq = append(yValuesSeq, float64(micro))

// 		if float64(micro) > maxVal {
// 			maxVal = float64(micro)
// 		}

// 	}

// 	var yValuesConc []float64
// 	for i := 0; i < loops; i++ {

// 		start_time := time.Now()
// 		for j := 0; j < avrgLoops; j++ {
// 			conc_sieve(n)
// 		}
// 		end_time := time.Now()

// 		micro := end_time.Sub(start_time).Microseconds() / int64(avrgLoops)

// 		yValuesConc = append(yValuesConc, float64(micro))

// 		if float64(micro) > maxVal {
// 			maxVal = float64(micro)
// 		}
// 	}

// 	for i := 0; i <= 10; i++ {
// 		val := i * (roundToNextThousand(int(maxVal)) / 10)
// 		yTicks = append(yTicks, chart.Tick{Value: float64(val), Label: fmt.Sprintf("%v", val)})
// 	}

// 	graph := chart.Chart{
// 		Title:      fmt.Sprintf("Comparação para N = %.0e", float64(n)),
// 		TitleStyle: chart.Style{FontSize: 14},
// 		XAxis: chart.XAxis{
// 			Name:  "Execução",
// 			Ticks: xTicks,
// 		},
// 		YAxis: chart.YAxis{
// 			Name:  "Microsegundos",
// 			Ticks: yTicks,
// 		},
// 		Series: []chart.Series{
// 			chart.ContinuousSeries{
// 				Name: "Sequencial",
// 				Style: chart.Style{
// 					StrokeColor: chart.DefaultColors[0],
// 					StrokeWidth: 2,
// 				},
// 				XValues: xValues,
// 				YValues: yValuesSeq,
// 			},
// 			chart.ContinuousSeries{
// 				Name: "Concorrente",
// 				Style: chart.Style{
// 					StrokeColor: chart.DefaultColors[1],
// 					StrokeWidth: 2,
// 				},
// 				XValues: xValues,
// 				YValues: yValuesConc,
// 			},
// 		},
// 	}

// 	graph.Elements = []chart.Renderable{
// 		chart.Legend(&graph),
// 	}

// 	file, _ := os.Create("graphs/"+name + ".png")
// 	defer file.Close()
// 	graph.Render(chart.PNG, file)
// }

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
