package main

import (
	"encoding/json"
	"fmt"
	"math"
)

// GenerateCONE creates a PolarisDocument for a CONE (banana-shaped plate)
func GenerateCONE(p ShapeParams) (*PolarisDocument, string) {
	length := p.Length
	width := p.Width
	thickness := p.Thickness
	pieceHeight := p.PieceHeight
	leftOffset := p.LeftOffset
	bottomRadius := p.BottomRadius
	topRadius := p.TopRadius
	specificWeight := 7.85

	rightBottomX := length - leftOffset // symmetric

	grossWeight := calculateGrossWeight(length, width, thickness, specificWeight)
	netWeight := calculateConeNetWeight(length, pieceHeight, leftOffset, leftOffset, bottomRadius, topRadius, thickness, specificWeight)

	svgEntity := buildCONESvgEntity(length, width, pieceHeight, leftOffset, rightBottomX, bottomRadius, topRadius)
	operations := buildCONEOperations(length, pieceHeight, leftOffset, rightBottomX, bottomRadius, topRadius)

	piece := PieceObject{
		Contract:    p.Contract,
		Project:     p.Project,
		Drawing:     p.Drawing,
		Assembly:    "",
		Part:        p.Part,
		ProfileType: "P",
		SVGEntity:   svgEntity,
		Attributes: PieceAttributes{
			ExecutedQuantity:         0,
			TotalQuantity:           1,
			Length:                   PFloat(length),
			Width:                    PFloat(width),
			Thickness:                PFloat(thickness),
			MaterialCode:            p.MaterialCode,
			GrossWeight:             PFloat(grossWeight),
			NetWeight:               PFloat(netWeight),
			IsPieceToMeasure:        true,
			IsFrequentlyManufactured: false,
			Notes:                   "",
		},
		Operations: operations,
	}

	programName := fmt.Sprintf("PGM_%s_%s_%s", p.Contract, p.Part, p.WorkingArea)
	fileName := programName + ".Program.polaris"

	progAttrs := buildDefaultProgramAttrs(length, p.WorkingArea)

	program := ProgramObject{
		Name:        programName,
		ProgramType: "PieceToMeasure",
		SVGEntity:   `{"error":"unknown profile type" }`,
		Stock:       nil,
		Attributes:  progAttrs,
		RelatedPieces: []RelatedPiece{
			{
				Attributes: RelatedPieceAttributes{
					RepetitionsNumber: 1,
				},
				Contract: p.Contract,
				Project:  p.Project,
				Drawing:  p.Drawing,
				Assembly: "",
				Part:     p.Part,
				GroupId:  1,
			},
		},
		IsoCodeLines: []interface{}{},
	}

	doc := &PolarisDocument{
		Material: MaterialSection{
			Type: "MaterialObject",
			Objects: []MaterialObject{
				{
					Code: p.MaterialCode,
					Attributes: MaterialAttributes{
						SpecificWeight:        PFloat(specificWeight),
						MaterialTypeAttribute: "MildSteel",
					},
				},
			},
		},
		Piece: PieceSection{
			Type:    "PieceObject",
			Objects: []PieceObject{piece},
		},
		Program: ProgramSection{
			Type:    "ProgramObject",
			Objects: []ProgramObject{program},
		},
	}

	return doc, fileName
}

func buildCONESvgEntity(length, width, pieceHeight, leftOffset, rightBottomX, bottomRadius, topRadius float64) string {
	svg := SVGEntityData{
		PlateContourSections: []SVGPathOp{
			{Operation: "M", X: pfloatPtr(0.0), Y: pfloatPtr(0.0)},
			{Operation: "L", X: pfloatPtr(length), Y: pfloatPtr(0.0)},
			{Operation: "L", X: pfloatPtr(length), Y: pfloatPtr(width)},
			{Operation: "L", X: pfloatPtr(0.0), Y: pfloatPtr(width)},
			{Operation: "Z"},
		},
		NestedParts: []NestedPart{
			// P1: left diagonal
			{
				X: 0.001, Y: 0.001,
				Part: PartDetail{
					Name: "P1", Project: "", Index: 1,
					Segments: []SVGPathOp{
						{Operation: "M", X: pfloatPtr(0.0), Y: pfloatPtr(pieceHeight)},
						{Operation: "L", X: pfloatPtr(leftOffset), Y: pfloatPtr(0.0)},
					},
					InnerSegments: []SVGPathOp{},
				},
			},
			// P2: right diagonal
			{
				X: 0.001, Y: 0.001,
				Part: PartDetail{
					Name: "P2", Project: "", Index: 2,
					Segments: []SVGPathOp{
						{Operation: "M", X: pfloatPtr(rightBottomX), Y: pfloatPtr(0.0)},
						{Operation: "L", X: pfloatPtr(length), Y: pfloatPtr(pieceHeight)},
					},
					InnerSegments: []SVGPathOp{},
				},
			},
			// P3: bottom arc
			{
				X: 0.001, Y: 0.001,
				Part: PartDetail{
					Name: "P3", Project: "", Index: 3,
					Segments: []SVGPathOp{
						{Operation: "M", X: pfloatPtr(leftOffset), Y: pfloatPtr(0.0)},
						{
							Operation:    "A",
							X:            pfloatPtr(rightBottomX),
							Y:            pfloatPtr(0.0),
							RX:           pfloatPtr(bottomRadius),
							RY:           pfloatPtr(bottomRadius),
							ThetaX:       pfloatPtr(0.0),
							LargeArcFlag: intPtr(0),
							SweepFlag:    intPtr(0),
						},
					},
					InnerSegments: []SVGPathOp{},
				},
			},
			// P4: top arc
			{
				X: 0.001, Y: 0.001,
				Part: PartDetail{
					Name: "P4", Project: "", Index: 4,
					Segments: []SVGPathOp{
						{Operation: "M", X: pfloatPtr(length), Y: pfloatPtr(pieceHeight)},
						{
							Operation:    "A",
							X:            pfloatPtr(0.0),
							Y:            pfloatPtr(pieceHeight),
							RX:           pfloatPtr(topRadius),
							RY:           pfloatPtr(topRadius),
							ThetaX:       pfloatPtr(0.0),
							LargeArcFlag: intPtr(0),
							SweepFlag:    intPtr(1),
						},
					},
					InnerSegments: []SVGPathOp{},
				},
			},
		},
	}

	data, _ := json.Marshal(svg)
	return string(data)
}

