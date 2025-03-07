package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/fogleman/gg"
	"github.com/rwcarlsen/goexif/exif"
	xdw "golang.org/x/image/draw"
)

// Special EXIF text keys
var exifKeyMap = map[string]struct{}{
	"expose": {},
	"device": {},
	"lens":   {},
	"time":   {},
}

// Available fonts
var fontMap = map[string]struct{}{
	"SFCamera.ttf":        {},
	"SFCompactItalic.ttf": {},
	"NewYork.ttf":         {},
	"STHeiti Medium.ttc":  {},
}

// ExifData stores image metadata extracted from EXIF
type ExifData map[string]interface{}

// AddWhiteBorderWithText processes an image by adding borders and text overlays
func AddWhiteBorderWithText(imgPath, outputPath string, config Config) error {
	fmt.Println("Processing image:", imgPath)

	// Open and decode source image
	img, err := loadImage(imgPath)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	// Extract EXIF data
	exifData, err := getExifData(imgPath)
	if err != nil {
		fmt.Println("Warning: Could not extract EXIF data:", err)
		exifData = make(ExifData) // Use empty EXIF data, continue processing
	}

	// Create canvas with borders
	canvas, borderTopHeight, borderBottomHeight, borderWidth := createCanvasWithBorders(img, config.Border)

	// Create drawing context
	dc := gg.NewContextForImage(canvas)

	// Add logo if enabled
	logoWidth := 0.0
	if config.Logo.On {
		logoWidth, err = addLogoToCanvas(dc, exifData, img, config.Logo, borderWidth, borderBottomHeight, canvas.Bounds().Dy())
		if err != nil {
			fmt.Println("Warning: Could not add logo:", err)
			// Continue processing without logo
		}
	}

	// Add text overlays
	for _, textConfig := range config.Text {
		if !textConfig.On {
			continue
		}

		addTextOverlay(dc, textConfig, exifData, img, canvas, borderWidth, borderTopHeight,
			borderBottomHeight, config.Logo.On, logoWidth)
	}

	// Save processed image
	return saveImage(dc.Image(), outputPath, config.Quality)
}

