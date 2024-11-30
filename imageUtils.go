package main

import (
	"fmt"
	"github.com/fogleman/gg"
	"github.com/rwcarlsen/goexif/exif"
	xdw "golang.org/x/image/draw"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path"
	"strconv"
	"time"
)

var keySet map[string]struct{} = map[string]struct{}{
	"expose": struct{}{},
	"device": struct{}{},
	"lens":   struct{}{},
	"time":   struct{}{},
}

var fontNameSet map[string]struct{} = map[string]struct{}{
	"SFCamera.ttf":        struct{}{},
	"SFCompactItalic.ttf": struct{}{},
	"NewYork.ttf":         struct{}{},
	"STHeiti Medium.ttc":  struct{}{},
}

func getExifData(imgPath string) (map[string]interface{}, error) {
	exifData := make(map[string]interface{})

	file, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	x, err := exif.Decode(file)

	if err != nil {
		return nil, err
	}

	// 读取常用的 EXIF 信息
	if focal, err := x.Get(exif.FocalLength); err == nil {
		num, den, _ := focal.Rat2(0)
		exifData["FocalLength"] = strconv.FormatInt(num/den, 10)
	}
	if fNum, err := x.Get(exif.FNumber); err == nil {
		num, den, _ := fNum.Rat2(0)
		exifData["FNumber"] = strconv.FormatFloat(float64(num)/float64(den), 'g', -1, 64)
	}
	if exposure, err := x.Get(exif.ExposureTime); err == nil {
		num, den, _ := exposure.Rat2(0)
		exifData["ExposureTime"] = strconv.FormatInt(num, 10) + "/" + strconv.FormatInt(int64(den), 10)
	}
	if iso, err := x.Get(exif.ISOSpeedRatings); err == nil {
		exifData["ISOSpeedRatings"] = strconv.FormatUint(uint64(iso.Val[0]), 10)
	}
	if date, err := x.Get(exif.DateTimeOriginal); err == nil {
		dt, _ := date.StringVal()
		if t, err := time.Parse("2006:01:02 15:04:05", dt); err == nil {
			exifData["DateTimeOriginal"] = t.Format("2006-01-02 15:04:05")
		}
	}
	if make, err := x.Get(exif.Make); err == nil {
		exifData["Make"], _ = make.StringVal()
	}
	if model, err := x.Get(exif.Model); err == nil {
		exifData["Model"], _ = model.StringVal()
	}
	if lensModel, err := x.Get(exif.LensModel); err == nil {
		exifData["LensModel"], _ = lensModel.StringVal()
	}

	return exifData, nil
}

func getYouWantExifInfo(exifData map[string]interface{}, key string) string {
	if key == "expose" {
		focalLength, fNumber, exposureTime, iso := exifData["FocalLength"].(string), exifData["FNumber"].(string), exifData["ExposureTime"].(string), exifData["ISOSpeedRatings"].(string)
		return fmt.Sprintf("%s mm   f %s   %s s   ISO %s", focalLength, fNumber, exposureTime, iso)
	} else if key == "device" {
		makeInfo, model := exifData["Make"].(string), exifData["Model"].(string)
		return fmt.Sprintf("%s    %s", makeInfo, model)
	} else if key == "lens" {
		return exifData["LensModel"].(string)
	} else if key == "time" {
		return exifData["DateTimeOriginal"].(string)
	}
	return ""
}

func getYouWantFont(fontName string) string {
	basePath, _ := os.Getwd()
	_, found := fontNameSet[fontName]

	if found {
		return path.Join(basePath, "font", fontName)
	}

	_, err := os.Stat(fontName)

	if err == nil {
		return fontName
	}

	return path.Join(basePath, "font", "STHeiti Medium.ttc")

}

