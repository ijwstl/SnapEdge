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
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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

func addTextToImage(dc *gg.Context, text string, x, y float64, i int, fontPath string, fontSize float64, bold int) {
	dc.SetRGB(0, 0, 0)
	err := dc.LoadFontFace(fontPath, fontSize)
	if err != nil {
		fmt.Println(err)
		return
	}
	w, _ := dc.MeasureString(text)

	if i == 1 {
		x = x - w - 50
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

func addWhiteBorderWithText(imgPath, outputPath string) error {
	basePath, _ := os.Getwd()
	SFCompactItalicFontPath := path.Join(basePath, "font", "SFCompactItalic.ttf")
	NewYorkFontPath := path.Join(basePath, "font", "SFCamera.ttf")
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

	borderHeight := int(float64(img.Bounds().Dy()) * 0.07)
	newHeight := img.Bounds().Dy() + borderHeight
	newWidth := img.Bounds().Dx()

	// 创建带白色边框的图像
	border := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.Draw(border, border.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(border, img.Bounds(), img, image.Point{}, draw.Over)

	dc := gg.NewContextForImage(border)

	// 添加曝光信息
	focalLength, fNumber, exposureTime, iso := exifData["FocalLength"].(string), exifData["FNumber"].(string), exifData["ExposureTime"].(string), exifData["ISOSpeedRatings"].(string)
	exposeInfo := fmt.Sprintf("%s mm   f %s   %s s   ISO %s", focalLength, fNumber, exposureTime, iso)
	addTextToImage(dc, exposeInfo, float64(newWidth), float64(newHeight)-float64(borderHeight)*3/5, 1, SFCompactItalicFontPath, float64(borderHeight)*0.25, 0)
	// ----------------------------

	// 添加拍摄时间
	shotTime, _ := exifData["DateTimeOriginal"].(string)
	addTextToImage(dc, shotTime, float64(newWidth), float64(newHeight)-float64(borderHeight)/5, 1, NewYorkFontPath, float64(borderHeight)*0.25, 0)
	// ----------------------------

	// 添加Logo
	logoPath := path.Join(basePath, "logo", exifData["Make"].(string)+".png")
	logo, err := gg.LoadImage(logoPath)
	if err != nil {
		log.Fatalf("无法加载 logo 图片: %v", err)
	}

	logo = resizeImage(logo, logo.Bounds().Dx()*int(float64(borderHeight)*0.9)/logo.Bounds().Dy(), int(float64(borderHeight)*0.9))

	logoWidth := float64(logo.Bounds().Dx())

	addLogo(dc, logo, 0, float64(newHeight)-(float64(borderHeight)*0.95))
	// ----------------------------

	// 添加设备信息
	makeInfo, model := exifData["Make"].(string), exifData["Model"].(string)
	deviceInfo := fmt.Sprintf("%s    %s", makeInfo, model)
	addTextToImage(dc, deviceInfo, logoWidth+50, float64(newHeight)-float64(borderHeight)*3/5, 0, SFCompactItalicFontPath, float64(borderHeight)*0.25, 0)
	// ----------------------------

	// 添加镜头信息
	lens := exifData["LensModel"].(string)
	addTextToImage(dc, lens, logoWidth+50, float64(newHeight)-float64(borderHeight)*1/5, 0, SFCompactItalicFontPath, float64(borderHeight)*0.25, 0)
	// ----------------------------

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return jpeg.Encode(outFile, dc.Image(), &jpeg.Options{Quality: 100})
}

func addWhiteBorderWithTextWrapper(imagePath, outputPath string) error {
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		err := os.Mkdir(outputPath, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return err
		}
	}

	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	if fileInfo.IsDir() {
		files, err := os.ReadDir(imagePath)
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return err
		}

		for _, file := range files {
			if file.IsDir() {
				err := addWhiteBorderWithTextWrapper(filepath.Join(imagePath, file.Name()), filepath.Join(outputPath, file.Name()))
				if err != nil {
					return err
				}
			}

			if strings.HasSuffix(file.Name(), ".jpg") || strings.HasSuffix(file.Name(), ".png") {
				addWhiteBorderWithText(filepath.Join(imagePath, file.Name()), filepath.Join(outputPath, file.Name()))
			}
		}
	} else if fileInfo.Mode().IsRegular() {
		outputFilePath := filepath.Join(filepath.Dir(outputPath), fileInfo.Name())
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			err := os.Mkdir(outputPath, os.ModePerm)
			if err != nil {
				fmt.Println("Error creating directory:", err)
				return err
			}
		}
		addWhiteBorderWithText(imagePath, outputFilePath)
	} else {
		fmt.Println("Error: File not found")
	}

	return nil
}

func main() {

	imagePath := "/Users/wangqi/Desktop/2.35"
	outputPath := "/Users/wangqi/Desktop/tt"
	if err := addWhiteBorderWithTextWrapper(imagePath, outputPath); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Image saved to:", outputPath)
	}
}
