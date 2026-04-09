package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// ApplyDeviation loads a program, applies length/width deviations, and saves a new file
func ApplyDeviation(inputFile string, devLength, devWidth float64) (string, error) {
	doc, err := ReadPolarisFile(inputFile)
	if err != nil {
		return "", err
	}

	if len(doc.Piece.Objects) == 0 {
		return "", fmt.Errorf("no piece objects found in file")
	}
	if len(doc.Program.Objects) == 0 {
		return "", fmt.Errorf("no program objects found in file")
	}

	piece := &doc.Piece.Objects[0]
	program := &doc.Program.Objects[0]

	oldLength := float64(piece.Attributes.Length)
	oldWidth := float64(piece.Attributes.Width)
	newLength := oldLength + devLength
	newWidth := oldWidth + devWidth

	if newLength <= 0 || newWidth <= 0 {
		return "", fmt.Errorf("invalid dimensions after deviation: length=%.2f, width=%.2f", newLength, newWidth)
	}

	isCone := detectCone(piece.Operations)

	if isCone {
		applyConeDeviation(piece, program, oldLength, oldWidth, newLength, newWidth)
	} else {
		applyCanDeviation(piece, program, oldLength, oldWidth, newLength, newWidth)
	}

	// Update weights
	specificWeight := 7.85
	if len(doc.Material.Objects) > 0 {
		specificWeight = float64(doc.Material.Objects[0].Attributes.SpecificWeight)
	}
	thickness := float64(piece.Attributes.Thickness)

	newGross := calculateGrossWeight(newLength, newWidth, thickness, specificWeight)
	if isCone {
		oldGross := calculateGrossWeight(oldLength, oldWidth, thickness, specificWeight)
		if oldGross > 0 {
			ratio := float64(piece.Attributes.NetWeight) / oldGross
			piece.Attributes.NetWeight = PFloat(newGross * ratio)
		}
	} else {
		piece.Attributes.NetWeight = PFloat(newGross)
	}
	piece.Attributes.GrossWeight = PFloat(newGross)

	outputFile := generateDeviationFilename(inputFile, devLength, devWidth)

	if err := WritePolarisFile(outputFile, doc); err != nil {
		return "", err
	}

	return outputFile, nil
}

func detectCone(operations []Operation) bool {
	for _, op := range operations {
		if op.OperationType == "PATHM" {
			for _, item := range op.AdditionalItems {
				if float64(item.Attributes.Radius) != 0.0 {
					return true
				}
			}
		}
	}
	return false
}

func applyCanDeviation(piece *PieceObject, program *ProgramObject, oldL, oldW, newL, newW float64) {
	piece.Attributes.Length = PFloat(newL)
	piece.Attributes.Width = PFloat(newW)

	for i := range piece.Operations {
		if piece.Operations[i].OperationType == "PATHM" {
			for j := range piece.Operations[i].AdditionalItems {
				pv := &piece.Operations[i].AdditionalItems[j].Attributes
				if floatClose(float64(pv.X), oldL) {
					pv.X = PFloat(newL)
				}
				if floatClose(float64(pv.Y), oldW) {
					pv.Y = PFloat(newW)
				}
			}
		}
	}

	piece.SVGEntity = updateSvgEntityDimensions(piece.SVGEntity, oldL, oldW, newL, newW, false)

	program.Attributes.Length = PFloat(newL)

	if !strings.Contains(program.SVGEntity, "error") {
		program.SVGEntity = piece.SVGEntity
	}
}

func applyConeDeviation(piece *PieceObject, program *ProgramObject, oldL, oldW, newL, newW float64) {
	scaleX := newL / oldL
	scaleY := newW / oldW

	piece.Attributes.Length = PFloat(newL)
	piece.Attributes.Width = PFloat(newW)

	for i := range piece.Operations {
		if piece.Operations[i].OperationType == "PATHM" {
			for j := range piece.Operations[i].AdditionalItems {
				pv := &piece.Operations[i].AdditionalItems[j].Attributes
				pv.X = PFloat(float64(pv.X) * scaleX)
				pv.Y = PFloat(float64(pv.Y) * scaleY)
				if float64(pv.Radius) != 0 {
					pv.Radius = PFloat(scaleArcRadius(float64(pv.Radius), scaleX, scaleY))
				}
			}
		}
	}

	piece.SVGEntity = updateSvgEntityDimensions(piece.SVGEntity, oldL, oldW, newL, newW, true)

	program.Attributes.Length = PFloat(newL)
}

func scaleArcRadius(oldRadius float64, scaleX, scaleY float64) float64 {
	sign := 1.0
	r := oldRadius
	if r < 0 {
		sign = -1.0
		r = -r
	}
	if math.Abs(scaleX-scaleY) < 0.001 {
		return sign * r * scaleX
	}
	return sign * r * scaleX
}

func updateSvgEntityDimensions(svgStr string, oldL, oldW, newL, newW float64, isCone bool) string {
	var svg SVGEntityData
	if err := json.Unmarshal([]byte(svgStr), &svg); err != nil {
		return svgStr
	}

	if isCone {
		scaleX := newL / oldL
		scaleY := newW / oldW

		for i := range svg.PlateContourSections {
			s := &svg.PlateContourSections[i]
			if s.X != nil {
				*s.X = PFloat(float64(*s.X) * scaleX)
			}
			if s.Y != nil {
				*s.Y = PFloat(float64(*s.Y) * scaleY)
			}
		}

		for i := range svg.NestedParts {
			for j := range svg.NestedParts[i].Part.Segments {
				seg := &svg.NestedParts[i].Part.Segments[j]
				if seg.X != nil {
					*seg.X = PFloat(float64(*seg.X) * scaleX)
				}
				if seg.Y != nil {
					*seg.Y = PFloat(float64(*seg.Y) * scaleY)
				}
				if seg.RX != nil {
					*seg.RX = PFloat(float64(*seg.RX) * scaleX)
				}
				if seg.RY != nil {
					*seg.RY = PFloat(float64(*seg.RY) * scaleY)
				}
			}
		}
	} else {
		for i := range svg.PlateContourSections {
			s := &svg.PlateContourSections[i]
			if s.X != nil && floatClose(float64(*s.X), oldL) {
				*s.X = PFloat(newL)
			}
			if s.Y != nil && floatClose(float64(*s.Y), oldW) {
				*s.Y = PFloat(newW)
			}
		}

		for i := range svg.NestedParts {
			for j := range svg.NestedParts[i].Part.Segments {
				seg := &svg.NestedParts[i].Part.Segments[j]
				if seg.X != nil && floatClose(float64(*seg.X), oldL) {
					*seg.X = PFloat(newL)
				}
				if seg.Y != nil && floatClose(float64(*seg.Y), oldW) {
					*seg.Y = PFloat(newW)
				}
			}
		}
	}

	data, _ := json.Marshal(svg)
	return string(data)
}

func floatClose(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func generateDeviationFilename(inputFile string, devL, devW float64) string {
	base := strings.TrimSuffix(inputFile, ".Program.polaris")
	if base == inputFile {
		base = strings.TrimSuffix(inputFile, ".polaris")
	}
	suffix := fmt.Sprintf("_DEV_L%.0f_W%.0f", devL, devW)
	return base + suffix + ".Program.polaris"
}