func buildCONEOperations(length, pieceHeight, leftOffset, rightBottomX, bottomRadius, topRadius float64) []Operation {
	return []Operation{
		{
			OperationType: "NODE", LineNumber: 0, Level: 0, ParentLineNumber: 0, ToBeSkipped: false,
			Attributes:      marshalNodeAttrs("Root", 0, "NotDefined", 0),
			AdditionalItems: []AdditionalItem{},
		},
		{
			OperationType: "NODE", LineNumber: 1, Level: 1, ParentLineNumber: 0, ToBeSkipped: false,
			Attributes:      marshalNodeAttrs("FixGroup 0", 1, "SingleSide", 1),
			AdditionalItems: []AdditionalItem{},
		},
		// P1 - left diagonal
		{
			OperationType: "PATHM", LineNumber: 2, Level: 2, ParentLineNumber: 1, ToBeSkipped: false,
			Attributes: marshalPathMAttrs(-1),
			AdditionalItems: []AdditionalItem{
				makePathVertex(0.0, pieceHeight, 0.0),
				makePathVertex(leftOffset, 0.0, 0.0),
			},
		},
		// P3 - bottom arc (negative radius)
		{
			OperationType: "PATHM", LineNumber: 3, Level: 2, ParentLineNumber: 1, ToBeSkipped: false,
			Attributes: marshalPathMAttrs(1),
			AdditionalItems: []AdditionalItem{
				makePathVertex(leftOffset, 0.0, 0.0),
				makePathVertex(rightBottomX, 0.0, -bottomRadius),
			},
		},
		// P2 - right diagonal
		{
			OperationType: "PATHM", LineNumber: 4, Level: 2, ParentLineNumber: 1, ToBeSkipped: false,
			Attributes: marshalPathMAttrs(-1),
			AdditionalItems: []AdditionalItem{
				makePathVertex(rightBottomX, 0.0, 0.0),
				makePathVertex(length, pieceHeight, 0.0),
			},
		},
		// P4 - top arc (positive radius)
		{
			OperationType: "PATHM", LineNumber: 5, Level: 2, ParentLineNumber: 1, ToBeSkipped: false,
			Attributes: marshalPathMAttrs(1),
			AdditionalItems: []AdditionalItem{
				makePathVertex(length, pieceHeight, 0.0),
				makePathVertex(0.0, pieceHeight, topRadius),
			},
		},
	}
}

func calculateConeNetWeight(length, pieceHeight, leftOffset, rightOffset, bottomRadius, topRadius, thickness, specificWeight float64) float64 {
	bottomChord := length - leftOffset - rightOffset
	topChord := length

	trapArea := (bottomChord + topChord) / 2.0 * pieceHeight

	bottomSegArea := circularSegmentArea(bottomChord, bottomRadius)
	topSegArea := circularSegmentArea(topChord, topRadius)

	netArea := trapArea - bottomSegArea + topSegArea

	return netArea * thickness * specificWeight / 1e6
}

func circularSegmentArea(chord, radius float64) float64 {
	if radius <= 0 || chord <= 0 || chord >= 2*radius {
		return 0
	}
	halfChord := chord / 2.0
	h := radius - math.Sqrt(radius*radius-halfChord*halfChord)
	area := radius*radius*math.Acos((radius-h)/radius) - (radius-h)*math.Sqrt(2*radius*h-h*h)
	return area
}
