package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nsf/termbox-go"
)

//DisplayHeaderMessage show a header message
func DisplayHeaderMessage(message string) {
	x_offset := 0
	y_offset := 0
	x := x_offset
	for _, char := range message {
		termbox.SetChar(x, y_offset, char)
		termbox.SetFg(x, y_offset, termbox.ColorWhite)
		x++
	}
	termbox.Flush()

}

//GetSection compute a text matrix of lineCount x cellCount, starting at lineStart, cellStart of a given csvReader
func GetSection(lineStart, lineCount, cellStart, cellCount int, csvReader *csv.Reader) [][]string {
	for skip := 0; skip < lineStart; skip++ {
		_, err := csvReader.Read()
		if err != nil {
			panic(err)
		}
		if skip%100000 == 0 {
			percent := 100 * skip / lineStart
			DisplayHeaderMessage(fmt.Sprintf("Readind... %v%%", percent))
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

//GetColumnSizes compute from a given section the maximum size of each column
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

//GetColumnOffsets computes the start offset of each column from columnSizes which contains the size of each column
func GetColumnOffsets(columnSizes []int) []int {
	var result = make([]int, len(columnSizes))
	result[0] = 0
	for i := 0; i < len(columnSizes)-1; i++ {
		result[i+1] = result[i] + columnSizes[i] + 1
	}

	return result
}

// PrintSection display a text matrix in a termbox
// Text are displayed with column alignment.
func PrintSection(section [][]string) {
	termbox.Clear(termbox.ColorLightGray, termbox.ColorBlack)

	//TODO move this to init
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
