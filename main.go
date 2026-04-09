package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	initScanner()

	fmt.Println("========================================")
	fmt.Println("  POLARIS Program Generator v1.0")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("  1. Stworz nowy program")
	fmt.Println("  2. Zaladuj istniejacy program (dewiacja)")
	fmt.Println("  0. Wyjscie")
	fmt.Println()

	choice := promptString("Wybierz opcje", "")

	switch choice {
	case "1":
		handleCreateNew()
	case "2":
		handleLoadDeviation()
	case "0":
		fmt.Println("Do widzenia!")
		os.Exit(0)
	default:
		fmt.Println("Nieprawidlowy wybor.")
		os.Exit(1)
	}
}

func handleCreateNew() {
	fmt.Println()
	fmt.Println("--- Tworzenie nowego programu ---")
	fmt.Println()
	fmt.Println("  Typ ksztaltu:")
	fmt.Println("  1. CAN  (prosta plyta prostokatna)")
	fmt.Println("  2. CONE (plyta w ksztalcie banana)")
	fmt.Println()

	shapeChoice := promptString("Wybierz typ", "")

	params := ShapeParams{}

	switch shapeChoice {
	case "1":
		params.ShapeType = "CAN"
	case "2":
		params.ShapeType = "CONE"
	default:
		fmt.Println("Nieprawidlowy wybor typu.")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("--- Parametry podstawowe ---")
	params.Contract = promptString("Contract (np. B2149)", "")
	params.Project = promptString("Project", params.Contract)
	params.Drawing = promptString("Drawing (np. MONOPILE_PARTS_A)", "")
	params.Part = promptString("Part (np. POS_296_R01)", "")
	params.Length = promptFloat("Length [mm]", 0)
	params.Width = promptFloat("Width [mm]", 0)
	params.Thickness = promptFloat("Thickness [mm]", 80.0)
	params.MaterialCode = promptString("Material Code", "S355ML")
	params.WorkingArea = promptString("Working Area", "A")

	if params.Length <= 0 || params.Width <= 0 || params.Thickness <= 0 {
		fmt.Println("Blad: Length, Width i Thickness musza byc wieksze od 0.")
		os.Exit(1)
	}

	if params.ShapeType == "CONE" {
		fmt.Println()
		fmt.Println("--- Parametry CONE (banan) ---")
		params.PieceHeight = promptFloat("Piece Height [mm] (wysokosc ksztaltu)", 0)
		params.LeftOffset = promptFloat("Left Offset [mm] (przesuniecie lewej krawedzi)", 0)
		params.BottomRadius = promptFloat("Bottom Radius [mm] (promien dolnego luku)", 0)
		params.TopRadius = promptFloat("Top Radius [mm] (promien gornego luku)", 0)

		if params.PieceHeight <= 0 || params.LeftOffset < 0 || params.BottomRadius <= 0 || params.TopRadius <= 0 {
			fmt.Println("Blad: Nieprawidlowe parametry CONE.")
			os.Exit(1)
		}
		if params.PieceHeight > params.Width {
			fmt.Println("Blad: PieceHeight nie moze byc wiekszy niz Width (wysokosc plyty).")
			os.Exit(1)
		}
	}

	// Generate
	var doc *PolarisDocument
	var fileName string

	if params.ShapeType == "CAN" {
		doc, fileName = GenerateCAN(params)
	} else {
		doc, fileName = GenerateCONE(params)
	}

	// Save
	if err := WritePolarisFile(fileName, doc); err != nil {
		fmt.Printf("Blad zapisu: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Printf("  Program zapisany: %s\n", fileName)
	fmt.Println("========================================")
}

func handleLoadDeviation() {
	fmt.Println()
	fmt.Println("--- Ladowanie programu z dewiacja ---")
	fmt.Println()

	// List available .Program.polaris files in current directory
	files, _ := filepath.Glob("*.Program.polaris")
	if len(files) > 0 {
		fmt.Println("Dostepne pliki:")
		for i, f := range files {
			fmt.Printf("  %d. %s\n", i+1, f)
		}
		fmt.Println()
	}

	inputFile := promptString("Podaj sciezke do pliku .Program.polaris", "")
	if inputFile == "" {
		fmt.Println("Blad: nie podano pliku.")
		os.Exit(1)
	}

	// Check if user entered a number (selection from list)
	if len(files) > 0 {
		idx := 0
		fmt.Sscanf(inputFile, "%d", &idx)
		if idx >= 1 && idx <= len(files) {
			inputFile = files[idx-1]
		}
	}

	if !strings.HasSuffix(inputFile, ".Program.polaris") {
		fmt.Println("Uwaga: plik nie ma rozszerzenia .Program.polaris")
	}

	// Read and display current dimensions
	doc, err := ReadPolarisFile(inputFile)
	if err != nil {
		fmt.Printf("Blad odczytu: %v\n", err)
		os.Exit(1)
	}

	if len(doc.Piece.Objects) == 0 {
		fmt.Println("Blad: brak obiektow PieceObject w pliku.")
		os.Exit(1)
	}

	piece := doc.Piece.Objects[0]
	isCone := detectCone(piece.Operations)
	shapeType := "CAN"
	if isCone {
		shapeType = "CONE"
	}

	fmt.Println()
	fmt.Printf("  Typ ksztaltu:  %s\n", shapeType)
	fmt.Printf("  Part:          %s\n", piece.Part)
	fmt.Printf("  Aktualny Length: %.2f mm\n", piece.Attributes.Length)
	fmt.Printf("  Aktualny Width:  %.2f mm\n", piece.Attributes.Width)
	fmt.Printf("  Thickness:       %.2f mm\n", piece.Attributes.Thickness)
	fmt.Println()

	devLength := promptFloat("Deviation Length [mm] (np. -25 albo +10)", 0)
	devWidth := promptFloat("Deviation Width [mm] (np. -30 albo +15)", 0)

	if devLength == 0 && devWidth == 0 {
		fmt.Println("Brak dewiacji - nie wprowadzono zmian.")
		os.Exit(0)
	}

	newLength := float64(piece.Attributes.Length) + devLength
	newWidth := float64(piece.Attributes.Width) + devWidth

	fmt.Println()
	fmt.Printf("  Nowy Length: %.2f mm (zmiana: %+.2f)\n", newLength, devLength)
	fmt.Printf("  Nowy Width:  %.2f mm (zmiana: %+.2f)\n", newWidth, devWidth)

	outputFile, err := ApplyDeviation(inputFile, devLength, devWidth)
	if err != nil {
		fmt.Printf("Blad dewiacji: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Printf("  Program z dewiacja zapisany: %s\n", outputFile)
	fmt.Println("========================================")
}
