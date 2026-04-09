package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// PFloat is a float64 that always marshals with at least one decimal place (e.g., 0.0 not 0)
type PFloat float64

func (f PFloat) MarshalJSON() ([]byte, error) {
	v := float64(f)
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return json.Marshal(v)
	}
	if v == math.Floor(v) {
		return []byte(fmt.Sprintf("%.1f", v)), nil
	}
	s := strconv.FormatFloat(v, 'f', -1, 64)
	return []byte(s), nil
}

func (f *PFloat) UnmarshalJSON(data []byte) error {
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*f = PFloat(v)
	return nil
}

// PolarisDocument represents the top-level array: [MaterialSection, PieceSection, ProgramSection]
type PolarisDocument struct {
	Material MaterialSection
	Piece    PieceSection
	Program  ProgramSection
}

func (d PolarisDocument) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{d.Material, d.Piece, d.Program})
}

func (d *PolarisDocument) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 3 {
		return fmt.Errorf("expected 3 sections, got %d", len(raw))
	}
	if err := json.Unmarshal(raw[0], &d.Material); err != nil {
		return fmt.Errorf("material: %w", err)
	}
	if err := json.Unmarshal(raw[1], &d.Piece); err != nil {
		return fmt.Errorf("piece: %w", err)
	}
	if err := json.Unmarshal(raw[2], &d.Program); err != nil {
		return fmt.Errorf("program: %w", err)
	}
	return nil
}

// --- MaterialSection ---

type MaterialSection struct {
	Type    string           `json:"Type"`
	Objects []MaterialObject `json:"Objects"`
}

type MaterialObject struct {
	Code       string             `json:"Code"`
	Attributes MaterialAttributes `json:"Attributes"`
}

type MaterialAttributes struct {
	SpecificWeight        PFloat `json:"SpecificWeight"`
	MaterialTypeAttribute string `json:"MaterialTypeAttribute"`
}

// --- PieceSection ---

type PieceSection struct {
	Type    string        `json:"Type"`
	Objects []PieceObject `json:"Objects"`
}

type PieceObject struct {
	Contract    string          `json:"Contract"`
	Project     string          `json:"Project"`
	Drawing     string          `json:"Drawing"`
	Assembly    string          `json:"Assembly"`
	Part        string          `json:"Part"`
	ProfileType string          `json:"ProfileType"`
	SVGEntity   string          `json:"SVGEntity"`
	Attributes  PieceAttributes `json:"Attributes"`
	Operations  []Operation     `json:"Operations"`
}

type PieceAttributes struct {
	ExecutedQuantity         int    `json:"ExecutedQuantity"`
	TotalQuantity            int    `json:"TotalQuantity"`
	Length                   PFloat `json:"Length"`
	Width                    PFloat `json:"Width"`
	Thickness                PFloat `json:"Thickness"`
	MaterialCode             string `json:"MaterialCode"`
	GrossWeight              PFloat `json:"GrossWeight"`
	NetWeight                PFloat `json:"NetWeight"`
	IsPieceToMeasure         bool   `json:"IsPieceToMeasure"`
	IsFrequentlyManufactured bool   `json:"IsFrequentlyManufactured"`
	Notes                    string `json:"Notes"`
}

// --- Operations ---

type Operation struct {
	OperationType    string           `json:"OperationType"`
	LineNumber       int              `json:"LineNumber"`
	Level            int              `json:"Level"`
	ParentLineNumber int              `json:"ParentLineNumber"`
	ToBeSkipped      bool             `json:"ToBeSkipped"`
	Attributes       json.RawMessage  `json:"Attributes"`
	AdditionalItems  []AdditionalItem `json:"AdditionalItems"`
}

type AdditionalItem struct {
	OperationType string               `json:"OperationType"`
	Attributes    PathVertexAttributes `json:"Attributes"`
}

type PathVertexAttributes struct {
	X                       PFloat `json:"X"`
	Y                       PFloat `json:"Y"`
	Radius                  PFloat `json:"Radius"`
	InterpolationType       string `json:"InterpolationType"`
	TangentialSpeedOverride PFloat `json:"TangentialSpeedOverride"`
}

// NodeAttributes for NODE operations (Root, FixGroup)
type NodeAttributes struct {
	Name                string `json:"Name"`
	SequenceIndex       int    `json:"SequenceIndex"`
	OrderType           string `json:"OrderType"`
	Ox                  PFloat `json:"Ox"`
	Oy                  PFloat `json:"Oy"`
	TypeOX              string `json:"TypeOX"`
	TypeOY              string `json:"TypeOY"`
	NodeCustomType      int    `json:"NodeCustomType"`
	ProbeCode           string `json:"ProbeCode"`
	ProbeAllAoperations bool   `json:"ProbeAllAoperations"`
	ProbeReference      string `json:"ProbeReference"`
}

