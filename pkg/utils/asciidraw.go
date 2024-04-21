package utils

import (
	"fmt"
	"strings"
)

type AsciiFrameField struct {
	// Name of the field
	Name string

	// Units within the frame the field begins from
	Begin int

	// Field width
	Width int
}

// The last unit within the frame used by this field
func (f *AsciiFrameField) TopUnit() int {
	return f.PastTopUnit() - 1
}

// The first unit within the frame used by the next field
func (f *AsciiFrameField) PastTopUnit() int {
	return f.Begin + f.Width
}

type AsciiFrameUnitLayout uint

const (
	// Units increase left to right
	AsciiFrameUnitLayout_LeftToRight AsciiFrameUnitLayout = iota
	// Units increase right to left
	AsciiFrameUnitLayout_RightToLeft
)

type asciiFrame struct {
	fields     []AsciiFrameField
	frameWidth int
	unit       string
	leftpad    int
	layout     AsciiFrameUnitLayout
}

func (f *asciiFrame) TopUnit() int {
	return f.frameWidth - 1
}

func writeRow(text string, textDecorationExtraLenght int, filler string, length int, builder *strings.Builder) {
	if len(filler) > 1 {
		panic(fmt.Errorf("filler '%v' must be one character long", filler))
	}

	if len(text) > length {
		panic(fmt.Errorf("text '%v' is %v chars long but target length is only %v chars", text, len(text), length))
	}

	leftpad_length := (length - len(text) - textDecorationExtraLenght) / 2
	rightpad_length := (length - len(text) - textDecorationExtraLenght) / 2
	rightpad_length += (length - leftpad_length - len(text) - textDecorationExtraLenght - rightpad_length)

	for i := 0; i < leftpad_length; i++ {
		builder.WriteString(filler)
	}
	builder.WriteString(text)
	for i := 0; i < rightpad_length; i++ {
		builder.WriteString(filler)
	}
}

func (f *asciiFrame) Draw() string {
	const (
		body_splitter   string = "|"
		border_splitter string = "+"
		border_body     string = "-"
		arrow_tip_left  string = "<-"
		arrow_body      string = "-"
		arrow_tip_right string = "->"
		index_body      string = " "
		arrow_splitter  string = " "
	)

	type Entry struct {
		index     string
		name      string
		width     string
		minLength int
	}

	leftpad := strings.Repeat(" ", f.leftpad)

	entries := make([]Entry, len(f.fields))

	for i := range entries {
		field := &f.fields[i]

		if f.layout == AsciiFrameUnitLayout_RightToLeft {
			field = &f.fields[len(f.fields)-i-1]
		}

		entry := &entries[i]

		entry.index = fmt.Sprintf("%v", field.Begin)

		if f.layout == AsciiFrameUnitLayout_RightToLeft {
			entry.index = fmt.Sprintf("%v", field.TopUnit())
		}

		entry.name = fmt.Sprintf(" %v ", field.Name)
		entry.width = fmt.Sprintf(" %v %v ", field.Width, f.unit)
		entry.minLength = Max([]int{len(entry.index), len(entry.name), len(arrow_tip_left) + len(entry.width) + len(arrow_tip_right)})
	}

	var indices_row strings.Builder
	var header_row strings.Builder
	var body_row strings.Builder
	var footer_row strings.Builder
	var widths_row strings.Builder

	indices_row.WriteString(leftpad)
	header_row.WriteString(leftpad)
	body_row.WriteString(leftpad)
	footer_row.WriteString(leftpad)
	widths_row.WriteString(leftpad)

	for _, entry := range entries {
		indices_row.WriteString(entry.index)
		indices_row.WriteString(strings.Repeat(index_body, (entry.minLength-len(entry.index)+1)/len(index_body)))
		header_row.WriteString(border_splitter)
		header_row.WriteString(strings.Repeat(border_body, entry.minLength/len(border_body)))
		body_row.WriteString(body_splitter)
		writeRow(entry.name, 0, " ", entry.minLength, &body_row)
		footer_row.WriteString(border_splitter)
		footer_row.WriteString(strings.Repeat(border_body, entry.minLength/len(border_body)))
		widths_row.WriteString(arrow_splitter)
		widths_row.WriteString(arrow_tip_left)
		writeRow(entry.width, len(arrow_tip_left)+len(arrow_tip_right), arrow_body, entry.minLength, &widths_row)
		widths_row.WriteString(arrow_tip_right)
	}

	if f.layout == AsciiFrameUnitLayout_LeftToRight {
		indices_row.WriteString(fmt.Sprint(f.TopUnit()))
	} else {
		indices_row.WriteString("0")
	}

	header_row.WriteString(border_splitter)
	body_row.WriteString(body_splitter)
	footer_row.WriteString(border_splitter)
	widths_row.WriteString(" ")

	var result strings.Builder

	result.WriteString(indices_row.String())
	result.WriteString("\n")
	result.WriteString(header_row.String())
	result.WriteString("\n")
	result.WriteString(body_row.String())
	result.WriteString("\n")
	result.WriteString(footer_row.String())
	result.WriteString("\n")
	result.WriteString(widths_row.String())
	result.WriteString("\n")

	return result.String()
}

func fillAsciiFrameGaps(fields []AsciiFrameField, frameWidth int) []AsciiFrameField {
	result := make([]AsciiFrameField, 0, len(fields))
	currentUnit := 0

	for _, field := range fields {
		if field.Begin > currentUnit {
			result = append(result, AsciiFrameField{
				Name:  "(unused)",
				Begin: currentUnit,
				Width: field.Begin - currentUnit,
			})
		} else if field.Begin < currentUnit {
			panic("make sure fields are sorted by position and are not overlapping")
		}

		result = append(result, field)

		currentUnit = field.PastTopUnit()
	}

	if currentUnit < frameWidth {
		result = append(result, AsciiFrameField{
			Name:  "(unused)",
			Begin: currentUnit,
			Width: frameWidth - currentUnit,
		})
	}

	return result
}

// Prints an ascii diagram of a binary frame composed of contiguous fields of different unit lenghts
func AsciiFrame(fields []AsciiFrameField, frameWidth int, unit string, layout AsciiFrameUnitLayout, leftpad int) string {
	allFields := fillAsciiFrameGaps(fields, frameWidth)

	frame := asciiFrame{
		fields:     allFields,
		frameWidth: allFields[len(allFields)-1].PastTopUnit(),
		unit:       unit,
		leftpad:    leftpad,
		layout:     layout,
	}

	return frame.Draw()
}