func addTextToImage(dc *gg.Context, text string, x, y float64, i int, fontPath string, fontSize float64, bold int) {
	dc.SetRGB(0, 0, 0)
	err := dc.LoadFontFace(fontPath, fontSize)
	if err != nil {
		fmt.Println(err)
		return
	}
	w, _ := dc.MeasureString(text)

	if i == 1 {
		x = x - w - 20
	} else if i == 2 {
		x = x - w/2
	}

	for dx := 0 - bold; dx <= 1; dx++ {
		for dy := 0 - bold; dy <= 1; dy++ {
			if dx != 0 || dy != 0 { // 不在原点绘制，防止重复
				dc.DrawString(text, x+float64(dx), y+float64(dy))
			}
		}
	}
}

func addLogo(canvas *gg.Context, logo image.Image, x, y float64) {
	canvas.DrawImage(logo, int(x), int(y))
}

func resizeImage(img image.Image, width, height int) image.Image {
	// 创建一个空白的目标图像，指定新的宽度和高度
	resizedImg := image.NewRGBA(image.Rect(0, 0, width, height))
	// 使用 draw 包进行缩放
	xdw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, img.Bounds(), xdw.Over, nil)
	return resizedImg
}

func AddWhiteBorderWithText(imgPath, outputPath string, config Config) error {
	fmt.Println("处理图片: ", imgPath)
	basePath, _ := os.Getwd()
	imgFile, err := os.Open(imgPath)
	if err != nil {
		return err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return err
	}

	exifData, err := getExifData(imgPath)
	if err != nil {
		fmt.Println("Get exif info error:", err)
		return err
	}

	var newHeight, newWidth, borderTopHeight, borderBottomHeight, borderWidth int
	var border *image.RGBA

	switch config.Border.BorderLocation {
	case TOP:
		borderTopHeight = int(float64(img.Bounds().Dy()) * config.Border.BorderTopHeight)
		newHeight = img.Bounds().Dy() + borderTopHeight
		newWidth = img.Bounds().Dx()

		border = image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.Draw(border, border.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(border, image.Rect(0, borderTopHeight, newWidth, newHeight), img, image.Point{0, 0}, draw.Over)
	case LEFT:
		borderWidth = int(float64(img.Bounds().Dx()) * config.Border.BorderWidth)
		newHeight = img.Bounds().Dy()
		newWidth = img.Bounds().Dx() + borderWidth
		border = image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.Draw(border, border.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		//draw.Draw(border, img.Bounds(), img, image.Point{borderWidth, 0}, draw.Over)
		draw.Draw(border, image.Rect(borderWidth, 0, newWidth, newHeight), img, image.Point{0, 0}, draw.Src)
	case RIGHT:
		borderWidth = int(float64(img.Bounds().Dx()) * config.Border.BorderWidth)
		newHeight = img.Bounds().Dy()
		newWidth = img.Bounds().Dx() + borderWidth
		border = image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.Draw(border, border.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(border, img.Bounds(), img, image.Point{0, 0}, draw.Src)
	case BOTTOM:
		borderBottomHeight = int(float64(img.Bounds().Dy()) * config.Border.BorderBottomHeight)
		newHeight = img.Bounds().Dy() + borderBottomHeight
		newWidth = img.Bounds().Dx()
		border = image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.Draw(border, border.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(border, img.Bounds(), img, image.Point{}, draw.Over)
	case ALL:
		borderTopHeight = int(float64(img.Bounds().Dy()) * config.Border.BorderTopHeight)
		borderBottomHeight = int(float64(img.Bounds().Dy()) * config.Border.BorderBottomHeight)
		borderWidth = int(float64(img.Bounds().Dx()) * config.Border.BorderWidth)
		newHeight = img.Bounds().Dy() + borderTopHeight + borderBottomHeight
		newWidth = img.Bounds().Dx() + borderWidth*2
		border = image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.Draw(border, border.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(border, image.Rect(borderWidth, borderTopHeight, newWidth, newHeight), img, image.Point{0, 0}, draw.Over)
	default:
		borderBottomHeight = int(float64(img.Bounds().Dy()) * config.Border.BorderBottomHeight)
		newHeight = img.Bounds().Dy() + borderBottomHeight
		newWidth = img.Bounds().Dx()
		border = image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.Draw(border, border.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(border, img.Bounds(), img, image.Point{}, draw.Over)
	}

	dc := gg.NewContextForImage(border)

	// 添加Logo
	logoConfig := config.Logo
	var logoWidth float64 = 0

	if logoConfig.On {
		var logoPath string
		if logoConfig.FilePath == "default" {
			logoPath = path.Join(basePath, "logo", exifData["Make"].(string)+".png")
		} else {
			logoPath = logoConfig.FilePath
		}
		logo, err := gg.LoadImage(logoPath)
		if err != nil {
			fmt.Println("可能是因为Exif存储相机厂商未预制, 厂商：", exifData["Make"])
			fmt.Println("无法加载 logo 图片: ", err)
			return err
		}

		var logoResize float64
		if logoConfig.Resize == "auto" {
			logoResize = 0.8
		} else {
			logoResize, err = strconv.ParseFloat(logoConfig.Resize, 64)
			if err != nil {
				fmt.Println("Logo 缩放配置异常:", err)
				return err
			}
		}
		logo = resizeImage(logo, logo.Bounds().Dx()*int(float64(borderBottomHeight)*logoResize)/logo.Bounds().Dy(), int(float64(borderBottomHeight)*logoResize))

		logoWidth = float64(logo.Bounds().Dx()) + float64(borderWidth)

		addLogo(dc, logo, float64(borderWidth), float64(newHeight)-(float64(borderBottomHeight)*(1-(1-logoResize)/2)))
	}

	textConfigList := config.Text
	for _, textConfig := range textConfigList {
		var imageText string
		var textFontSize float64
		var textFont string

		_, found := keySet[textConfig.Text]
		if found {
			imageText = getYouWantExifInfo(exifData, textConfig.Text)
		} else {
			imageText = textConfig.Text
		}

		textFont = getYouWantFont(textConfig.FontPath)

		if textConfig.FontSize == "auto" {
			textFontSize = float64(borderBottomHeight) * 0.25
		} else {
			textFontSize, _ = strconv.ParseFloat(textConfig.FontSize, 64)
		}

		if textConfig.Location == "UpperRight" && textConfig.On {
			addTextToImage(dc, imageText, float64(img.Bounds().Dx()+borderWidth), float64(newHeight)-float64(borderBottomHeight)*3/5, 1, textFont, textFontSize, textConfig.Bold)
		} else if textConfig.Location == "LowerRight" && textConfig.On {
			addTextToImage(dc, imageText, float64(img.Bounds().Dx()+borderWidth), float64(newHeight)-float64(borderBottomHeight)/5, 1, textFont, textFontSize, textConfig.Bold)
		} else if textConfig.Location == "UpperLeft" && textConfig.On {
			if !config.Logo.On {
				addTextToImage(dc, imageText, float64(borderWidth+20), float64(newHeight)-float64(borderBottomHeight)*3/5, 0, textFont, textFontSize, textConfig.Bold)
			} else {
				addTextToImage(dc, imageText, logoWidth+50, float64(newHeight)-float64(borderBottomHeight)*3/5, 0, textFont, textFontSize, textConfig.Bold)
			}
		} else if textConfig.Location == "LowerLeft" && textConfig.On {
			if !config.Logo.On {
				addTextToImage(dc, imageText, float64(borderWidth+20), float64(newHeight)-float64(borderBottomHeight)*1/5, 0, textFont, textFontSize, textConfig.Bold)
			} else {
				addTextToImage(dc, imageText, logoWidth+50, float64(newHeight)-float64(borderBottomHeight)*1/5, 0, textFont, textFontSize, textConfig.Bold)
			}
		} else if textConfig.Location == "UpperCenter" && textConfig.On {
			addTextToImage(dc, imageText, float64(newWidth/2), float64(newHeight)-float64(borderBottomHeight)*3/5, 2, textFont, textFontSize, textConfig.Bold)
		} else if textConfig.Location == "LowerCenter" && textConfig.On {
			addTextToImage(dc, imageText, float64(newWidth/2), float64(newHeight)-float64(borderBottomHeight)*1/5, 2, textFont, textFontSize, textConfig.Bold)
		}
	}

	// ----------------------------

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return jpeg.Encode(outFile, dc.Image(), &jpeg.Options{Quality: config.Quality})
}
