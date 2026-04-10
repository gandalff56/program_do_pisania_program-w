package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app   *tview.Application
	pages *tview.Pages
)

func main() {
	// Fix double-input issue on some Windows terminals
	if runtime.GOOS == "windows" {
		os.Setenv("TCELL_CONSOLE", "true")
	}

	app = tview.NewApplication()
	pages = tview.NewPages()

	showMainMenu()

	pages.SetBackgroundColor(tcell.ColorDefault)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// ==================== MAIN MENU ====================

func showMainMenu() {
	list := tview.NewList().
		AddItem("Create new program", "Create a new CAN or CONE program", '1', func() {
			showShapeSelection()
		}).
		AddItem("Load program (deviation)", "Load an existing program and adjust dimensions", '2', func() {
			showLoadFileSelection()
		}).
		AddItem("Add J-BEVEL to program", "Load a program and create a J-BEVEL variant", '3', func() {
			showJBevelFileSelection()
		}).
		AddItem("Exit", "Close the application", '0', func() {
			app.Stop()
		})

	list.SetBorder(true).
		SetTitle(" POLARIS Program Generator v1.0 ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDodgerBlue)

	frame := centerWidget(list, 60, 13)
	pages.AddAndSwitchToPage("main", frame, true)
}

// ==================== SHAPE SELECTION ====================

func showShapeSelection() {
	list := tview.NewList().
		AddItem("CAN", "Flat rectangular plate", '1', func() {
			showCreateForm("CAN")
		}).
		AddItem("CONE (manual)", "Banana-shaped plate - enter all parameters", '2', func() {
			showCreateForm("CONE")
		}).
		AddItem("CONE (auto-calc)", "Banana-shaped plate - calculate from x1, x2, w", '3', func() {
			showConeAutoCalcForm()
		}).
		AddItem("Back", "Return to main menu", '0', func() {
			pages.SwitchToPage("main")
		})

	list.SetBorder(true).
		SetTitle(" Select shape type ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorGreen)

	frame := centerWidget(list, 60, 13)
	pages.AddAndSwitchToPage("shape", frame, true)
}

// ==================== CREATE FORM ====================

func showCreateForm(shapeType string) {
	form := tview.NewForm()

	// Common fields
	form.AddInputField("Contract", "B2149", 30, nil, nil)
	form.AddInputField("Project", "B2149", 30, nil, nil)
	form.AddInputField("Drawing", "MONOPILE_PARTS_A", 30, nil, nil)
	form.AddInputField("Part", "", 30, nil, nil)
	form.AddInputField("Length [mm]", "", 20, nil, nil)
	if shapeType == "CONE" {
		form.AddInputField("Width [mm]", "", 20, nil, nil).
			GetFormItemByLabel("Width [mm]").(*tview.InputField).
			SetPlaceholder("auto if empty")
	} else {
		form.AddInputField("Width [mm]", "", 20, nil, nil)
	}
	form.AddInputField("Thickness [mm]", "80.0", 20, nil, nil)
	form.AddInputField("Material Code", "S355ML", 20, nil, nil)
	form.AddInputField("Working Area", "A", 10, nil, nil)

	// CONE-specific fields
	if shapeType == "CONE" {
		form.AddInputField("Piece Height [mm]", "", 20, nil, nil)
		form.AddInputField("Left Offset [mm]", "", 20, nil, nil)
		form.AddInputField("Bottom Radius [mm]", "", 20, nil, nil)
		form.AddInputField("Top Radius [mm]", "", 20, nil, nil)
	}

	form.AddButton("Generate", func() {
		handleCreate(form, shapeType)
	})
	form.AddButton("Cancel", func() {
		pages.SwitchToPage("shape")
	})

	title := fmt.Sprintf(" New %s program ", shapeType)
	form.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorGreen)

	height := 27
	if shapeType == "CONE" {
		height = 35
	}
	frame := centerWidget(form, 65, height)
	pages.AddAndSwitchToPage("create", frame, true)
}

func handleCreate(form *tview.Form, shapeType string) {
	params := ShapeParams{
		ShapeType:    shapeType,
		Contract:     form.GetFormItemByLabel("Contract").(*tview.InputField).GetText(),
		Project:      form.GetFormItemByLabel("Project").(*tview.InputField).GetText(),
		Drawing:      form.GetFormItemByLabel("Drawing").(*tview.InputField).GetText(),
		Part:         form.GetFormItemByLabel("Part").(*tview.InputField).GetText(),
		MaterialCode: form.GetFormItemByLabel("Material Code").(*tview.InputField).GetText(),
		WorkingArea:  form.GetFormItemByLabel("Working Area").(*tview.InputField).GetText(),
	}

	var err error
	params.Length, err = parseFloat(form, "Length [mm]")
	if err != nil {
		showError("Invalid Length value")
		return
	}
	params.Thickness, err = parseFloat(form, "Thickness [mm]")
	if err != nil {
		showError("Invalid Thickness value")
		return
	}

	// Width: required for CAN, optional for CONE (auto-calculated if empty)
	widthText := form.GetFormItemByLabel("Width [mm]").(*tview.InputField).GetText()
	if widthText != "" {
		params.Width, err = strconv.ParseFloat(strings.TrimSpace(widthText), 64)
		if err != nil {
			showError("Invalid Width value")
			return
		}
	}

	if params.Length <= 0 || params.Thickness <= 0 {
		showError("Length and Thickness must be > 0")
		return
	}
	if shapeType == "CAN" && params.Width <= 0 {
		showError("Width is required for CAN")
		return
	}
	if params.Part == "" {
		showError("Part cannot be empty")
		return
	}

	if shapeType == "CONE" {
		params.PieceHeight, err = parseFloat(form, "Piece Height [mm]")
		if err != nil || params.PieceHeight <= 0 {
			showError("Invalid Piece Height value")
			return
		}
		params.LeftOffset, err = parseFloat(form, "Left Offset [mm]")
		if err != nil || params.LeftOffset < 0 {
			showError("Invalid Left Offset value")
			return
		}
		params.BottomRadius, err = parseFloat(form, "Bottom Radius [mm]")
		if err != nil || params.BottomRadius <= 0 {
			showError("Invalid Bottom Radius value")
			return
		}
		params.TopRadius, err = parseFloat(form, "Top Radius [mm]")
		if err != nil || params.TopRadius <= 0 {
			showError("Invalid Top Radius value")
			return
		}

		// Auto-calculate Width if left empty
		if params.Width <= 0 {
			params.Width = autoCalcBoundingWidth(params.PieceHeight, params.Length, params.TopRadius)
		}
		if params.PieceHeight > params.Width {
			showError("Piece Height cannot be greater than Width")
			return
		}
	}

	var doc *PolarisDocument
	var fileName string

	if shapeType == "CAN" {
		doc, fileName = GenerateCAN(params)
	} else {
		doc, fileName = GenerateCONE(params)
	}

	if err := WritePolarisFile(fileName, doc); err != nil {
		showError(fmt.Sprintf("Write error: %v", err))
		return
	}

	showSuccess(fmt.Sprintf("Program saved:\n\n[yellow]%s", fileName))
}

// ==================== CONE AUTO-CALC ====================

func showConeAutoCalcForm() {
	form := tview.NewForm()

	// Drawing dimensions
	form.AddInputField("x1 - top arc length [mm]", "", 20, nil, nil)
	form.AddInputField("x2 - bottom arc length [mm]", "", 20, nil, nil)
	form.AddInputField("w - side edge length [mm]", "", 20, nil, nil)

	// Common fields
	form.AddInputField("Thickness [mm]", "80.0", 20, nil, nil)
	form.AddInputField("Contract", "B2149", 30, nil, nil)
	form.AddInputField("Project", "B2149", 30, nil, nil)
	form.AddInputField("Drawing", "MONOPILE_PARTS_A", 30, nil, nil)
	form.AddInputField("Part", "", 30, nil, nil)
	form.AddInputField("Material Code", "S355ML", 20, nil, nil)
	form.AddInputField("Working Area", "A", 10, nil, nil)

	form.AddButton("Calculate & Generate", func() {
		handleConeAutoCalc(form)
	})
	form.AddButton("Cancel", func() {
		pages.SwitchToPage("shape")
	})

	form.SetBorder(true).
		SetTitle(" New CONE program (auto-calculate from x1, x2, w) ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorGreen)

	frame := centerWidget(form, 70, 30)
	pages.AddAndSwitchToPage("cone-auto", frame, true)
}

func handleConeAutoCalc(form *tview.Form) {
	x1, err := parseFloat(form, "x1 - top arc length [mm]")
	if err != nil || x1 <= 0 {
		showError("Invalid x1 (top arc length)")
		return
	}
	x2, err := parseFloat(form, "x2 - bottom arc length [mm]")
	if err != nil || x2 <= 0 {
		showError("Invalid x2 (bottom arc length)")
		return
	}
	w, err := parseFloat(form, "w - side edge length [mm]")
	if err != nil || w <= 0 {
		showError("Invalid w (side edge length)")
		return
	}
	thickness, err := parseFloat(form, "Thickness [mm]")
	if err != nil || thickness <= 0 {
		showError("Invalid Thickness")
		return
	}

	if x1 <= x2 {
		showError("x1 (top arc) must be longer than x2 (bottom arc)")
		return
	}

	// Calculate CONE geometry from arc lengths
	calc, err := CalculateConeFromArcs(x1, x2, w)
	if err != nil {
		showError(fmt.Sprintf("Calculation error: %v", err))
		return
	}

	contract := form.GetFormItemByLabel("Contract").(*tview.InputField).GetText()
	project := form.GetFormItemByLabel("Project").(*tview.InputField).GetText()
	drawing := form.GetFormItemByLabel("Drawing").(*tview.InputField).GetText()
	part := form.GetFormItemByLabel("Part").(*tview.InputField).GetText()
	materialCode := form.GetFormItemByLabel("Material Code").(*tview.InputField).GetText()
	workingArea := form.GetFormItemByLabel("Working Area").(*tview.InputField).GetText()

	if part == "" {
		showError("Part cannot be empty")
		return
	}

	params := ShapeParams{
		ShapeType:    "CONE",
		Contract:     contract,
		Project:      project,
		Drawing:      drawing,
		Part:         part,
		Length:       calc.Length,
		Width:        calc.BoundingWidth,
		Thickness:    thickness,
		MaterialCode: materialCode,
		WorkingArea:  workingArea,
		PieceHeight:  calc.PieceHeight,
		LeftOffset:   calc.LeftOffset,
		BottomRadius: calc.BottomRadius,
		TopRadius:    calc.TopRadius,
	}

	doc, fileName := GenerateCONE(params)

	if err := WritePolarisFile(fileName, doc); err != nil {
		showError(fmt.Sprintf("Write error: %v", err))
		return
	}

	showSuccess(fmt.Sprintf(
		"CONE program created (auto-calculated)!\n\n"+
			"[white]Calculated values:\n"+
			"[white]  Length:       [cyan]%.2f mm\n"+
			"[white]  Width:        [cyan]%.2f mm\n"+
			"[white]  PieceHeight:  [cyan]%.2f mm\n"+
			"[white]  LeftOffset:   [cyan]%.2f mm\n"+
			"[white]  BottomRadius: [cyan]%.2f mm\n"+
			"[white]  TopRadius:    [cyan]%.2f mm\n\n"+
			"[white]File: [yellow]%s",
		calc.Length, calc.BoundingWidth, calc.PieceHeight,
		calc.LeftOffset, calc.BottomRadius, calc.TopRadius, fileName,
	))
}

// ==================== LOAD / DEVIATION ====================

func showLoadFileSelection() {
	files, _ := filepath.Glob("*.Program.polaris")

	if len(files) == 0 {
		showError("No .Program.polaris files found in current directory")
		return
	}

	list := tview.NewList()
	for i, f := range files {
		file := f // capture
		shortcut := rune('a' + i)
		if i > 25 {
			shortcut = 0
		}
		list.AddItem(file, "", shortcut, func() {
			showDeviationForm(file)
		})
	}
	list.AddItem("Back", "Return to main menu", '0', func() {
		pages.SwitchToPage("main")
	})

	list.SetBorder(true).
		SetTitle(" Select file to load ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorYellow)

	height := len(files)*2 + 6
	if height > 30 {
		height = 30
	}
	frame := centerWidget(list, 70, height)
	pages.AddAndSwitchToPage("load", frame, true)
}

func showDeviationForm(filePath string) {
	doc, err := ReadPolarisFile(filePath)
	if err != nil {
		showError(fmt.Sprintf("Read error: %v", err))
		return
	}

	if len(doc.Piece.Objects) == 0 {
		showError("No PieceObject found in file")
		return
	}

	piece := doc.Piece.Objects[0]
	isCone := detectCone(piece.Operations)
	shapeType := "CAN"
	if isCone {
		shapeType = "CONE"
	}

	oldLength := float64(piece.Attributes.Length)
	oldWidth := float64(piece.Attributes.Width)
	thickness := float64(piece.Attributes.Thickness)

	// Info panel
	info := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf(
			"[white]File:      [yellow]%s\n"+
				"[white]Type:      [green]%s\n"+
				"[white]Part:      [yellow]%s\n"+
				"[white]Length:    [cyan]%.2f mm\n"+
				"[white]Width:     [cyan]%.2f mm\n"+
				"[white]Thickness: [cyan]%.2f mm",
			filePath, shapeType, piece.Part, oldLength, oldWidth, thickness,
		))
	info.SetBorder(true).
		SetTitle(" Program info ").
		SetBorderColor(tcell.ColorDarkCyan)

	// Deviation form
	form := tview.NewForm()
	form.AddInputField("Deviation Length [mm]", "0", 20, nil, nil)
	form.AddInputField("Deviation Width [mm]", "0", 20, nil, nil)

	form.AddButton("Apply", func() {
		devL, err1 := parseFloat(form, "Deviation Length [mm]")
		devW, err2 := parseFloat(form, "Deviation Width [mm]")
		if err1 != nil || err2 != nil {
			showError("Invalid deviation values")
			return
		}
		if devL == 0 && devW == 0 {
			showError("No deviation entered")
			return
		}

		newL := oldLength + devL
		newW := oldWidth + devW
		if newL <= 0 || newW <= 0 {
			showError(fmt.Sprintf("Invalid dimensions after deviation: L=%.2f, W=%.2f", newL, newW))
			return
		}

		outputFile, err := ApplyDeviation(filePath, devL, devW)
		if err != nil {
			showError(fmt.Sprintf("Deviation error: %v", err))
			return
		}

		showSuccess(fmt.Sprintf(
			"Deviation applied!\n\n"+
				"[white]Length: [cyan]%.2f[white] -> [green]%.2f[white] (change: [yellow]%+.2f[white])\n"+
				"[white]Width:  [cyan]%.2f[white] -> [green]%.2f[white] (change: [yellow]%+.2f[white])\n\n"+
				"[white]File: [yellow]%s",
			oldLength, newL, devL, oldWidth, newW, devW, outputFile,
		))
	})
	form.AddButton("Cancel", func() {
		showLoadFileSelection()
	})

	form.SetBorder(true).
		SetTitle(" Dimension deviation ").
		SetBorderColor(tcell.ColorYellow)

	// Layout: info on top, form on bottom
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(info, 9, 0, false).
		AddItem(form, 0, 1, true)

	frame := centerWidget(flex, 65, 24)
	pages.AddAndSwitchToPage("deviation", frame, true)
}

// ==================== J-BEVEL ====================

func showJBevelFileSelection() {
	files, _ := filepath.Glob("*.Program.polaris")

	if len(files) == 0 {
		showError("No .Program.polaris files found in current directory")
		return
	}

	list := tview.NewList()
	for i, f := range files {
		file := f
		shortcut := rune('a' + i)
		if i > 25 {
			shortcut = 0
		}
		list.AddItem(file, "", shortcut, func() {
			showJBevelForm(file)
		})
	}
	list.AddItem("Back", "Return to main menu", '0', func() {
		pages.SwitchToPage("main")
	})

	list.SetBorder(true).
		SetTitle(" Select program for J-BEVEL ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorOrange)

	height := len(files)*2 + 6
	if height > 30 {
		height = 30
	}
	frame := centerWidget(list, 70, height)
	pages.AddAndSwitchToPage("jbevel-load", frame, true)
}

func showJBevelForm(filePath string) {
	doc, err := ReadPolarisFile(filePath)
	if err != nil {
		showError(fmt.Sprintf("Read error: %v", err))
		return
	}

	if len(doc.Piece.Objects) == 0 || len(doc.Program.Objects) == 0 {
		showError("Invalid program file")
		return
	}

	piece := doc.Piece.Objects[0]
	isCone := detectCone(piece.Operations)
	shapeType := "CAN"
	if isCone {
		shapeType = "CONE"
	}

	oldLength := float64(piece.Attributes.Length)
	oldWidth := float64(piece.Attributes.Width)
	oldThickness := float64(piece.Attributes.Thickness)

	// Info panel
	info := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf(
			"[white]File:      [yellow]%s\n"+
				"[white]Type:      [green]%s\n"+
				"[white]Part:      [yellow]%s\n"+
				"[white]Length:    [cyan]%.2f mm\n"+
				"[white]Width:     [cyan]%.2f mm\n"+
				"[white]Thickness: [cyan]%.2f mm",
			filePath, shapeType, piece.Part, oldLength, oldWidth, oldThickness,
		))
	info.SetBorder(true).
		SetTitle(" Source program ").
		SetBorderColor(tcell.ColorDarkCyan)

	// J-BEVEL form
	form := tview.NewForm()
	form.AddInputField("J-BEVEL number", "2", 10, nil, nil)
	form.AddInputField("New Length [mm]", fmt.Sprintf("%g", oldLength), 20, nil, nil)
	form.AddInputField("New Width [mm]", fmt.Sprintf("%g", oldWidth), 20, nil, nil)
	form.AddInputField("New Thickness [mm]", fmt.Sprintf("%g", oldThickness), 20, nil, nil)
	form.AddInputField("Working Area", "AB", 10, nil, nil)

	form.AddButton("Generate J-BEVEL", func() {
		handleJBevel(form, filePath, doc)
	})
	form.AddButton("Cancel", func() {
		showJBevelFileSelection()
	})

	form.SetBorder(true).
		SetTitle(" J-BEVEL configuration ").
		SetBorderColor(tcell.ColorOrange)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(info, 9, 0, false).
		AddItem(form, 0, 1, true)

	frame := centerWidget(flex, 65, 30)
	pages.AddAndSwitchToPage("jbevel-form", frame, true)
}

func handleJBevel(form *tview.Form, sourcePath string, doc *PolarisDocument) {
	bevelNum := form.GetFormItemByLabel("J-BEVEL number").(*tview.InputField).GetText()
	workingArea := form.GetFormItemByLabel("Working Area").(*tview.InputField).GetText()

	newLength, err := parseFloat(form, "New Length [mm]")
	if err != nil || newLength <= 0 {
		showError("Invalid Length value")
		return
	}
	newWidth, err := parseFloat(form, "New Width [mm]")
	if err != nil || newWidth <= 0 {
		showError("Invalid Width value")
		return
	}
	newThickness, err := parseFloat(form, "New Thickness [mm]")
	if err != nil || newThickness <= 0 {
		showError("Invalid Thickness value")
		return
	}

	outputFile, err := ApplyJBevel(doc, bevelNum, newLength, newWidth, newThickness, workingArea)
	if err != nil {
		showError(fmt.Sprintf("J-BEVEL error: %v", err))
		return
	}

	showSuccess(fmt.Sprintf(
		"J-BEVEL program created!\n\n"+
			"[white]Part: [yellow]%s\n"+
			"[white]Length: [cyan]%.2f mm\n"+
			"[white]Width:  [cyan]%.2f mm\n"+
			"[white]Thickness: [cyan]%.2f mm\n\n"+
			"[white]File: [yellow]%s",
		doc.Piece.Objects[0].Part, newLength, newWidth, newThickness, outputFile,
	))
}

// ==================== DIALOGS ====================

func showError(msg string) {
	modal := tview.NewModal().
		SetText("[red]ERROR\n\n[white]" + msg).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("error")
		}).
		SetBackgroundColor(tcell.ColorDarkRed)

	pages.AddAndSwitchToPage("error", modal, true)
}

func showSuccess(msg string) {
	modal := tview.NewModal().
		SetText("[green]SUCCESS\n\n" + msg).
		AddButtons([]string{"OK", "Main menu"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("success")
			if buttonLabel == "Main menu" {
				pages.SwitchToPage("main")
			}
		}).
		SetBackgroundColor(tcell.ColorDarkGreen)

	pages.AddAndSwitchToPage("success", modal, true)
}

// ==================== HELPERS ====================

func parseFloat(form *tview.Form, label string) (float64, error) {
	text := form.GetFormItemByLabel(label).(*tview.InputField).GetText()
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("empty")
	}
	return strconv.ParseFloat(text, 64)
}

func centerWidget(widget tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(widget, height, 0, true).
				AddItem(nil, 0, 1, false),
			width, 0, true,
		).
		AddItem(nil, 0, 1, false)
}
