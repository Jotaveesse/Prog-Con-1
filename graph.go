package main

import (
	"fmt"
	"math"
	"os"
	"time"

	chart "github.com/wcharczuk/go-chart/v2"
)

func main() {
	makeDiffLineGraph([]int{500000, 1000000, 5000000, 10000000, 50000000}, "diff-chart")
	makeDiffPercLineGraph([]int{500000, 1000000, 5000000, 10000000, 50000000}, "perc-chart")

	makeLineGraph(500000, "chart1")
	makeLineGraph(1000000, "chart2")
	makeLineGraph(5000000, "chart3")
	makeLineGraph(10000000, "chart4")
	makeLineGraph(50000000, "chart5")
}

func makeDiffPercLineGraph(nArr []int, name string) {
	avrgLoops := 20
	var xValues []float64
	var yValuesSeq []float64
	var xTicks []chart.Tick
	var yTicks []chart.Tick

	for i, n := range nArr {

		start_time := time.Now()
		for j := 0; j < avrgLoops; j++ {
			sieve(n)
		}
		end_time := time.Now()

		micro := end_time.Sub(start_time).Microseconds() / int64(avrgLoops)

		xValues = append(xValues, float64(i))
		xTicks = append(xTicks, chart.Tick{Value: float64(i), Label: fmt.Sprintf("%.0e", float64(nArr[i]))})

		yValuesSeq = append(yValuesSeq, float64(micro))

	}

	var yValuesConc []float64
	for _, n := range nArr {

		start_time := time.Now()
		for j := 0; j < avrgLoops; j++ {
			conc_sieve(n)
		}
		end_time := time.Now()

		micro := end_time.Sub(start_time).Microseconds() / int64(avrgLoops)

		yValuesConc = append(yValuesConc, float64(micro))

	}

	ratios := ratioArrays(yValuesSeq, yValuesConc)

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

	graph := chart.Chart{
		Title:      fmt.Sprintf("Diferença Relativa Entre Sequencial e Concorrente"),
		TitleStyle: chart.Style{FontSize: 14},
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

	file, _ := os.Create(name + ".png")
	defer file.Close()
	graph.Render(chart.PNG, file)
}

func makeDiffLineGraph(nArr []int, name string) {
	avrgLoops := 20
	var xValues []float64
	var yValuesSeq []float64
	var xTicks []chart.Tick
	var yTicks []chart.Tick

	for i, n := range nArr {

		start_time := time.Now()
		for j := 0; j < avrgLoops; j++ {
			sieve(n)
		}
		end_time := time.Now()

		micro := end_time.Sub(start_time).Microseconds() / int64(avrgLoops)

		xValues = append(xValues, float64(i))
		xTicks = append(xTicks, chart.Tick{Value: float64(i), Label: fmt.Sprintf("%.0e", float64(nArr[i]))})

		yValuesSeq = append(yValuesSeq, float64(micro))

	}

	var yValuesConc []float64
	for _, n := range nArr {

		start_time := time.Now()
		for j := 0; j < avrgLoops; j++ {
			conc_sieve(n)
		}
		end_time := time.Now()

		micro := end_time.Sub(start_time).Microseconds() / int64(avrgLoops)

		yValuesConc = append(yValuesConc, float64(micro))

	}

	differences := subtractArrays(yValuesSeq, yValuesConc)

	maxDiff := math.Inf(-1)
	minDiff := math.Inf(1)

	for _, diff := range differences {
		if diff > maxDiff {
			maxDiff = diff
		}
		if diff < minDiff {
			minDiff = diff
		}
	}
	for i := 0; i <= 10; i++ {
		val := int(minDiff) + i*(roundToNextThousand(int(maxDiff-minDiff))/10)
		yTicks = append(yTicks, chart.Tick{Value: float64(val), Label: fmt.Sprintf("%v", val)})
	}

	graph := chart.Chart{
		Title:      fmt.Sprintf("Diferença Entre Sequencial e Concorrente"),
		TitleStyle: chart.Style{FontSize: 14},
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
				YValues: differences,
			},
		},
	}

	file, _ := os.Create(name + ".png")
	defer file.Close()
	graph.Render(chart.PNG, file)
}

func makeLineGraph(n int, name string) {
	loops := 8
	avrgLoops := 6
	var xValues []float64
	var yValuesSeq []float64
	var xTicks []chart.Tick
	var yTicks []chart.Tick

	var maxVal float64

	for i := 0; i < loops; i++ {

		start_time := time.Now()
		for j := 0; j < avrgLoops; j++ {
			sieve(n)
		}
		end_time := time.Now()

		micro := end_time.Sub(start_time).Microseconds() / int64(avrgLoops)

		xValues = append(xValues, float64(i))
		xTicks = append(xTicks, chart.Tick{Value: float64(i), Label: fmt.Sprintf("%v", i)})

		yValuesSeq = append(yValuesSeq, float64(micro))

		if float64(micro) > maxVal {
			maxVal = float64(micro)
		}

	}

	var yValuesConc []float64
	for i := 0; i < loops; i++ {

		start_time := time.Now()
		for j := 0; j < avrgLoops; j++ {
			conc_sieve(n)
		}
		end_time := time.Now()

		micro := end_time.Sub(start_time).Microseconds() / int64(avrgLoops)

		yValuesConc = append(yValuesConc, float64(micro))

		if float64(micro) > maxVal {
			maxVal = float64(micro)
		}
	}

	for i := 0; i <= 10; i++ {
		val := i * (roundToNextThousand(int(maxVal)) / 10)
		yTicks = append(yTicks, chart.Tick{Value: float64(val), Label: fmt.Sprintf("%v", val)})
	}

	graph := chart.Chart{
		Title:      fmt.Sprintf("Comparação para N = %v", n),
		TitleStyle: chart.Style{FontSize: 14},
		XAxis: chart.XAxis{
			Name:  "Execução",
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
				YValues: yValuesSeq,
			},
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: chart.DefaultColors[1],
					StrokeWidth: 2,
				},
				XValues: xValues,
				YValues: yValuesConc,
			},
		},
	}

	file, _ := os.Create(name + ".png")
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
		result[i] = 100 * (array1[i] - array2[i]) / array1[i]
	}

	return result
}
