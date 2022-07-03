package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

var columnColors = [5]termbox.Attribute{
	termbox.ColorCyan,
	termbox.ColorGreen,
	termbox.ColorLightYellow,
	termbox.ColorLightRed,
	termbox.ColorLightMagenta,
}

var COLOR_COUNT int = len(columnColors)
var BUFFER_SIZE int = 10000

type sectionInfo struct {
	headers        []string
	data           [][]string
	columnCount    int
	firstLineIndex int
	lastLineIndex  int
	eof            bool
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
func GetSection(lineOffset int, filename string, delimiter rune) *sectionInfo {
	var section sectionInfo
	section.data = make([][]string, BUFFER_SIZE)
	section.eof = false

	sectionLineIndexStart := lineOffset - BUFFER_SIZE/2
	if sectionLineIndexStart < 0 {
		sectionLineIndexStart = 0
	}

	section.firstLineIndex = sectionLineIndexStart
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = delimiter

	record, err := csvReader.Read()
	if err == io.EOF {
		return &section
	}
	if err != nil {
		log.Fatal(err)
	}
	section.headers = record
	section.columnCount = len(record)

	//Skip some lines
	for skip := 0; skip < sectionLineIndexStart; skip++ {
		_, err := csvReader.Read()
		if err != nil {
			section.eof = true
		}
		if skip%100000 == 0 {
			percent := 100 * skip / sectionLineIndexStart
			DisplayHeaderMessage(fmt.Sprintf("Readind... %v%%", percent))
		}
	}

	//Extract content
	for line := 0; line < BUFFER_SIZE; line++ {

		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		section.lastLineIndex = sectionLineIndexStart + line
		section.data[line] = record
	}
	return &section
}

//GetColumnSizes compute from a given section the maximum size of each column
func GetColumnSizes(section sectionInfo, columnCount int) []int {
	var result = make([]int, columnCount)
	for i := 0; i < columnCount; i++ {
		result[i] = len(section.headers[i])
	}

	for _, line := range section.data {
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

func PrintHeaders(section sectionInfo, digitOffset int, columnOffsets []int, columnOffsetIndex int) {
	termSizeX, _ := termbox.Size()
	for headerIndex, cell := range section.headers[columnOffsetIndex:] {
		x := digitOffset + columnOffsets[headerIndex]
		for _, char := range cell {
			if x >= termSizeX-2 {
				break
			}
			termbox.SetChar(x, 0, char)
			termbox.SetFg(x, 0, columnColors[(headerIndex+columnOffsetIndex)%COLOR_COUNT]|termbox.AttrBold)
			x += 1
		}
	}
}

func PrintLineIndex(termY int, lineIndex int) {
	lineIndexStr := strconv.Itoa(lineIndex)
	for digitIndex, digit := range lineIndexStr {
		termbox.SetChar(digitIndex, termY, digit)
		termbox.SetFg(digitIndex, termY, termbox.ColorDarkGray)
	}
}

func PrintLineContent(section sectionInfo, lineIndex, termY, termSizeX, digitOffset int, columnOffsets []int, columnOffsetIndex int) {

	for cellIndex, cell := range section.data[lineIndex-section.firstLineIndex][columnOffsetIndex:] {
		termX := digitOffset + columnOffsets[cellIndex]
		for _, char := range cell {
			if termX >= termSizeX-2 {
				break
			}
			if char == '\n' {
				char, _ = utf8.DecodeRuneInString("\u23CE")
			}
			termbox.SetChar(termX, termY, char)
			if termY == 0 {
			} else {
				termbox.SetFg(termX, termY, columnColors[(cellIndex+columnOffsetIndex)%COLOR_COUNT])
			}

			termX++
		}
	}

}

// PrintSection display a text matrix in a termbox
// Text are displayed with column alignment.
func PrintSection(section sectionInfo, lineOffset, columnOffsetIndex int) {
	error := termbox.Clear(termbox.ColorLightGray, termbox.ColorBlack)
	if error != nil {
		log.Fatal(error)
	}

	if section.eof {
		termbox.Flush()
		return
	}

	//TODO move this to init

	columnSizes := GetColumnSizes(section, section.columnCount)
	columnOffsets := GetColumnOffsets(columnSizes[columnOffsetIndex:])
	digitOffset := int(math.Log10(float64(lineOffset+100)) + 2) // if more thatn 100 lines , skip 2 more spaces

	PrintHeaders(section, digitOffset, columnOffsets, columnOffsetIndex)

	termSizeX, termSizeY := termbox.Size()
	firstLineLindex := lineOffset - termSizeY/2
	if firstLineLindex < 0 {
		firstLineLindex = 0
	}

	for termYIndex := 1; termYIndex < termSizeY; termYIndex++ {
		lineIndex := firstLineLindex + termYIndex - 1
		PrintLineIndex(termYIndex, lineIndex+1)
		PrintLineContent(section, lineIndex, termYIndex, termSizeX, digitOffset, columnOffsets, columnOffsetIndex)
		if termYIndex >= termSizeY-2 || lineIndex >= section.lastLineIndex {
			break
		}
	}

	errorSync := termbox.Sync()
	if errorSync != nil {
		log.Fatal(errorSync)
	}
	termbox.Flush()
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	if len(os.Args) < 2 {
		panic("Need a file as argument")
	}

	lineOffset := 0

	if len(os.Args) == 3 {
		lineOffset, err = strconv.Atoi(os.Args[2])
		if lineOffset < 0 {
			lineOffset = 0
		}

		if err != nil {
			panic("Second arguement must be an valid interger")
		}

	}
	filename := os.Args[1]
	delimiter := ','
	columnOffset := 0

	section := GetSection(lineOffset, filename, delimiter)

	for {
		PrintSection(*section, lineOffset, columnOffset)
		event := termbox.PollEvent()
		if event.Key == termbox.KeyEsc || event.Key == termbox.KeyCtrlC {
			break
		} else if event.Type == termbox.EventResize {
		} else if event.Key == termbox.KeyArrowRight && columnOffset < section.columnCount-1 {
			columnOffset += 1
		} else if event.Key == termbox.KeyArrowLeft && columnOffset > 0 {
			columnOffset -= 1
		} else if event.Key == termbox.KeyArrowUp && lineOffset > 0 {
			lineOffset -= 1
		} else if event.Key == termbox.KeyArrowDown && !section.eof {
			lineOffset += 1
		} else if event.Key == termbox.KeyPgup && lineOffset > 0 {
			lineOffset -= 100
		} else if event.Key == termbox.KeyPgdn && !section.eof {
			lineOffset += 100
		} else if event.Key == termbox.KeyHome {
			lineOffset = 0
		}

		if lineOffset < section.firstLineIndex || lineOffset > section.lastLineIndex {
			section = GetSection(lineOffset, filename, delimiter)
		}
	}

	termbox.Close()
}
