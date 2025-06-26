package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
)

const (
	pageWidthInches  = 4.0
	pageHeightInches = 6.0
	marginInches     = 0.125
	fontFamily       = "Courier"

	titleFontSize   = 48
	descFontSize    = 10
	priceFontSize   = 14
	skuFontSize     = 10
	barcodeFontSize = 8
)

type LabelData struct {
	Title          string
	Description    string
	ReturnLocation string
	SKU            string
	Barcode        string
	CheckoutDate   string
	ReturnDate     string
	URL1           string
	URL2           string
	URL3           string
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

	lg.layoutPDF(pdf, data)

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
		OrientationStr: "L",
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

	contentWidth := pageHeightInches - (2 * marginInches) // Height is used as we are in landscape
	contentHeight := pageWidthInches - (2 * marginInches) // Width is used as we are in landscape

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
	words := strings.Fields(title)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	title = strings.Join(words, " ")

	// Center the title using CellFormat with alignStr
	pdf.SetXY(marginInches, marginInches+0.1)
	pdf.CellFormat(contentWidth, 0.3, title, "0", 0, "C", false, 0, "")
}

func (lg *LabelGenerator) drawDescription(pdf *gofpdf.Fpdf, description string, contentWidth float64) {
	pdf.SetFont(fontFamily, "", descFontSize)
	pdf.SetXY(marginInches, marginInches+0.625)

	words := strings.Fields(description)
	line := ""
	yPos := marginInches + 0.625
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
				pdf.CellFormat(pdf.GetStringWidth(line), 0.2, line, "0", 0, "L", false, 0, "")
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
		pdf.CellFormat(pdf.GetStringWidth(line), 0.2, line, "0", 0, "L", false, 0, "")
	}
}

func (lg *LabelGenerator) truncateWord(pdf *gofpdf.Fpdf, word string, maxWidth float64) string {
	line := word
	for len(line) > 0 && pdf.GetStringWidth(line) > maxWidth {
		line = line[:len(line)-1]
	}
	return line
}

func (lg *LabelGenerator) generateQRCode(url string, filename string) error {
	if strings.TrimSpace(url) == "" {
		return nil // Skip empty URLs
	}

	// Generate QR code
	err := qrcode.WriteFile(url, qrcode.Medium, 256, filename)
	return err
}

func (lg *LabelGenerator) drawBottomInfo(pdf *gofpdf.Fpdf, data LabelData, contentWidth float64) {
	bottomY := pageWidthInches - marginInches - 0.6

	// Barcode (above bottom row, right aligned)
	pdf.SetFont(fontFamily, "B", barcodeFontSize+1)
	barcodeText := "BC: " + data.Barcode
	pdf.SetXY(pageWidthInches - marginInches - pdf.GetStringWidth(barcodeText), bottomY-0.35)
	pdf.CellFormat(pdf.GetStringWidth(barcodeText), 0.3, barcodeText, "0", 0, "R", false, 0, "")

	// SKU (bottom center)
	pdf.SetFont(fontFamily, "B", skuFontSize+1)
	skuText := "SKU: " + data.SKU
	pdf.SetXY(marginInches, bottomY)
	pdf.CellFormat(contentWidth, 0.3, skuText, "0", 0, "C", false, 0, "")

	// Return Date (bottom right)
	pdf.SetFont(fontFamily, "B", priceFontSize)
	returnText := "Return By: " + data.ReturnDate
	pdf.SetXY(pageWidthInches - marginInches - pdf.GetStringWidth(returnText), bottomY)
	pdf.CellFormat(pdf.GetStringWidth(returnText), 0.3, returnText, "0", 0, "R", false, 0, "")

	// Add QR codes for URLs
	lg.drawQRCodes(pdf, data)
}