// loadImage opens and decodes an image file
func loadImage(imgPath string) (image.Image, error) {
	imgFile, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// getExifData extracts metadata from an image's EXIF tags
func getExifData(imgPath string) (ExifData, error) {
	exifData := make(ExifData)

	file, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	x, err := exif.Decode(file)
	if err != nil {
		return nil, err
	}

	// Extract common EXIF fields
	if extractRatField(x, exif.FocalLength, &exifData, "FocalLength", true) != nil {
		// Continue if field is missing
	}

	if extractRatField(x, exif.FNumber, &exifData, "FNumber", false) != nil {
		// Continue if field is missing
	}

	if extractRatField(x, exif.ExposureTime, &exifData, "ExposureTime", false) != nil {
		// Continue if field is missing
	}

	if val, err := x.Get(exif.ISOSpeedRatings); err == nil {
		exifData["ISOSpeedRatings"] = strconv.FormatUint(uint64(val.Val[0]), 10)
	}

	if val, err := x.Get(exif.DateTimeOriginal); err == nil {
		dt, _ := val.StringVal()
		if t, err := time.Parse("2006:01:02 15:04:05", dt); err == nil {
			exifData["DateTimeOriginal"] = t.Format("2006-01-02 15:04:05")
		}
	}

	extractStringField(x, exif.Make, &exifData, "Make")
	extractStringField(x, exif.Model, &exifData, "Model")
	extractStringField(x, exif.LensModel, &exifData, "LensModel")

	return exifData, nil
}

// extractRatField extracts a rational EXIF field and formats it appropriately
func extractRatField(x *exif.Exif, field exif.FieldName, data *ExifData, key string, isInteger bool) error {
	val, err := x.Get(field)
	if err != nil {
		return err
	}

	num, den, _ := val.Rat2(0)

	if isInteger {
		(*data)[key] = strconv.FormatInt(num/den, 10)
	} else if field == exif.ExposureTime {
		(*data)[key] = strconv.FormatInt(num, 10) + "/" + strconv.FormatInt(den, 10)
	} else {
		(*data)[key] = strconv.FormatFloat(float64(num)/float64(den), 'g', -1, 64)
	}

	return nil
}

// extractStringField extracts a string EXIF field
func extractStringField(x *exif.Exif, field exif.FieldName, data *ExifData, key string) {
	if val, err := x.Get(field); err == nil {
		(*data)[key], _ = val.StringVal()
	}
}

// formatExifText formats EXIF data according to predefined templates
func formatExifText(exifData ExifData, key string) string {
	switch key {
	case "expose":
		focalLength, ok1 := exifData["FocalLength"].(string)
		fNumber, ok2 := exifData["FNumber"].(string)
		exposureTime, ok3 := exifData["ExposureTime"].(string)
		iso, ok4 := exifData["ISOSpeedRatings"].(string)

		if !ok1 || !ok2 || !ok3 || !ok4 {
			return "EXIF data incomplete"
		}

		return fmt.Sprintf("%s mm   f %s   %s s   ISO %s", focalLength, fNumber, exposureTime, iso)

	case "device":
		makeInfo, ok1 := exifData["Make"].(string)
		model, ok2 := exifData["Model"].(string)

		if !ok1 || !ok2 {
			return "EXIF data incomplete"
		}

		return fmt.Sprintf("%s    %s", makeInfo, model)

	case "lens":
		if lens, ok := exifData["LensModel"].(string); ok {
			return lens
		}
		return "Lens data unavailable"

	case "time":
		if datetime, ok := exifData["DateTimeOriginal"].(string); ok {
			return datetime
		}
		return "Time data unavailable"
	}

	return ""
}

// resolveFontPath resolves a font name to a filesystem path
func resolveFontPath(fontName string) string {
	basePath, _ := os.Getwd()

	// Check if it's a known font
	if _, found := fontMap[fontName]; found {
		return path.Join(basePath, "font", fontName)
	}

	// Check if it's a direct file path
	if _, err := os.Stat(fontName); err == nil {
		return fontName
	}

	// Fall back to default font
	return path.Join(basePath, "font", "STHeiti Medium.ttc")
}

// createCanvasWithBorders creates a new image with white borders according to configuration
func createCanvasWithBorders(img image.Image, borderConfig Border) (canvas *image.RGBA, borderTopHeight, borderBottomHeight, borderWidth int) {
	imgWidth, imgHeight := img.Bounds().Dx(), img.Bounds().Dy()

	// Calculate border dimensions
	switch borderConfig.BorderLocation {
	case TOP:
		borderTopHeight = int(float64(imgHeight) * borderConfig.BorderTopHeight)
		canvas = image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight+borderTopHeight))
		draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(canvas, image.Rect(0, borderTopHeight, imgWidth, imgHeight+borderTopHeight),
			img, image.Point{0, 0}, draw.Over)

	case LEFT:
		borderWidth = int(float64(imgWidth) * borderConfig.BorderWidth)
		canvas = image.NewRGBA(image.Rect(0, 0, imgWidth+borderWidth, imgHeight))
		draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(canvas, image.Rect(borderWidth, 0, imgWidth+borderWidth, imgHeight),
			img, image.Point{0, 0}, draw.Src)

	case RIGHT:
		borderWidth = int(float64(imgWidth) * borderConfig.BorderWidth)
		canvas = image.NewRGBA(image.Rect(0, 0, imgWidth+borderWidth, imgHeight))
		draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(canvas, img.Bounds(), img, image.Point{0, 0}, draw.Src)

	case BOTTOM:
		borderBottomHeight = int(float64(imgHeight) * borderConfig.BorderBottomHeight)
		canvas = image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight+borderBottomHeight))
		draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(canvas, img.Bounds(), img, image.Point{}, draw.Over)

	case ALL:
		borderTopHeight = int(float64(imgHeight) * borderConfig.BorderTopHeight)
		borderBottomHeight = int(float64(imgHeight) * borderConfig.BorderBottomHeight)
		borderWidth = int(float64(imgWidth) * borderConfig.BorderWidth)
		canvas = image.NewRGBA(image.Rect(0, 0, imgWidth+borderWidth*2, imgHeight+borderTopHeight+borderBottomHeight))
		draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(canvas, image.Rect(borderWidth, borderTopHeight, imgWidth+borderWidth, imgHeight+borderTopHeight),
			img, image.Point{0, 0}, draw.Over)

	default:
		// Default to bottom border
		borderBottomHeight = int(float64(imgHeight) * borderConfig.BorderBottomHeight)
		canvas = image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight+borderBottomHeight))
		draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(canvas, img.Bounds(), img, image.Point{}, draw.Over)
	}

	return canvas, borderTopHeight, borderBottomHeight, borderWidth
}