// PathMAttributes for PATHM operations (cut paths)
type PathMAttributes struct {
	Side                    string `json:"Side"`
	CompensationType        string `json:"CompensationType"`
	Depth                   PFloat `json:"Depth"`
	PitchStart              PFloat `json:"PitchStart"`
	PitchDepth              PFloat `json:"PitchDepth"`
	HeadCycle               string `json:"HeadCycle"`
	DeltaY                  PFloat `json:"DeltaY"`
	TS                      string `json:"TS"`
	BevelTopHeight          PFloat `json:"BevelTopHeight"`
	DN                      PFloat `json:"DN"`
	BevelTopAngle           PFloat `json:"BevelTopAngle"`
	COD                     string `json:"COD"`
	BevelBottomHeight       PFloat `json:"BevelBottomHeight"`
	BevelBottomAngle        PFloat `json:"BevelBottomAngle"`
	BevelPathEnumeration    int    `json:"BevelPathEnumeration"`
	XOut                    PFloat `json:"XOut"`
	YOut                    PFloat `json:"YOut"`
	ProbeTS                 string `json:"ProbeTS"`
	ProbeDN                 PFloat `json:"ProbeDN"`
	ProbeCOD                string `json:"ProbeCOD"`
	AuxiliariesCycle        string `json:"AuxiliariesCycle"`
	TangentialSpeedOverride PFloat `json:"TangentialSpeedOverride"`
	CounteractingCycle      bool   `json:"CounteractingCycle"`
	ProbeCode               string `json:"ProbeCode"`
	ProbeReference          string `json:"ProbeReference"`
}

// --- ProgramSection ---

type ProgramSection struct {
	Type    string          `json:"Type"`
	Objects []ProgramObject `json:"Objects"`
}

type ProgramObject struct {
	Name          string            `json:"Name"`
	ProgramType   string            `json:"ProgramType"`
	SVGEntity     string            `json:"SVGEntity"`
	Stock         *string           `json:"Stock"`
	Attributes    ProgramAttributes `json:"Attributes"`
	RelatedPieces []RelatedPiece    `json:"RelatedPieces"`
	IsoCodeLines  []interface{}     `json:"IsoCodeLines"`
}

type ProgramAttributes struct {
	TotalQuantity                     int    `json:"TotalQuantity"`
	ExecutedQuantity                  int    `json:"ExecutedQuantity"`
	ProfileType                       string `json:"ProfileType"`
	NestingLength                     PFloat `json:"NestingLength"`
	PlasmaTechnology                  string `json:"PlasmaTechnology"`
	Length                            PFloat `json:"Length"`
	WorkingArea                       string `json:"WorkingArea"`
	DestinationCode                   int    `json:"DestinationCode"`
	PlateXPosition                    PFloat `json:"PlateXPosition"`
	NestingScrap                      PFloat `json:"NestingScrap"`
	FreeStock                         PFloat `json:"FreeStock"`
	CarriageType                      string `json:"CarriageType"`
	BatchNumber                       int    `json:"BatchNumber"`
	PincherAPosition                  PFloat `json:"PincherAPosition"`
	PackageNumber                     int    `json:"PackageNumber"`
	PincherBPosition                  PFloat `json:"PincherBPosition"`
	SandblustingCode                  int    `json:"SandblustingCode"`
	PincherCPosition                  PFloat `json:"PincherCPosition"`
	ProcessingModality                string `json:"ProcessingModality"`
	DrillingToolSelection             string `json:"DrillingToolSelection"`
	FinalCutModality                  bool   `json:"FinalCutModality"`
	ScribingMarkingType               string `json:"ScribingMarkingType"`
	PreHolesOrderModality             string `json:"PreHolesOrderModality"`
	ChipRemovalModality               string `json:"ChipRemovalModality"`
	HolesNumberBrushing               int    `json:"HolesNumberBrushing"`
	ProbingDistance                    PFloat `json:"ProbingDistance"`
	MultipleTool                      bool   `json:"MultipleTool"`
	SurplusType                       string `json:"SurplusType"`
	CutsType                          string `json:"CutsType"`
	LeadingEdgeScrap                  PFloat `json:"LeadingEdgeScrap"`
	SeparationCutsType                string `json:"SeparationCutsType"`
	FinalEdgeScrap                    PFloat `json:"FinalEdgeScrap"`
	Notes                             string `json:"Notes"`
	CutScrap                          PFloat `json:"CutScrap"`
	CustomerName                      string `json:"CustomerName"`
	SawingRotationScrap               PFloat `json:"SawingRotationScrap"`
	PLMVersion                        string `json:"PLMVersion"`
	MaxPercentageScrap                PFloat `json:"MaxPercentageScrap"`
	UnloadingWidth                    PFloat `json:"UnloadingWidth"`
	MinRequiredConsole                string `json:"MinRequiredConsole"`
	SurplusEnableStock                bool   `json:"SurplusEnableStock"`
	SurplusMinLengthStock             PFloat `json:"SurplusMinLengthStock"`
	SurplusMarkingText1               string `json:"SurplusMarkingText1"`
	SurplusMarkingText2               string `json:"SurplusMarkingText2"`
	SurplusMarkingText3               string `json:"SurplusMarkingText3"`
	SurplusMarkingText4               string `json:"SurplusMarkingText4"`
	ProgramPlasmaCurrent              PFloat `json:"ProgramPlasmaCurrent"`
	PlasmaMarkingType                 string `json:"PlasmaMarkingType"`
	NozzleShape                       string `json:"NozzleShape"`
	PlasmaGasCut                      string `json:"PlasmaGasCut"`
	ShieldGasCut                      string `json:"ShieldGasCut"`
	GasMarking                        string `json:"GasMarking"`
	SimulatedCuts                     bool   `json:"SimulatedCuts"`
	TransverseSectionsProfile         bool   `json:"TransverseSectionsProfile"`
	DoublePassSlantedCutsFlanges      bool   `json:"DoublePassSlantedCutsFlanges"`
	DoublePassTubeProfile             bool   `json:"DoublePassTubeProfile"`
	AutomaticReprocessing             bool   `json:"AutomaticReprocessing"`
	AutomaticPiecesQuantityProcessing bool   `json:"AutomaticPiecesQuantityProcessing"`
	DrillsCoolingLimit                PFloat `json:"DrillsCoolingLimit"`
	BatchTotalNumber                  int    `json:"BatchTotalNumber"`
	PuFlangesPosition                 string `json:"PuFlangesPosition"`
	RequestedStationNumber            int    `json:"RequestedStationNumber"`
	RequestedHeatNumber               string `json:"RequestedHeatNumber"`
	RequestedSupplier                 string `json:"RequestedSupplier"`
	ProfileCode                       string `json:"ProfileCode"`
	MaterialCode                      string `json:"MaterialCode"`
}