func (lg *LabelGenerator) drawQRCodes(pdf *gofpdf.Fpdf, data LabelData) {
	// QR code size in inches
	qrSize := 0.5

	// Generate and add QR code (only URL1)
	url := data.URL1
	label := "Finalize Restock"
	qrY := pageWidthInches - marginInches - qrSize // Position QR code at bottom

	// Add Borrowed date above Return to
	pdf.SetFont(fontFamily, "B", priceFontSize)
	borrowedY := qrY - 0.8 // Position above Return to
	pdf.SetXY(marginInches, borrowedY)
	borrowedText := "Borrowed: " + data.CheckoutDate
	pdf.CellFormat(pdf.GetStringWidth(borrowedText), 0.3, borrowedText, "0", 0, "L", false, 0, "")

	// Add Return Location above QR code
	pdf.SetFont(fontFamily, "B", priceFontSize)
	returnY := qrY - 0.5 // Position above QR code
	pdf.SetXY(marginInches, returnY)
	returnText := "Return To: " + data.ReturnLocation
	pdf.CellFormat(pdf.GetStringWidth(returnText), 0.3, returnText, "0", 0, "L", false, 0, "")

	if strings.TrimSpace(url) != "" {
		// Generate QR code file
		qrFilename := "qr_1.png"
		if err := lg.generateQRCode(url, qrFilename); err == nil {
			// Position QR code at bottom right (where QR3 was)
			qrX := pageHeightInches - marginInches - qrSize

			// Add label above QR code
			pdf.SetFont(fontFamily, "B", 8)
			labelWidth := pdf.GetStringWidth(label)
			labelX := qrX + (qrSize-labelWidth)/2 // Center label over QR code
			labelY := qrY - 0.15                  // Position label above QR code
			pdf.SetXY(labelX, labelY)
			pdf.CellFormat(labelWidth, 0.1, label, "0", 0, "C", false, 0, "")

			// Add QR code to PDF
			pdf.SetXY(qrX, qrY)
			pdf.Image(qrFilename, qrX, qrY, qrSize, qrSize, false, "", 0, "")

			// Clean up temporary file
			os.Remove(qrFilename)
		}
	}
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
	w.Resize(fyne.NewSize(500, 700))

	// Create form fields
	titleEntry := widget.NewEntry()
	titleEntry.SetText("Sample Equipment Tag")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetText("This is a sample item description that can span multiple lines. We can see this as the default description is quite long.")

	returnLocationEntry := widget.NewEntry()
	returnLocationEntry.SetText("Engineering Hall 317")

	skuEntry := widget.NewEntry()
	skuEntry.SetText("SKU123456")

	barcodeEntry := widget.NewEntry()
	barcodeEntry.SetText("1234567890123")

	checkoutDateEntry := widget.NewEntry()
	checkoutDateEntry.SetText("01/15/2024")

	returnDateEntry := widget.NewEntry()
	returnDateEntry.SetText("01/22/2024")

	url1Entry := widget.NewEntry()
	url1Entry.SetText("https://example.com/product1")

	statusLabel := widget.NewLabel("Ready to generate PDF")

	// Create label generator
	generator := NewLabelGenerator(w, statusLabel)

	// Button handlers
	clearFields := func() {
		titleEntry.SetText("")
		descEntry.SetText("")
		returnLocationEntry.SetText("")
		skuEntry.SetText("")
		barcodeEntry.SetText("")
		checkoutDateEntry.SetText("")
		returnDateEntry.SetText("")
		url1Entry.SetText("")
		statusLabel.SetText("Fields cleared")
	}

	generatePDF := func() {
		data := LabelData{
			Title:          titleEntry.Text,
			Description:    descEntry.Text,
			ReturnLocation: returnLocationEntry.Text,
			SKU:            skuEntry.Text,
			Barcode:        barcodeEntry.Text,
			CheckoutDate:   checkoutDateEntry.Text,
			ReturnDate:     returnDateEntry.Text,
			URL1:           url1Entry.Text,
			URL2:           "",
			URL3:           "",
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
		container.NewPadded(descEntry),
		widget.NewLabel("Return Location:"),
		returnLocationEntry,
		widget.NewLabel("SKU:"),
		skuEntry,
		widget.NewLabel("Barcode:"),
		barcodeEntry,
		widget.NewLabel("Checkout Date:"),
		checkoutDateEntry,
		widget.NewLabel("Return Date:"),
		returnDateEntry,
		widget.NewLabel("URL (QR Code):"),
		url1Entry,
		container.NewHBox(
			widget.NewButton("Generate PDF", generatePDF),
			widget.NewButton("Clear Fields", clearFields),
		),
		statusLabel,
	)

	w.SetContent(form)
	w.ShowAndRun()
}
