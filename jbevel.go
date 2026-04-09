package main

import (
	"fmt"
	"strings"
)

// ApplyJBevel creates a J-BEVEL variant of an existing program
// It modifies the part name, recalculates geometry with new dimensions, and saves a new file
func ApplyJBevel(doc *PolarisDocument, bevelNum string, newLength, newWidth, newThickness float64, workingArea string) (string, error) {
	if len(doc.Piece.Objects) == 0 || len(doc.Program.Objects) == 0 {
		return "", fmt.Errorf("invalid program: missing piece or program objects")
	}

	piece := &doc.Piece.Objects[0]
	program := &doc.Program.Objects[0]

	oldLength := float64(piece.Attributes.Length)
	oldWidth := float64(piece.Attributes.Width)

	// Build new part name with J-BEVEL suffix
	// e.g., POS_296_R01 -> POS_296_R_J-BEVEL-2
	basePart := piece.Part
	// Remove existing J-BEVEL suffix if present
	if idx := strings.Index(basePart, "_J-BEVEL"); idx >= 0 {
		basePart = basePart[:idx]
	}
	newPart := fmt.Sprintf("%s_J-BEVEL-%s", basePart, bevelNum)

	// Detect shape type
	isCone := detectCone(piece.Operations)

	// Update piece part name
	piece.Part = newPart

	// Update dimensions and geometry
	if isCone {
		// CONE: proportional scaling
		scaleX := newLength / oldLength
		scaleY := newWidth / oldWidth

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
		piece.SVGEntity = updateSvgEntityDimensions(piece.SVGEntity, oldLength, oldWidth, newLength, newWidth, true)
	} else {
		// CAN: direct coordinate replacement
		for i := range piece.Operations {
			if piece.Operations[i].OperationType == "PATHM" {
				for j := range piece.Operations[i].AdditionalItems {
					pv := &piece.Operations[i].AdditionalItems[j].Attributes
					if floatClose(float64(pv.X), oldLength) {
						pv.X = PFloat(newLength)
					}
					if floatClose(float64(pv.Y), oldWidth) {
						pv.Y = PFloat(newWidth)
					}
				}
			}
		}
		piece.SVGEntity = updateSvgEntityDimensions(piece.SVGEntity, oldLength, oldWidth, newLength, newWidth, false)
	}

	// Update piece attributes
	piece.Attributes.Length = PFloat(newLength)
	piece.Attributes.Width = PFloat(newWidth)
	piece.Attributes.Thickness = PFloat(newThickness)
	piece.Attributes.ExecutedQuantity = 0

	// Recalculate weights
	specificWeight := 7.85
	if len(doc.Material.Objects) > 0 {
		specificWeight = float64(doc.Material.Objects[0].Attributes.SpecificWeight)
	}
	newGross := calculateGrossWeight(newLength, newWidth, newThickness, specificWeight)
	piece.Attributes.GrossWeight = PFloat(newGross)
	if isCone {
		oldGross := calculateGrossWeight(oldLength, oldWidth, float64(piece.Attributes.Thickness), specificWeight)
		if oldGross > 0 {
			ratio := float64(piece.Attributes.NetWeight) / oldGross
			piece.Attributes.NetWeight = PFloat(newGross * ratio)
		}
	} else {
		piece.Attributes.NetWeight = PFloat(newGross)
	}

	// Update program object
	programName := fmt.Sprintf("PGM_%s_%s_%s", piece.Contract, newPart, workingArea)
	program.Name = programName
	program.Attributes.Length = PFloat(newLength)
	program.Attributes.WorkingArea = workingArea
	program.Attributes.PlateXPosition = 0.0
	program.Attributes.ExecutedQuantity = 0

	// Update ProgramObject SVGEntity
	if !strings.Contains(program.SVGEntity, "error") {
		program.SVGEntity = piece.SVGEntity
	}

	// Update RelatedPieces
	for i := range program.RelatedPieces {
		program.RelatedPieces[i].Part = newPart
	}

	// Save
	fileName := programName + ".Program.polaris"
	if err := WritePolarisFile(fileName, doc); err != nil {
		return "", err
	}

	return fileName, nil
}