type RelatedPiece struct {
	Attributes RelatedPieceAttributes `json:"Attributes"`
	Contract   string                 `json:"Contract"`
	Project    string                 `json:"Project"`
	Drawing    string                 `json:"Drawing"`
	Assembly   string                 `json:"Assembly"`
	Part       string                 `json:"Part"`
	GroupId    int                    `json:"GroupId"`
}

type RelatedPieceAttributes struct {
	OffsetX              PFloat `json:"OffsetX"`
	OffsetY              PFloat `json:"OffsetY"`
	OffsetAngle          PFloat `json:"OffsetAngle"`
	RepetitionsNumber    int    `json:"RepetitionsNumber"`
	LongitudinalRotation bool   `json:"LongitudinalRotation"`
	TransverseRotation   bool   `json:"TransverseRotation"`
	InitialDelta         PFloat `json:"InitialDelta"`
	FinalDelta           PFloat `json:"FinalDelta"`
	UnloadingCodeArea    int    `json:"UnloadingCodeArea"`
	Lot                  string `json:"Lot"`
	LotNumber            int    `json:"LotNumber"`
	LotTotalNumber       int    `json:"LotTotalNumber"`
	ExtraProcessingCode  int    `json:"ExtraProcessingCode"`
	GeneratedBarName     string `json:"GeneratedBarName"`
	DestinationCode      int    `json:"DestinationCode"`
}

// --- SVGEntity sub-structures (parsed from JSON string inside SVGEntity field) ---

type SVGEntityData struct {
	PlateContourSections []SVGPathOp  `json:"plateContourSections"`
	NestedParts          []NestedPart `json:"nestedParts"`
}

type SVGPathOp struct {
	Operation    string  `json:"operation"`
	X            *PFloat `json:"x,omitempty"`
	Y            *PFloat `json:"y,omitempty"`
	RX           *PFloat `json:"rx,omitempty"`
	RY           *PFloat `json:"ry,omitempty"`
	ThetaX       *PFloat `json:"thetax,omitempty"`
	LargeArcFlag *int    `json:"largeArcFlag,omitempty"`
	SweepFlag    *int    `json:"sweepFlag,omitempty"`
}

type NestedPart struct {
	X    PFloat     `json:"x"`
	Y    PFloat     `json:"y"`
	Part PartDetail `json:"part"`
}

type PartDetail struct {
	Name          string      `json:"name"`
	Project       string      `json:"project"`
	Index         int         `json:"index"`
	Segments      []SVGPathOp `json:"segments"`
	InnerSegments []SVGPathOp `json:"innnerSegments"`
}

// --- Input parameters for creating new programs ---

type ShapeParams struct {
	ShapeType    string // "CAN" or "CONE"
	Contract     string
	Project      string
	Drawing      string
	Part         string
	Length       float64
	Width        float64
	Thickness    float64
	MaterialCode string
	WorkingArea  string
	// CONE-specific
	PieceHeight  float64
	LeftOffset   float64
	BottomRadius float64
	TopRadius    float64
}
