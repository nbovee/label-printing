package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/jung-kurt/gofpdf"
)

const (
	pageWidthInches  = 4.0
	pageHeightInches = 6.0
	marginInches     = 0.25
	fontFamily       = "Courier"

	titleFontSize   = 24
	descFontSize    = 10
	priceFontSize   = 14
	skuFontSize     = 10
	barcodeFontSize = 8
)

type LabelData struct {
	Title       string
	Description string
	Price       string
	SKU         string
	Barcode     string
}

type LabelGenerator struct {
	window      fyne.Window
	statusLabel *widget.Label
}

func NewLabelGenerator(window fyne.Window, statusLabel *widget.Label) *LabelGenerator {
	return &LabelGenerator{
		window:      window,
		statusLabel: statusLabel,
	}
}

func (lg *LabelGenerator) generatePDF(data LabelData) error {
	if strings.TrimSpace(data.Title) == "" {
		return fmt.Errorf("title is required")
	}

	pdf := lg.createPDF()
	
	// Layout the content in standard coordinates first
	lg.layoutPDF(pdf, data)
	
	// Apply rotation transform for export
	// pdf.TransformBegin()
	// // pdf.TransformRotate(90, pageWidthInches/2, pageHeightInches/2)
	// pdf.TransformEnd()
	
	filename := lg.generateFilename(data.SKU)
	if err := pdf.OutputFileAndClose(filename); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	absPath, _ := filepath.Abs(filename)
	lg.statusLabel.SetText(fmt.Sprintf("PDF saved: %s", absPath))
	dialog.ShowInformation("Success", fmt.Sprintf("PDF saved: %s", absPath), lg.window)
	return nil
}

func (lg *LabelGenerator) createPDF() *gofpdf.Fpdf {
	return gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "in",
		SizeStr:        "Custom",
		Size: gofpdf.SizeType{
			Wd: pageWidthInches,
			Ht: pageHeightInches,
		},
	})
}

func (lg *LabelGenerator) layoutPDF(pdf *gofpdf.Fpdf, data LabelData) {
	pdf.AddPage()
	pdf.SetAutoPageBreak(false, marginInches) // Prevent automatic page breaks
	
	contentWidth := pageWidthInches - (2 * marginInches)
	contentHeight := pageHeightInches - (2 * marginInches)

	// Draw title
	lg.drawTitle(pdf, data.Title, contentWidth)
	
	// Draw description
	lg.drawDescription(pdf, data.Description, contentWidth)
	
	// Draw bottom information
	lg.drawBottomInfo(pdf, data, contentWidth)
	
	// Draw border
	lg.drawBorder(pdf, contentWidth, contentHeight)
}

func (lg *LabelGenerator) drawTitle(pdf *gofpdf.Fpdf, title string, contentWidth float64) {
	pdf.SetFont(fontFamily, "B", titleFontSize)
	titleWidth := pdf.GetStringWidth(title)
	titleX := marginInches + (contentWidth-titleWidth)/2
	pdf.SetXY(titleX, marginInches+0.1)
	pdf.Cell(0, 0.3, title)
}

func (lg *LabelGenerator) drawDescription(pdf *gofpdf.Fpdf, description string, contentWidth float64) {
	pdf.SetFont(fontFamily, "", descFontSize)
	pdf.SetXY(marginInches, marginInches+0.5)

	words := strings.Fields(description)
	line := ""
	yPos := marginInches + 0.5
	maxWidth := contentWidth - 0.1
	maxY := pageHeightInches - marginInches - 1.0

	for _, word := range words {
		testLine := line
		if line != "" {
			testLine += " " + word
		} else {
			testLine = word
		}

		if pdf.GetStringWidth(testLine) > maxWidth {
			if line != "" {
				pdf.SetXY(marginInches, yPos)
				pdf.Cell(0, 0.2, line)
				yPos += 0.25
				line = word
			} else {
				line = lg.truncateWord(pdf, word, maxWidth)
			}
		} else {
			line = testLine
		}

		// Check if we're about to exceed the available space
		if yPos > maxY {
			break // Stop adding more lines to prevent page overflow
		}
	}

	// Write the last line if there's space
	if line != "" && yPos <= maxY {
		pdf.SetXY(marginInches, yPos)
		pdf.Cell(0, 0.2, line)
	}
}

