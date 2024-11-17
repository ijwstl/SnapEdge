package main

import (
	"encoding/json"
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
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*
	creator: ijwstl
	bilibili: https://space.bilibili.com/352927833  KillNullPointer
	email: ijwstl.coder@outlook.com
*/

type Logo struct {
	On       bool   `json:"on"`
	FilePath string `json:"filePath"`
	Resize   string `json:"resize"`
}

type Text struct {
	On       bool   `json:"on"`
	Text     string `json:"text"`
	FontPath string `json:"fontPath"`
	FontSize string `json:"fontSize"`
	Bold     int    `json:"bold"`
}

type Config struct {
	ImagePath   string  `json:"imagePath"`
	OutputPath  string  `json:"outputPath"`
	Quality     int     `json:"quality"`
	BorderWidth float64 `json:"borderWidth"`
	Logo        Logo    `json:"logo"`
	UpperLeft   Text    `json:"upperLeft"`
	LowerLeft   Text    `json:"lowerLeft"`
	UpperRight  Text    `json:"upperRight"`
	LowerRight  Text    `json:"lowerRight"`
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

func addWhiteBorderWithText(imgPath, outputPath string, config Config) error {
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

	borderHeight := int(float64(img.Bounds().Dy()) * config.BorderWidth)
	newHeight := img.Bounds().Dy() + borderHeight
	newWidth := img.Bounds().Dx()

	// 创建带白色边框的图像
	border := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.Draw(border, border.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	draw.Draw(border, img.Bounds(), img, image.Point{}, draw.Over)

	dc := gg.NewContextForImage(border)

	// 添加曝光信息
	upperRightConfig := config.UpperRight

	if upperRightConfig.On {
		var upperRightText string
		if upperRightConfig.Text == "default" {
			focalLength, fNumber, exposureTime, iso := exifData["FocalLength"].(string), exifData["FNumber"].(string), exifData["ExposureTime"].(string), exifData["ISOSpeedRatings"].(string)
			upperRightText = fmt.Sprintf("%s mm   f %s   %s s   ISO %s", focalLength, fNumber, exposureTime, iso)
		} else {
			upperRightText = upperRightConfig.Text
		}

		var upperRightTextFont string
		if upperRightConfig.FontPath == "default" {
			upperRightTextFont = path.Join(basePath, "font", "SFCompactItalic.ttf")
		} else {
			upperRightTextFont = upperRightConfig.FontPath
		}

		var upperRightFontSize float64
		if upperRightConfig.FontSize == "auto" {
			upperRightFontSize = float64(borderHeight) * 0.25
		} else {
			upperRightFontSize, _ = strconv.ParseFloat(upperRightConfig.FontSize, 64)
		}

		addTextToImage(dc, upperRightText, float64(newWidth), float64(newHeight)-float64(borderHeight)*3/5, 1, upperRightTextFont, upperRightFontSize, upperRightConfig.Bold)
	}

	// ----------------------------

	// 添加拍摄时间
	lowerRightConfig := config.LowerRight

	if lowerRightConfig.On {
		var lowerRightText string
		if lowerRightConfig.Text == "default" {
			lowerRightText = exifData["DateTimeOriginal"].(string)
		} else {
			lowerRightText = lowerRightConfig.Text
		}

		var lowerRightTextFont string
		if lowerRightConfig.FontPath == "default" {
			lowerRightTextFont = path.Join(basePath, "font", "SFCamera.ttf")
		} else {
			lowerRightTextFont = lowerRightConfig.FontPath
		}

		var lowerRightFontSize float64
		if lowerRightConfig.FontSize == "auto" {
			lowerRightFontSize = float64(borderHeight) * 0.25
		} else {
			lowerRightFontSize, _ = strconv.ParseFloat(lowerRightConfig.FontSize, 64)
		}

		addTextToImage(dc, lowerRightText, float64(newWidth), float64(newHeight)-float64(borderHeight)/5, 1, lowerRightTextFont, lowerRightFontSize, lowerRightConfig.Bold)
	}

	// ----------------------------

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
		logo = resizeImage(logo, logo.Bounds().Dx()*int(float64(borderHeight)*logoResize)/logo.Bounds().Dy(), int(float64(borderHeight)*logoResize))

		logoWidth = float64(logo.Bounds().Dx())

		addLogo(dc, logo, 0, float64(newHeight)-(float64(borderHeight)*(1-(1-logoResize)/2)))
	}

	// ----------------------------

	// 添加设备信息
	upperLeftConfig := config.UpperLeft

	if upperLeftConfig.On {
		var upperLeftText string
		if upperLeftConfig.Text == "default" {
			makeInfo, model := exifData["Make"].(string), exifData["Model"].(string)
			upperLeftText = fmt.Sprintf("%s    %s", makeInfo, model)
		} else {
			upperLeftText = upperLeftConfig.Text
		}

		var upperLeftTextFont string
		if upperLeftConfig.FontPath == "default" {
			upperLeftTextFont = path.Join(basePath, "font", "SFCompactItalic.ttf")
		} else {
			upperLeftTextFont = upperLeftConfig.FontPath
		}

		var upperLeftFontSize float64
		if upperLeftConfig.FontSize == "auto" {
			upperLeftFontSize = float64(borderHeight) * 0.25
		} else {
			upperLeftFontSize, _ = strconv.ParseFloat(upperLeftConfig.FontSize, 64)
		}

		addTextToImage(dc, upperLeftText, logoWidth+50, float64(newHeight)-float64(borderHeight)*3/5, 0, upperLeftTextFont, upperLeftFontSize, upperLeftConfig.Bold)
	}

	// ----------------------------

	// 添加镜头信息
	lowerLeftConfig := config.LowerLeft

	if lowerLeftConfig.On {
		var lowerLeftText string
		if lowerLeftConfig.Text == "default" {
			lowerLeftText = exifData["LensModel"].(string)
		} else {
			lowerLeftText = lowerLeftConfig.Text
		}

		var lowerLeftTextFont string
		if lowerLeftConfig.FontPath == "default" {
			lowerLeftTextFont = path.Join(basePath, "font", "SFCompactItalic.ttf")
		} else {
			lowerLeftTextFont = lowerLeftConfig.FontPath
		}

		var lowerLeftFontSize float64
		if lowerLeftConfig.FontSize == "auto" {
			lowerLeftFontSize = float64(borderHeight) * 0.25
		} else {
			lowerLeftFontSize, _ = strconv.ParseFloat(lowerLeftConfig.FontSize, 64)
		}

		addTextToImage(dc, lowerLeftText, logoWidth+50, float64(newHeight)-float64(borderHeight)*1/5, 0, lowerLeftTextFont, lowerLeftFontSize, lowerLeftConfig.Bold)
	}

	// ----------------------------

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return jpeg.Encode(outFile, dc.Image(), &jpeg.Options{Quality: config.Quality})
}

func addWhiteBorderWithTextWrapper(imagePath, outputPath string, config Config) error {
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
				err := addWhiteBorderWithTextWrapper(filepath.Join(imagePath, file.Name()), filepath.Join(outputPath, file.Name()), config)
				if err != nil {
					return err
				}
			}

			if strings.HasSuffix(file.Name(), ".jpg") || strings.HasSuffix(file.Name(), ".png") {
				addWhiteBorderWithText(filepath.Join(imagePath, file.Name()), filepath.Join(outputPath, file.Name()), config)
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
		addWhiteBorderWithText(imagePath, outputFilePath, config)
	} else {
		fmt.Println("Error: File not found")
	}

	return nil
}

func readConfig() (Config, error) {
	// 读取文件内容
	bytes, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return Config{}, err
	}

	// 解析JSON
	var config Config

	if err := json.Unmarshal(bytes, &config); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return Config{}, err
	}
	return config, nil
}

func main() {

	config, err := readConfig()
	if err != nil {
		fmt.Println("Error reading config:", err)
	}

	fmt.Println("Config info:", config)

	if err := addWhiteBorderWithTextWrapper(config.ImagePath, config.OutputPath, config); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Image saved to:", config.OutputPath)
	}
}
