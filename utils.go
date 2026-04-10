package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var scanner *bufio.Scanner

func initScanner() {
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
}

func promptString(prompt string, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	scanner.Scan()
	val := strings.TrimSpace(scanner.Text())
	if val == "" {
		return defaultVal
	}
	return val
}

func promptFloat(prompt string, defaultVal float64) float64 {
	for {
		var defStr string
		if defaultVal != 0 {
			defStr = strconv.FormatFloat(defaultVal, 'f', -1, 64)
		}
		s := promptString(prompt, defStr)
		if s == "" {
			return defaultVal
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			fmt.Println("  Nieprawidlowa wartosc, sprobuj ponownie.")
			continue
		}
		return v
	}
}

func promptFloatAllowZero(prompt string) float64 {
	for {
		s := promptString(prompt, "0")
		if s == "" || s == "0" {
			return 0
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			fmt.Println("  Nieprawidlowa wartosc, sprobuj ponownie.")
			continue
		}
		return v
	}
}

// calculateGrossWeight computes weight from bounding rectangle dimensions
// dimensions in mm, specificWeight in g/cm3, result in kg
func calculateGrossWeight(length, width, thickness, specificWeight float64) float64 {
	return roundTo2(length * width * thickness * specificWeight / 1e6)
}

func marshalNodeAttrs(name string, seqIndex int, orderType string, customType int) json.RawMessage {
	attrs := NodeAttributes{
		Name:                name,
		SequenceIndex:       seqIndex,
		OrderType:           orderType,
		Ox:                  0.0,
		Oy:                  0.0,
		TypeOX:              "Standard",
		TypeOY:              "Standard",
		NodeCustomType:      customType,
		ProbeCode:           "0",
		ProbeAllAoperations: false,
		ProbeReference:      "None",
	}
	data, _ := json.Marshal(attrs)
	return json.RawMessage(data)
}

func marshalPathMAttrs(bevelPathEnum int) json.RawMessage {
	attrs := PathMAttributes{
		Side:                    "C",
		CompensationType:        "None",
		Depth:                   0.0,
		PitchStart:              0.0,
		PitchDepth:              0.0,
		HeadCycle:               "SingleHead",
		DeltaY:                  0.0,
		TS:                      "POCKET_75",
		BevelTopHeight:          0.0,
		DN:                      0.0,
		BevelTopAngle:           0.0,
		COD:                     "",
		BevelBottomHeight:       0.0,
		BevelBottomAngle:        0.0,
		BevelPathEnumeration:    bevelPathEnum,
		XOut:                    0.0,
		YOut:                    0.0,
		ProbeTS:                 "NotDefined",
		ProbeDN:                 0.0,
		ProbeCOD:                "",
		AuxiliariesCycle:        "Active",
		TangentialSpeedOverride: 0.0,
		CounteractingCycle:      false,
		ProbeCode:               "0",
		ProbeReference:          "None",
	}
	data, _ := json.Marshal(attrs)
	return json.RawMessage(data)
}

func makePathVertex(x, y, radius float64) AdditionalItem {
	return AdditionalItem{
		OperationType: "PATHVERTEX",
		Attributes: PathVertexAttributes{
			X:                       PFloat(x),
			Y:                       PFloat(y),
			Radius:                  PFloat(radius),
			InterpolationType:       "None",
			TangentialSpeedOverride: 0.0,
		},
	}
}

func buildDefaultProgramAttrs(length float64, workingArea string) ProgramAttributes {
	return ProgramAttributes{
		TotalQuantity:                     1,
		ExecutedQuantity:                  0,
		ProfileType:                       "P",
		NestingLength:                     0.0,
		PlasmaTechnology:                  "NotSpecified",
		Length:                            PFloat(length),
		WorkingArea:                       workingArea,
		DestinationCode:                   0,
		PlateXPosition:                    0.0,
		NestingScrap:                      0.0,
		FreeStock:                         0.0,
		CarriageType:                      "VerticalClamp",
		BatchNumber:                       0,
		PincherAPosition:                  0.0,
		PackageNumber:                     0,
		PincherBPosition:                  0.0,
		SandblustingCode:                  0,
		PincherCPosition:                  0.0,
		ProcessingModality:                "0",
		DrillingToolSelection:             "ByMaxLife",
		FinalCutModality:                  false,
		ScribingMarkingType:               "Standard",
		PreHolesOrderModality:             "GroupAhead",
		ChipRemovalModality:               "None",
		HolesNumberBrushing:               0,
		ProbingDistance:                    0.0,
		MultipleTool:                      false,
		SurplusType:                       "FinalEdge",
		CutsType:                          "NotDefined",
		LeadingEdgeScrap:                  0.0,
		SeparationCutsType:                "NotDefined",
		FinalEdgeScrap:                    0.0,
		Notes:                             "",
		CutScrap:                          0.0,
		CustomerName:                      "",
		SawingRotationScrap:               0.0,
		PLMVersion:                        "2025 build 33",
		MaxPercentageScrap:                0.0,
		UnloadingWidth:                    0.0,
		MinRequiredConsole:                "NotSpecified",
		SurplusEnableStock:                false,
		SurplusMinLengthStock:             0.0,
		SurplusMarkingText1:               "",
		SurplusMarkingText2:               "",
		SurplusMarkingText3:               "",
		SurplusMarkingText4:               "",
		ProgramPlasmaCurrent:              0.0,
		PlasmaMarkingType:                 "Standard",
		NozzleShape:                       "Bevel",
		PlasmaGasCut:                      "NotSelected",
		ShieldGasCut:                      "NotSelected",
		GasMarking:                        "NotSelected",
		SimulatedCuts:                     false,
		TransverseSectionsProfile:         false,
		DoublePassSlantedCutsFlanges:      false,
		DoublePassTubeProfile:             false,
		AutomaticReprocessing:             false,
		AutomaticPiecesQuantityProcessing: false,
		DrillsCoolingLimit:                0.0,
		BatchTotalNumber:                  0,
		PuFlangesPosition:                 "FlangesDownward",
		RequestedStationNumber:            1,
		RequestedHeatNumber:               "",
		RequestedSupplier:                 "",
		ProfileCode:                       "",
		MaterialCode:                      "",
	}
}