func (lg *LabelGenerator) truncateWord(pdf *gofpdf.Fpdf, word string, maxWidth float64) string {
	line := word
	for len(line) > 0 && pdf.GetStringWidth(line) > maxWidth {
		line = line[:len(line)-1]
	}
	return line
}

func (lg *LabelGenerator) drawBottomInfo(pdf *gofpdf.Fpdf, data LabelData, contentWidth float64) {
	bottomY := pageHeightInches - marginInches - 0.6

	// Price (bottom left)
	pdf.SetFont(fontFamily, "B", priceFontSize)
	pdf.SetXY(marginInches, bottomY)
	pdf.Cell(0, 0.3, data.Price)

	// SKU (bottom center)
	pdf.SetFont(fontFamily, "B", skuFontSize+1)
	skuText := "SKU: " + data.SKU
	skuWidth := pdf.GetStringWidth(skuText)
	skuX := marginInches + (contentWidth-skuWidth)/2
	pdf.SetXY(skuX, bottomY)
	pdf.Cell(0, 0.3, skuText)

	// Barcode (bottom right)
	pdf.SetFont(fontFamily, "B", barcodeFontSize+1)
	barcodeText := "BC: " + data.Barcode
	barcodeWidth := pdf.GetStringWidth(barcodeText)
	barcodeX := pageWidthInches - marginInches - barcodeWidth
	pdf.SetXY(barcodeX, bottomY)
	pdf.Cell(0, 0.3, barcodeText)
}

func (lg *LabelGenerator) drawBorder(pdf *gofpdf.Fpdf, contentWidth, contentHeight float64) {
	pdf.SetLineWidth(0.01)
	// Border should match the content area where text is drawn
	pdf.Rect(marginInches, marginInches, contentWidth, contentHeight, "D")
}

func (lg *LabelGenerator) generateFilename(sku string) string {
	safeSKU := strings.ReplaceAll(sku, " ", "_")
	return fmt.Sprintf("label_%s.pdf", safeSKU)
}

func main() {
	a := app.New()
	w := a.NewWindow("Label Printer - 4\"x6\" PDF Generator")
	w.Resize(fyne.NewSize(500, 600))

	// Create form fields
	titleEntry := widget.NewEntry()
	titleEntry.SetText("Sample Product")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetText("This is a sample product description that can span multiple lines.")

	priceEntry := widget.NewEntry()
	priceEntry.SetText("$19.99")

	skuEntry := widget.NewEntry()
	skuEntry.SetText("SKU123456")

	barcodeEntry := widget.NewEntry()
	barcodeEntry.SetText("1234567890123")

	statusLabel := widget.NewLabel("Ready to generate PDF")

	// Create label generator
	generator := NewLabelGenerator(w, statusLabel)

	// Button handlers
	clearFields := func() {
		titleEntry.SetText("")
		descEntry.SetText("")
		priceEntry.SetText("")
		skuEntry.SetText("")
		barcodeEntry.SetText("")
		statusLabel.SetText("Fields cleared")
	}

	generatePDF := func() {
		data := LabelData{
			Title:       titleEntry.Text,
			Description: descEntry.Text,
			Price:       priceEntry.Text,
			SKU:         skuEntry.Text,
			Barcode:     barcodeEntry.Text,
		}

		if err := generator.generatePDF(data); err != nil {
			statusLabel.SetText(fmt.Sprintf("Error: %s", err.Error()))
			dialog.ShowError(err, w)
		}
	}

	// Layout
	form := container.NewVBox(
		widget.NewLabel("Title:"),
		titleEntry,
		widget.NewLabel("Description:"),
		descEntry,
		widget.NewLabel("Price:"),
		priceEntry,
		widget.NewLabel("SKU:"),
		skuEntry,
		widget.NewLabel("Barcode:"),
		barcodeEntry,
		container.NewHBox(
			widget.NewButton("Generate PDF", generatePDF),
			widget.NewButton("Clear Fields", clearFields),
		),
		statusLabel,
	)

	w.SetContent(form)
	w.ShowAndRun()
}
