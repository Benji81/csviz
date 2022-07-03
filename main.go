package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

var COLUMN_COLORS = [5]termbox.Attribute{
	termbox.ColorCyan,
	termbox.ColorGreen,
	termbox.ColorLightYellow,
	termbox.ColorLightRed,
	termbox.ColorLightMagenta,
}

var COLOR_COUNT int = len(COLUMN_COLORS)
var BUFFER_SIZE int = 10000

type sectionInfo struct {
	headers        []string
	data           [][]string
	columnCount    int
	firstLineIndex int
	lastLineIndex  int
	eof            bool //eof has been reach during the skip phase so data should be empty
}

//displayHeaderMessage shows a header message at term first lins
func displayHeaderMessage(message string) {
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

//getSection create and fill a sectionInfo struct from a lineOffset in the file
func getSection(lineOffset int, filename string, delimiter rune) *sectionInfo {
	var section sectionInfo
	section.data = make([][]string, BUFFER_SIZE)
	section.eof = false

	//Get some lines before and after de derised line (if possible)
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
			displayHeaderMessage(fmt.Sprintf("Readind... %v%%", percent))
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

//getColumnSizes compute from a given section the maximum size of each column
func getColumnSizes(section sectionInfo, columnCount int) []int {
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

//getColumnOffsets computes the start offset of each column from columnSizes which contains the size of each column
func getColumnOffsets(columnSizes []int) []int {
	var result = make([]int, len(columnSizes))
	result[0] = 0
	for i := 0; i < len(columnSizes)-1; i++ {
		result[i+1] = result[i] + columnSizes[i] + 1
	}

	return result
}

//printHeaders prints a subset of columns titles at term first line with desired columns
func printHeaders(section sectionInfo, digitOffset int, columnOffsets []int, columnOffsetIndex int) {
	termSizeX, _ := termbox.Size()
	for headerIndex, cell := range section.headers[columnOffsetIndex:] {
		x := digitOffset + columnOffsets[headerIndex]
		for _, char := range cell {
			if x >= termSizeX-2 {
				break
			}
			termbox.SetChar(x, 0, char)
			termbox.SetFg(x, 0, COLUMN_COLORS[(headerIndex+columnOffsetIndex)%COLOR_COUNT]|termbox.AttrBold)
			x += 1
		}
	}
}

//printLineIndex prints the line index at the left of a given term line
func printLineIndex(termY int, lineIndex int) {
	lineIndexStr := strconv.Itoa(lineIndex)
	for digitIndex, digit := range lineIndexStr {
		termbox.SetChar(digitIndex, termY, digit)
		termbox.SetFg(digitIndex, termY, termbox.ColorDarkGray)
	}
}

//printLineContent prints a line (record) right of the index (digitoffset)
func printLineContent(section sectionInfo, lineIndex, termY, termSizeX, digitOffset int, columnOffsets []int, columnOffsetIndex int) {
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
			termbox.SetFg(termX, termY, COLUMN_COLORS[(cellIndex+columnOffsetIndex)%COLOR_COUNT])
			termX++
		}
	}
}

// printSection display a text matrix in a termbox
// Text are displayed with column alignment.
func printSection(section sectionInfo, lineOffset, columnOffsetIndex int) {
	error := termbox.Clear(termbox.ColorLightGray, termbox.ColorBlack)
	if error != nil {
		log.Fatal(error)
	}

	if section.eof {
		termbox.Flush()
		return
	}

	columnSizes := getColumnSizes(section, section.columnCount)
	columnOffsets := getColumnOffsets(columnSizes[columnOffsetIndex:])
	digitOffset := int(math.Log10(float64(lineOffset+100)) + 2) // if more than 100 lines , skip 2 more spaces

	printHeaders(section, digitOffset, columnOffsets, columnOffsetIndex)

	termSizeX, termSizeY := termbox.Size()
	firstLineLindex := lineOffset - termSizeY/2 //vertically center the required line
	if firstLineLindex < 0 {
		firstLineLindex = 0
	}

	for termYIndex := 1; termYIndex < termSizeY; termYIndex++ {
		lineIndex := firstLineLindex + termYIndex - 1
		printLineIndex(termYIndex, lineIndex+1)
		printLineContent(section, lineIndex, termYIndex, termSizeX, digitOffset, columnOffsets, columnOffsetIndex)
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

var usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "csviz [-delimiter=\",\"] [-line=42000] file.csv\n")

	flag.PrintDefaults()
}

func main() {
	delimiter := flag.String("delimiter", ",", "Fields delimiter. Default ,")
	line := flag.Int("line", 0, "Start line index")
	flag.Parse()

	lineOffset := *line

	if lineOffset < 0 {
		lineOffset = 0
	}

	fmt.Println("tail:", flag.Args())

	if flag.NArg() == 0 {
		usage()
		os.Exit(1)
	}

	filename := flag.Args()[0]
	columnOffset := 0
	delimRune, _ := utf8.DecodeRuneInString(*delimiter)
	section := getSection(lineOffset, filename, delimRune)

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	//Event loop
	for {
		printSection(*section, lineOffset, columnOffset)
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
			section = getSection(lineOffset, filename, delimRune)
		}
	}

	termbox.Close()
}
