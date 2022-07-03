package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/nsf/termbox-go"
)

var columnColors = [5]termbox.Attribute{
	termbox.ColorCyan,
	termbox.ColorGreen,
	termbox.ColorLightYellow,
	termbox.ColorLightRed,
	termbox.ColorLightMagenta,
}

type sectionInfo struct {
	data          [][]string
	columnCount   int
	lastLineIndex int
}

//DisplayHeaderMessage shows a header message
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

//GetSection computes a text matrix of lineCount x cellCount, starting at lineStart, cellStart of a given csvReader
func GetSection(lineStart, lineCount, cellStart int, csvReader *csv.Reader) *sectionInfo {
	var section sectionInfo
	section.data = make([][]string, lineCount)

	record, err := csvReader.Read()
	if err == io.EOF {
		return &section
	}
	if err != nil {
		log.Fatal(err)
	}
	section.data[0] = record
	section.columnCount = len(record)

	//Skip some lines
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

	//Extract content
	for line := 1; line < lineCount; line++ {

		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		section.lastLineIndex = lineStart + line
		section.data[line] = record
	}
	return &section
}

//GetColumnSizes compute from a given section the maximum size of each column
func GetColumnSizes(section [][]string, columnCount int) []int {
	var result = make([]int, columnCount)
	for i := 0; i < columnCount; i++ {
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
func PrintSection(section sectionInfo, columnOffset int, lineOffset int) {
	data := section.data
	error := termbox.Clear(termbox.ColorLightGray, termbox.ColorBlack)
	if error != nil {
		log.Fatal(error)
	}

	//TODO move this to init
	colorCount := len(columnColors)

	columnSizes := GetColumnSizes(section.data, section.columnCount)
	columnOffsets := GetColumnOffsets(columnSizes[columnOffset:])
	termX, termY := termbox.Size()
	x := 0
	y := 0
	lineIndex := lineOffset
	digitOffset := int(math.Log10(float64(lineIndex+100)) + 2) // if more thatn 100 lines , skip 2 more spaces
	for _, line := range data {
		lineIndexStr := strconv.Itoa(lineIndex + y)
		if y > 0 {
			for digitIndex, digit := range lineIndexStr {
				termbox.SetChar(digitIndex, y, digit)
				termbox.SetFg(digitIndex, y, termbox.ColorDarkGray)
			}
		}

		for cellIndex, cell := range line[columnOffset:] {
			x = digitOffset + columnOffsets[cellIndex]
			for _, char := range cell {
				if x >= termX-2 {
					break
				}
				termbox.SetChar(x, y, char)
				if y == 0 {
					termbox.SetFg(x, y, columnColors[(cellIndex+columnOffset)%colorCount]|termbox.AttrBold)
				} else {
					termbox.SetFg(x, y, columnColors[(cellIndex+columnOffset)%colorCount])
				}

				x++
			}
		}
		if y >= termY-2 {
			break
		}
		y++
	}
	errorSync := termbox.Sync()
	if errorSync != nil {
		log.Fatal(errorSync)
	}
	termbox.Flush()
}

// ReadAndDraw read the given CSV given with filename augument and draws it
func ReadAndDraw(filename string, columnOffset int) *sectionInfo {

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = ','
	_, lineCount := termbox.Size()
	lineStart := 0
	section := GetSection(lineStart, lineCount, 0, csvReader)

	PrintSection(*section, columnOffset, lineStart)
	return section
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

	columnOffset := 0
	section := ReadAndDraw(filename, columnOffset)
	fmt.Println(section.columnCount)
	for {
		event := termbox.PollEvent()
		if event.Key == termbox.KeyEsc || event.Key == termbox.KeyCtrlC {
			break
		} else if event.Type == termbox.EventResize {
			ReadAndDraw(filename, columnOffset)
		} else if event.Key == termbox.KeyArrowRight && columnOffset < section.columnCount {
			columnOffset += 1
			ReadAndDraw(filename, columnOffset)
		} else if event.Key == termbox.KeyArrowLeft && columnOffset > 0 {
			columnOffset -= 1
			ReadAndDraw(filename, columnOffset)
		}
	}

	termbox.Close()
}
