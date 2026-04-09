package main

import (
	"encoding/json"
	"fmt"
)

// GenerateCAN creates a PolarisDocument for a CAN (flat rectangular plate)
func GenerateCAN(p ShapeParams) (*PolarisDocument, string) {
	length := p.Length
	width := p.Width
	thickness := p.Thickness
	specificWeight := 7.85

	grossWeight := calculateGrossWeight(length, width, thickness, specificWeight)

	svgEntity := buildCANSvgEntity(length, width)
	operations := buildCANOperations(length, width)

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
			NetWeight:               PFloat(grossWeight),
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
		SVGEntity:   svgEntity,
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

func buildCANSvgEntity(length, width float64) string {
	svg := SVGEntityData{
		PlateContourSections: []SVGPathOp{
			{Operation: "M", X: pfloatPtr(0.0), Y: pfloatPtr(0.0)},
			{Operation: "L", X: pfloatPtr(length), Y: pfloatPtr(0.0)},
			{Operation: "L", X: pfloatPtr(length), Y: pfloatPtr(width)},
			{Operation: "L", X: pfloatPtr(0.0), Y: pfloatPtr(width)},
			{Operation: "Z"},
		},
		NestedParts: []NestedPart{
			{
				X: 0.001, Y: 0.001,
				Part: PartDetail{
					Name: "P1", Project: "", Index: 1,
					Segments: []SVGPathOp{
						{Operation: "M", X: pfloatPtr(0.0), Y: pfloatPtr(width)},
						{Operation: "L", X: pfloatPtr(0.0), Y: pfloatPtr(0.0)},
					},
					InnerSegments: []SVGPathOp{},
				},
			},
			{
				X: 0.001, Y: 0.001,
				Part: PartDetail{
					Name: "P2", Project: "", Index: 2,
					Segments: []SVGPathOp{
						{Operation: "M", X: pfloatPtr(0.0), Y: pfloatPtr(0.0)},
						{Operation: "L", X: pfloatPtr(length), Y: pfloatPtr(0.0)},
					},
					InnerSegments: []SVGPathOp{},
				},
			},
			{
				X: 0.001, Y: 0.001,
				Part: PartDetail{
					Name: "P3", Project: "", Index: 3,
					Segments: []SVGPathOp{
						{Operation: "M", X: pfloatPtr(length), Y: pfloatPtr(0.0)},
						{Operation: "L", X: pfloatPtr(length), Y: pfloatPtr(width)},
					},
					InnerSegments: []SVGPathOp{},
				},
			},
			{
				X: 0.001, Y: 0.001,
				Part: PartDetail{
					Name: "P4", Project: "", Index: 4,
					Segments: []SVGPathOp{
						{Operation: "M", X: pfloatPtr(length), Y: pfloatPtr(width)},
						{Operation: "L", X: pfloatPtr(0.0), Y: pfloatPtr(width)},
					},
					InnerSegments: []SVGPathOp{},
				},
			},
		},
	}

	data, _ := json.Marshal(svg)
	return string(data)
}

func buildCANOperations(length, width float64) []Operation {
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
		// P1 - left edge
		{
			OperationType: "PATHM", LineNumber: 2, Level: 2, ParentLineNumber: 1, ToBeSkipped: false,
			Attributes: marshalPathMAttrs(-1),
			AdditionalItems: []AdditionalItem{
				makePathVertex(0.0, width, 0.0),
				makePathVertex(0.0, 0.0, 0.0),
			},
		},
		// P2 - bottom edge
		{
			OperationType: "PATHM", LineNumber: 3, Level: 2, ParentLineNumber: 1, ToBeSkipped: false,
			Attributes: marshalPathMAttrs(1),
			AdditionalItems: []AdditionalItem{
				makePathVertex(0.0, 0.0, 0.0),
				makePathVertex(length, 0.0, 0.0),
			},
		},
		// P3 - right edge
		{
			OperationType: "PATHM", LineNumber: 4, Level: 2, ParentLineNumber: 1, ToBeSkipped: false,
			Attributes: marshalPathMAttrs(-1),
			AdditionalItems: []AdditionalItem{
				makePathVertex(length, 0.0, 0.0),
				makePathVertex(length, width, 0.0),
			},
		},
		// P4 - top edge
		{
			OperationType: "PATHM", LineNumber: 5, Level: 2, ParentLineNumber: 1, ToBeSkipped: false,
			Attributes: marshalPathMAttrs(1),
			AdditionalItems: []AdditionalItem{
				makePathVertex(length, width, 0.0),
				makePathVertex(0.0, width, 0.0),
			},
		},
	}
}

func pfloatPtr(v float64) *PFloat {
	pf := PFloat(v)
	return &pf
}

func intPtr(v int) *int {
	return &v
}
