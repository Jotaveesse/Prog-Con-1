package graph

import (
	"exercicio4/client"
	"fmt"
	chart "github.com/wcharczuk/go-chart/v2"
	"math"
	"os"
	"strconv"
	"strings"
)

type TestResult struct {
	Name    string
	Ns      []int
	Results []float64
}

func Run() {
	var iterations int

	fmt.Print("Choose how many iterations: ")
	fmt.Scan(&iterations)

	names := []string{"UDP", "TCP", "RCP"}
	//functions := []func(int, string) ([]int, time.Duration){client.sendMessageUDP, client.sendMessageTCP}
	testNs1 := []int{300, 1000, 3000}
	testNs2 := []int{10000, 30000, 100000}

	//divididos em 2 testes para terem 2 graficos de barras, ja que ficaria feio todos em um unico grafico
	tests1 := runTests(iterations, names, testNs1)
	tests2 := runTests(iterations, names, testNs2)
	testsJoined := joinTests(tests1, tests2)

	makeBarChart(tests1, "comp-tempo")
	makeBarChart(tests2, "comp-tempo2")
	makeDiffLineGraph(testsJoined[1], testsJoined[2], "diff")
	makeDiffPercLineGraph(testsJoined[1], testsJoined[2], "diff-perc")
	makeGrowthLineGraph(testsJoined, "growth")
}

func runTests(iterations int, names []string, nArr []int) []TestResult {
	var results []TestResult

	warmUpAmount := int(float64(iterations) * 0.1)

	var resTCP TestResult
	resTCP.Name = names[0]

	//para cada valor de N
	for _, n := range nArr {
		totalRtt := 0

		conn := client.StartConnectionUDP()

		for k := 0; k < iterations+(2*warmUpAmount); k++ {
			_, rtt := client.SendMessageUDP(conn, n, "blk_conc")
			//fmt.Println(totalRtt)

			//nao contabiliza os 10% primeiros e ultimos
			if k > warmUpAmount || k < iterations+warmUpAmount {
				totalRtt += int(rtt.Microseconds())
			}
		}
		client.CloseConnectionUDP(conn)

		avrgMicro := float64(totalRtt) / float64(iterations)

		resTCP.Ns = append(resTCP.Ns, n)
		resTCP.Results = append(resTCP.Results, avrgMicro)
	}

	results = append(results, resTCP)

	var resUDP TestResult
	resUDP.Name = names[1]

	//para cada valor de N
	for _, n := range nArr {
		totalRtt := 0

		conn := client.StartConnectionTCP()

		for k := 0; k < iterations+(2*warmUpAmount); k++ {
			_, rtt := client.SendMessageTCP(conn, n, "blk_conc")

			//nao contabiliza os 10% primeiros e ultimos
			if k > warmUpAmount || k < iterations+warmUpAmount {
				totalRtt += int(rtt.Microseconds())
			}
		}
		client.CloseConnectionTCP(conn)

		avrgMicro := float64(totalRtt) / float64(iterations)

		resUDP.Ns = append(resUDP.Ns, n)
		resUDP.Results = append(resUDP.Results, avrgMicro)
	}

	results = append(results, resUDP)

	var resRCP TestResult
	resRCP.Name = names[2]

	//para cada valor de N
	for _, n := range nArr {
		totalRtt := 0

		for k := 0; k < iterations+(2*warmUpAmount); k++ {
			_, rtt := client.ClientRPC(n, "blk_conc")

			//nao contabiliza os 10% primeiros e ultimos
			if k > warmUpAmount || k < iterations+warmUpAmount {
				totalRtt += int(rtt.Microseconds())
			}
		}

		avrgMicro := float64(totalRtt) / float64(iterations)

		resRCP.Ns = append(resRCP.Ns, n)
		resRCP.Results = append(resRCP.Results, avrgMicro)
	}

	results = append(results, resRCP)

	return results
}

func joinTests(tests1 []TestResult, tests2 []TestResult) []TestResult {
	var joinedTests []TestResult
	client.SieveClientTCP(2, "")

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

func makeBarChart(tests []TestResult, outputFile string) {
	var yTicks []chart.Tick
	maxDiff := math.Inf(-1)

	var bars []chart.Value

	//para cada N
	for i := 0; i < len(tests[0].Results); i++ {

		//para cada teste
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

	maxDiff = float64(roundToNextNum(int(maxDiff), 1000))

	//cria 10 marcadores verticais em valores arredondados com base no maximo e minimo
	for i := int64(0); i <= 10; i++ {
		val := float64(int(i*int64(maxDiff)) / 10)
		yTicks = append(yTicks, chart.Tick{Value: val, Label: fmt.Sprintf("%.0f", val)})
	}

	//cria o grafico
	graph := chart.BarChart{
		Title:      title,
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

	//cria o arquivo de imagem
	file, err := os.Create("graphs/" + outputFile + ".png")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Diretório não encontrado, execute o código dentro do diretório do exercício")
		} else {
			fmt.Println(err)
		}
	}
	defer file.Close()
	graph.Render(chart.PNG, file)
}

func makeDiffPercLineGraph(subj1 TestResult, subj2 TestResult, outputFile string) {
	var xTicks []chart.Tick
	var yTicks []chart.Tick
	var xValues []float64

	//cria pontos com base an quantidade de Ns
	for i := range subj1.Results {
		xValues = append(xValues, float64(i))
	}

	//cria os valores no eixo pra cada N
	for j, n := range subj1.Ns {
		xTicks = append(xTicks, chart.Tick{Value: float64(j), Label: fmt.Sprintf("%s", addSeparator(n, "."))})
	}

	//calcula as porcentagens
	ratios := ratioArrays(subj2.Results, subj1.Results)

	maxDiff := math.Inf(-1)
	minDiff := math.Inf(1)

	//calcula o valores maximo e minimo
	for _, diff := range ratios {
		if diff > maxDiff {
			maxDiff = diff
		}
		if diff < minDiff {
			minDiff = diff
		}
	}

	//cria 10 marcadores verticais com base no maximo e minimo
	for i := 0; i <= 10; i++ {
		val := minDiff + float64(i)*((maxDiff-minDiff)/10)
		yTicks = append(yTicks, chart.Tick{Value: val, Label: fmt.Sprintf("%.2f%%", val)})
	}

	title := "Diferença Relativa Entre " + subj1.Name + " e " + subj2.Name

	//cria o grafico
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

	//cria o arquivo de imagem
	file, err := os.Create("graphs/" + outputFile + ".png")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Diretório não encontrado, execute o código dentro do diretório do exercício")
		} else {
			fmt.Println(err)
		}
	}
	defer file.Close()
	graph.Render(chart.PNG, file)
}

