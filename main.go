package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

/*
creator: ijwstl
bilibili: https://space.bilibili.com/352927833  KillNullPointer
email: ijwstl.coder@outlook.com
*/
const (
	BOTTOM string = "bottom"
	TOP    string = "top"
	LEFT   string = "left"
	RIGHT  string = "right"
	ALL    string = "all"
)

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
	Location string `json:"location"`
}

type Border struct {
	BorderWidth        float64 `json:"borderWidth"`
	BorderBottomHeight float64 `json:"borderBottomHeight"`
	BorderTopHeight    float64 `json:"borderTopHeight"`
	BorderLocation     string  `json:"borderLocation"`
	BorderColor        string  `json:"borderColor"`
}

type Config struct {
	ImagePath  string `json:"imagePath"`
	OutputPath string `json:"outputPath"`
	Quality    int    `json:"quality"`
	TaskMax    int    `json:"taskMax"`
	Border     Border `json:"border"`
	Logo       Logo   `json:"logo"`
	Text       []Text `json:"text"`
}

type Task struct {
	FilePath    string
	OutFilePath string
}

func checkFile(imagePath, outputPath string, wg *sync.WaitGroup, taskChannel *chan Task) error {
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
				err := checkFile(filepath.Join(imagePath, file.Name()), filepath.Join(outputPath, file.Name()), wg, taskChannel)
				if err != nil {
					return err
				}
			}

			if strings.HasSuffix(file.Name(), ".jpg") || strings.HasSuffix(file.Name(), ".png") {
				wg.Add(1)
				*taskChannel <- Task{FilePath: filepath.Join(imagePath, file.Name()), OutFilePath: filepath.Join(outputPath, file.Name())}
			}
		}
	} else if fileInfo.Mode().IsRegular() {
		outputFilePath := filepath.Join(outputPath, fileInfo.Name())
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			err := os.Mkdir(outputPath, os.ModePerm)
			if err != nil {
				fmt.Println("Error creating directory:", err)
				return err
			}
		}
		wg.Add(1)
		*taskChannel <- Task{FilePath: imagePath, OutFilePath: outputFilePath}
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

	var wg sync.WaitGroup

	// 使用一个通道动态发送任务
	taskChannel := make(chan Task, config.TaskMax)

	go func() {
		for task := range taskChannel {
			go func(task Task) {
				defer wg.Done() // Goroutine 完成后减少计数
				err := AddWhiteBorderWithText(task.FilePath, task.OutFilePath, config)
				if err != nil {
					return
				}
			}(task)
		}
	}()

	if err := checkFile(config.ImagePath, config.OutputPath, &wg, &taskChannel); err != nil {
		fmt.Println("Error:", err)
	}

	close(taskChannel)

	wg.Wait()

	fmt.Println("Image saved to:", config.OutputPath)
}