// addLogoToCanvas adds a camera logo to the canvas if enabled
func addLogoToCanvas(dc *gg.Context, exifData ExifData, img image.Image, logoConfig Logo,
	borderWidth int, borderBottomHeight int, canvasHeight int) (float64, error) {
	basePath, _ := os.Getwd()

	// Determine logo path
	var logoPath string
	if logoConfig.FilePath == "default" {
		makeInfo, ok := exifData["Make"].(string)
		if !ok {
			return 0, fmt.Errorf("camera make not available in EXIF data")
		}
		logoPath = path.Join(basePath, "logo", makeInfo+".png")
	} else {
		logoPath = logoConfig.FilePath
	}

	// Load logo image
	logo, err := gg.LoadImage(logoPath)
	if err != nil {
		return 0, fmt.Errorf("failed to load logo: %w", err)
	}

	// Calculate resize factor
	var resizeFactor float64
	if logoConfig.Resize == "auto" {
		resizeFactor = 0.8
	} else {
		resizeFactor, err = strconv.ParseFloat(logoConfig.Resize, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid logo resize value: %w", err)
		}
	}

	// Resize logo
	targetHeight := int(float64(borderBottomHeight) * resizeFactor)
	targetWidth := logo.Bounds().Dx() * targetHeight / logo.Bounds().Dy()
	resizedLogo := resizeImage(logo, targetWidth, targetHeight)

	// Calculate position
	yPosition := float64(canvasHeight) - (float64(borderBottomHeight) * (1 - (1-resizeFactor)/2))

	// Draw logo
	dc.DrawImage(resizedLogo, borderWidth, int(yPosition))

	// Return width for text positioning
	return float64(resizedLogo.Bounds().Dx()) + float64(borderWidth), nil
}

// addTextOverlay adds a text overlay to the canvas based on configuration
func addTextOverlay(dc *gg.Context, textConfig Text, exifData ExifData, img image.Image,
	canvas *image.RGBA, borderWidth, borderTopHeight, borderBottomHeight int,
	logoEnabled bool, logoWidth float64) {

	// Get text content (from EXIF or direct)
	var text string
	if _, found := exifKeyMap[textConfig.Text]; found {
		text = formatExifText(exifData, textConfig.Text)
	} else {
		text = textConfig.Text
	}

	// Skip if text is empty
	if text == "" {
		return
	}

	// Resolve font and size
	fontPath := resolveFontPath(textConfig.FontPath)

	var fontSize float64
	if textConfig.FontSize == "auto" {
		fontSize = float64(borderBottomHeight) * 0.25
	} else {
		fontSize, _ = strconv.ParseFloat(textConfig.FontSize, 64)
	}

	// Set canvas dimensions
	canvasWidth, canvasHeight := canvas.Bounds().Dx(), canvas.Bounds().Dy()
	imgWidth := img.Bounds().Dx()

	// Position text based on location setting
	var x, y float64
	var alignment int // 0=left, 1=right, 2=center

	switch textConfig.Location {
	case "UpperRight":
		x = float64(imgWidth + borderWidth)
		y = float64(canvasHeight) - float64(borderBottomHeight)*3/5
		alignment = 1

	case "LowerRight":
		x = float64(imgWidth + borderWidth)
		y = float64(canvasHeight) - float64(borderBottomHeight)/5
		alignment = 1

	case "UpperLeft":
		if logoEnabled {
			x = logoWidth + 50
		} else {
			x = float64(borderWidth + 20)
		}
		y = float64(canvasHeight) - float64(borderBottomHeight)*3/5
		alignment = 0

	case "LowerLeft":
		if logoEnabled {
			x = logoWidth + 50
		} else {
			x = float64(borderWidth + 20)
		}
		y = float64(canvasHeight) - float64(borderBottomHeight)/5
		alignment = 0

	case "UpperCenter":
		x = float64(canvasWidth / 2)
		y = float64(canvasHeight) - float64(borderBottomHeight)*3/5
		alignment = 2

	case "LowerCenter":
		x = float64(canvasWidth / 2)
		y = float64(canvasHeight) - float64(borderBottomHeight)/5
		alignment = 2
	}

	// Draw text with optional bold effect
	drawText(dc, text, x, y, alignment, fontPath, fontSize, textConfig.Bold)
}

// drawText draws text on the canvas with optional bold effect
func drawText(dc *gg.Context, text string, x, y float64, alignment int, fontPath string, fontSize float64, bold int) {
	// Set color to black
	dc.SetRGB(0, 0, 0)

	// Load font
	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
		fmt.Println("Font loading error:", err)
		return
	}

	// Adjust position based on alignment and text width
	textWidth, _ := dc.MeasureString(text)
	switch alignment {
	case 1: // Right align
		x = x - textWidth - 20
	case 2: // Center align
		x = x - textWidth/2
	}

	// Apply bold effect if needed
	if bold > 0 {
		for dx := -bold; dx <= bold; dx++ {
			for dy := -bold; dy <= bold; dy++ {
				if dx != 0 || dy != 0 { // Skip center position to avoid overdraw
					dc.DrawString(text, x+float64(dx), y+float64(dy))
				}
			}
		}
	}

	// Draw main text
	dc.DrawString(text, x, y)
}

// resizeImage resizes an image while maintaining aspect ratio
func resizeImage(img image.Image, width, height int) image.Image {
	resizedImg := image.NewRGBA(image.Rect(0, 0, width, height))
	xdw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, img.Bounds(), xdw.Over, nil)
	return resizedImg
}

// saveImage saves an image to a file with specified JPEG quality
func saveImage(img image.Image, outputPath string, quality int) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	return jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
}
