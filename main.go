package main

import (
	"fmt"
	"path/filepath"
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
		AddItem("Exit", "Close the application", '0', func() {
			app.Stop()
		})

	list.SetBorder(true).
		SetTitle(" POLARIS Program Generator v1.0 ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDodgerBlue)

	frame := centerWidget(list, 60, 11)
	pages.AddAndSwitchToPage("main", frame, true)
}

// ==================== SHAPE SELECTION ====================

func showShapeSelection() {
	list := tview.NewList().
		AddItem("CAN", "Flat rectangular plate", '1', func() {
			showCreateForm("CAN")
		}).
		AddItem("CONE", "Banana-shaped plate", '2', func() {
			showCreateForm("CONE")
		}).
		AddItem("Back", "Return to main menu", '0', func() {
			pages.SwitchToPage("main")
		})

	list.SetBorder(true).
		SetTitle(" Select shape type ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorGreen)

	frame := centerWidget(list, 50, 11)
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
	form.AddInputField("Width [mm]", "", 20, nil, nil)
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
	params.Width, err = parseFloat(form, "Width [mm]")
	if err != nil {
		showError("Invalid Width value")
		return
	}
	params.Thickness, err = parseFloat(form, "Thickness [mm]")
	if err != nil {
		showError("Invalid Thickness value")
		return
	}

	if params.Length <= 0 || params.Width <= 0 || params.Thickness <= 0 {
		showError("Length, Width and Thickness must be > 0")
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
