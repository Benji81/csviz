package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"

	"github.com/nsf/termbox-go"
)

func GetSection(lineStart, lineCount, cellStart, cellCount int, csvReader *csv.Reader) [][]string {
	for skip := 0; skip < lineStart; skip++ {
		_, err := csvReader.Read()
		if err != nil {
			panic(err)
		}
	}

	var section = make([][]string, lineCount)

	for line := 0; line < lineCount; line++ {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var subSection = make([]string, cellCount)
		for cellIndex := 0; cellIndex < cellCount; cellIndex++ {
			subSection[cellIndex] = record[cellStart+cellIndex]
		}
		section[line] = subSection
	}
	return section
}

func GetColumnSizes(section [][]string) []int {
	var result = make([]int, len(section))
	for i := 0; i < len(section); i++ {
		result[i] = -1
	}

	for _, line := range section {
		for cellIndex, cell := range line {
			if result[cellIndex] < len(cell) {
				result[cellIndex] = len(cell)
			}
		}
	}
	return result
}

func GetColumnOffsets(columnSizes []int) []int {
	var result = make([]int, len(columnSizes))
	result[0] = 0
	for i := 0; i < len(columnSizes)-1; i++ {
		result[i+1] = result[i] + columnSizes[i] + 1
	}

	return result
}

func PrintSection(section [][]string) {
	colorCount := 5
	var colors = make([]termbox.Attribute, colorCount)
	colors[0] = termbox.ColorCyan         //termbox.RGBToAttribute(0, 127, 100)
	colors[1] = termbox.ColorGreen        //termbox.RGBToAttribute(0, 100, 127)
	colors[2] = termbox.ColorLightYellow  //termbox.RGBToAttribute(0, 100, 127)
	colors[3] = termbox.ColorLightRed     //termbox.RGBToAttribute(0, 100, 127)
	colors[4] = termbox.ColorLightMagenta //termbox.RGBToAttribute(0, 100, 127)

	columnSizes := GetColumnSizes(section)
	columnOffsets := GetColumnOffsets(columnSizes)

	x := 0
	y := 0
	for _, line := range section {
		for cellIndex, cell := range line {
			x = columnOffsets[cellIndex]
			for _, char := range cell {
				termbox.SetChar(x, y, char)
				termbox.SetFg(x, y, colors[cellIndex%colorCount])
				termbox.SetFg(x, y, colors[cellIndex%colorCount])
				//termbox.SetBg(x, y, termbox.ColorBlack)
				x++
			}
			x++
		}
		y++
	}
	termbox.Flush()
}

// ReadAndDraw read the given CSV given with filename augument and draws it
func ReadAndDraw(filename string) {

	f, err := os.Open("data.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	termbox.Clear(termbox.ColorLightGray, termbox.ColorBlack)

	csvReader := csv.NewReader(f)
	csvReader.Comma = ','
	section := GetSection(4000000, 100, 0, 8, csvReader)

	PrintSection(section)
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	if len(os.Args) != 2 {
		panic("Need a file as argument")
	}
	filename := os.Args[1]

	ReadAndDraw(filename)

	for {
		event := termbox.PollEvent()
		if event.Key == termbox.KeyEsc || event.Key == termbox.KeyCtrlC {
			break
		} else if event.Type == termbox.EventResize {
			ReadAndDraw(filename)
		}
	}

	termbox.Close()
}