func makeDiffLineGraph(subj1 TestResult, subj2 TestResult, outputFile string) {
	var xTicks []chart.Tick
	var yTicks []chart.Tick
	var xValues []float64

	//cria pontos com base an quantidade de Ns
	for i := range subj1.Results {
		xValues = append(xValues, float64(i))
	}

	//cria os valores no eixo pra cada N
	for j, n := range subj1.Ns {
		xTicks = append(xTicks, chart.Tick{Value: float64(j), Label: fmt.Sprintf("%s", addSeparator(n, "."))})
	}

	//substrai a diferença dos valores
	differences := subtractArrays(subj1.Results, subj2.Results)

	maxDiff := math.Inf(-1)
	minDiff := math.Inf(1)

	//calcula o valores maximo e minimo
	for _, diff := range differences {
		if diff > maxDiff {
			maxDiff = diff
		}
		if diff < minDiff {
			minDiff = diff
		}
	}

	minDiff = float64(roundToNextNum(int(minDiff), 1000)) - 1000
	maxDiff = float64(roundToNextNum(int(maxDiff), 1000))

	//cria 10 marcadores verticais em valores arredondados com base no maximo e minimo
	for i := 0; i <= 10; i++ {
		val := minDiff + float64(i)*(maxDiff-minDiff)/10
		yTicks = append(yTicks, chart.Tick{Value: val, Label: fmt.Sprintf("%.0f", val)})
	}

	title := "Diferença Entre " + subj1.Name + " e " + subj2.Name

	//cria o gráfico
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
				YValues: differences,
			},
		},
	}

	//cria o arquivo de imagem
	file, err := os.Create("graphs/" + outputFile + ".png")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Diretório não encontrado, execute o código dentro do diretório do exercício")
		} else {
			fmt.Println(err)
		}
	}
	defer file.Close()
	graph.Render(chart.PNG, file)
}

func makeGrowthLineGraph(tests []TestResult, outputFile string) {
	var xTicks []chart.Tick
	var yTicks []chart.Tick
	var xValues []float64

	//cria pontos com base an quantidade de Ns
	for i := range tests[0].Results {
		xValues = append(xValues, float64(i))
	}

	//cria os valores no eixo pra cada N
	for j, n := range tests[0].Ns {
		xTicks = append(xTicks, chart.Tick{Value: float64(j), Label: fmt.Sprintf("%s", addSeparator(n, "."))})
	}

	//calcula as porcentagens
	var growths [][]float64
	var lineSeries []chart.Series

	//calcula os crescimentos e cria as series do gráfico
	for i, test := range tests {
		growth := calculateGrowth(test.Results)

		growths = append(growths, growth)

		lineSeries = append(lineSeries, chart.ContinuousSeries{
			Style: chart.Style{
				StrokeColor: chart.DefaultColors[i],
				StrokeWidth: 2,
			},
			XValues: xValues,
			YValues: growth,
		})
	}

	maxDiff := math.Inf(-1)
	minDiff := math.Inf(1)

	//calcula o valores maximo e minimo
	for _, growth := range growths {
		for _, value := range growth {
			if value > maxDiff {
				maxDiff = value
			}
			if value < minDiff {
				minDiff = value
			}
		}
	}

	//cria 10 marcadores verticais em valores arredondados com base no maximo e minimo
	for i := 0; i <= 10; i++ {
		val := minDiff + float64(i)*((maxDiff-minDiff)/10)
		yTicks = append(yTicks, chart.Tick{Value: val, Label: fmt.Sprintf("%d%%", roundToNextNum(int(val), 10))})
	}

	title := "Aumento de Tempo"
	for _, test := range tests {
		title += " " + test.Name + ","
	}

	title = replaceLastOccurrence(title, ",", "")
	title = replaceLastOccurrence(title, ",", " e")

	//cria o grafico
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
		Series: lineSeries,
	}

	//cria o arquivo de imagem
	file, err := os.Create("graphs/" + outputFile + ".png")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Diretório não encontrado, execute o código dentro do diretório do exercício")
		} else {
			fmt.Println(err)
		}
	}
	defer file.Close()
	graph.Render(chart.PNG, file)
}

func calculateGrowth(data []float64) []float64 {
	growth := make([]float64, len(data))
	growth[0] = 0
	previous := data[0]

	for i, value := range data[1:] {
		if previous == 0 {
			growth[i+1] = 0
		} else {
			growth[i+1] = (value - previous) / previous * 100
		}
		previous = value
	}

	return growth
}

func roundToNextNum(value, rounder int) int {
	rounded := (value + (rounder-1)) / rounder * rounder
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
