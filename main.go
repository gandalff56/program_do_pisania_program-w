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
		AddItem("Stworz nowy program", "Utworz nowy program CAN lub CONE", '1', func() {
			showShapeSelection()
		}).
		AddItem("Zaladuj program (dewiacja)", "Zaladuj istniejacy program i zmien wymiary", '2', func() {
			showLoadFileSelection()
		}).
		AddItem("Wyjscie", "Zamknij program", '0', func() {
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
		AddItem("CAN", "Prosta plyta prostokatna", '1', func() {
			showCreateForm("CAN")
		}).
		AddItem("CONE", "Plyta w ksztalcie banana", '2', func() {
			showCreateForm("CONE")
		}).
		AddItem("Powrot", "Wroc do menu glownego", '0', func() {
			pages.SwitchToPage("main")
		})

	list.SetBorder(true).
		SetTitle(" Wybierz typ ksztaltu ").
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

	form.AddButton("Generuj", func() {
		handleCreate(form, shapeType)
	})
	form.AddButton("Anuluj", func() {
		pages.SwitchToPage("shape")
	})

	title := fmt.Sprintf(" Nowy program %s ", shapeType)
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
		showError("Nieprawidlowa wartosc Length")
		return
	}
	params.Width, err = parseFloat(form, "Width [mm]")
	if err != nil {
		showError("Nieprawidlowa wartosc Width")
		return
	}
	params.Thickness, err = parseFloat(form, "Thickness [mm]")
	if err != nil {
		showError("Nieprawidlowa wartosc Thickness")
		return
	}

	if params.Length <= 0 || params.Width <= 0 || params.Thickness <= 0 {
		showError("Length, Width i Thickness musza byc > 0")
		return
	}
	if params.Part == "" {
		showError("Part nie moze byc pusty")
		return
	}

	if shapeType == "CONE" {
		params.PieceHeight, err = parseFloat(form, "Piece Height [mm]")
		if err != nil || params.PieceHeight <= 0 {
			showError("Nieprawidlowa wartosc Piece Height")
			return
		}
		params.LeftOffset, err = parseFloat(form, "Left Offset [mm]")
		if err != nil || params.LeftOffset < 0 {
			showError("Nieprawidlowa wartosc Left Offset")
			return
		}
		params.BottomRadius, err = parseFloat(form, "Bottom Radius [mm]")
		if err != nil || params.BottomRadius <= 0 {
			showError("Nieprawidlowa wartosc Bottom Radius")
			return
		}
		params.TopRadius, err = parseFloat(form, "Top Radius [mm]")
		if err != nil || params.TopRadius <= 0 {
			showError("Nieprawidlowa wartosc Top Radius")
			return
		}
		if params.PieceHeight > params.Width {
			showError("Piece Height nie moze byc wiekszy niz Width")
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
		showError(fmt.Sprintf("Blad zapisu: %v", err))
		return
	}

	showSuccess(fmt.Sprintf("Program zapisany:\n\n[yellow]%s", fileName))
}

// ==================== LOAD / DEVIATION ====================

func showLoadFileSelection() {
	files, _ := filepath.Glob("*.Program.polaris")

	if len(files) == 0 {
		showError("Brak plikow .Program.polaris w biezacym katalogu")
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
	list.AddItem("Powrot", "Wroc do menu glownego", '0', func() {
		pages.SwitchToPage("main")
	})

	list.SetBorder(true).
		SetTitle(" Wybierz plik do zaladowania ").
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
		showError(fmt.Sprintf("Blad odczytu: %v", err))
		return
	}

	if len(doc.Piece.Objects) == 0 {
		showError("Brak obiektow PieceObject w pliku")
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
			"[white]Plik:      [yellow]%s\n"+
				"[white]Typ:       [green]%s\n"+
				"[white]Part:      [yellow]%s\n"+
				"[white]Length:    [cyan]%.2f mm\n"+
				"[white]Width:     [cyan]%.2f mm\n"+
				"[white]Thickness: [cyan]%.2f mm",
			filePath, shapeType, piece.Part, oldLength, oldWidth, thickness,
		))
	info.SetBorder(true).
		SetTitle(" Informacje o programie ").
		SetBorderColor(tcell.ColorDarkCyan)

	// Deviation form
	form := tview.NewForm()
	form.AddInputField("Deviation Length [mm]", "0", 20, nil, nil)
	form.AddInputField("Deviation Width [mm]", "0", 20, nil, nil)

	form.AddButton("Zastosuj", func() {
		devL, err1 := parseFloat(form, "Deviation Length [mm]")
		devW, err2 := parseFloat(form, "Deviation Width [mm]")
		if err1 != nil || err2 != nil {
			showError("Nieprawidlowe wartosci dewiacji")
			return
		}
		if devL == 0 && devW == 0 {
			showError("Nie wprowadzono zadnej dewiacji")
			return
		}

		newL := oldLength + devL
		newW := oldWidth + devW
		if newL <= 0 || newW <= 0 {
			showError(fmt.Sprintf("Wymiary po dewiacji sa nieprawidlowe: L=%.2f, W=%.2f", newL, newW))
			return
		}

		outputFile, err := ApplyDeviation(filePath, devL, devW)
		if err != nil {
			showError(fmt.Sprintf("Blad dewiacji: %v", err))
			return
		}

		showSuccess(fmt.Sprintf(
			"Dewiacja zastosowana!\n\n"+
				"[white]Length: [cyan]%.2f[white] -> [green]%.2f[white] (zmiana: [yellow]%+.2f[white])\n"+
				"[white]Width:  [cyan]%.2f[white] -> [green]%.2f[white] (zmiana: [yellow]%+.2f[white])\n\n"+
				"[white]Plik: [yellow]%s",
			oldLength, newL, devL, oldWidth, newW, devW, outputFile,
		))
	})
	form.AddButton("Anuluj", func() {
		showLoadFileSelection()
	})

	form.SetBorder(true).
		SetTitle(" Dewiacja wymiarow ").
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
		SetText("[red]BLAD\n\n[white]" + msg).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("error")
		}).
		SetBackgroundColor(tcell.ColorDarkRed)

	pages.AddAndSwitchToPage("error", modal, true)
}

func showSuccess(msg string) {
	modal := tview.NewModal().
		SetText("[green]SUKCES\n\n" + msg).
		AddButtons([]string{"OK", "Menu glowne"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("success")
			if buttonLabel == "Menu glowne" {
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
