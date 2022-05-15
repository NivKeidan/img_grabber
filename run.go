package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

type DebugObject struct {
	Input string
	CurrentPhase string
	Html string
	DownloadUrl string
	Error string
}

var debugObject *DebugObject

func main() {
	for {
		resetDebugObject()
		getInput()

		if inputIsExit() {
			doExit(0)
		}

		basePath := extractBasePathForImagesFromUrl()
		getHtmlResponse()

		imgPaths := extractImgPathsFromHtml()

		fmt.Printf("found %d images. Enter name: ", len(imgPaths))
		var wg sync.WaitGroup
		wg.Add(len(imgPaths))
		inputName := getInputName()
		createDir(inputName)

		for i, imgPath := range imgPaths {
			imgSuffix := getImageSuffix(imgPath)
			fileName := fmt.Sprintf("%d.%s", i+1, imgSuffix)
			filePath := path.Join(inputName, fileName)
			
			finalUrl := fmt.Sprintf("%s%s", basePath, imgPath)
			go func(url string, filePath string, ind int) {
				downloadFile(url, filePath)
				fmt.Printf("downloaded image %d\n", ind+1)
				wg.Done()
			}(finalUrl, filePath, i)
		}
		wg.Wait()
		fmt.Println()
	}
}

func resetDebugObject() {
	debugObject = &DebugObject{
		Input: "",
		CurrentPhase: "",
		Html: "",
		DownloadUrl: "",
		Error: "",
	}
}

func getInput() {
	debugObject.CurrentPhase = "getInput-scan"
	fmt.Print("paste url (or q to exit): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if scanner.Err() != nil {
		errorOut(scanner.Err())
	}
	debugObject.CurrentPhase = "getInput-parse"
	debugObject.Input = scanner.Text()
	if scanner.Err() != nil {
		errorOut(scanner.Err())
	}
	
	if debugObject.Input == "" {
		errorOut(fmt.Errorf("input is empty"))
	}
}

func inputIsExit() bool {
	return strings.Compare("q", debugObject.Input) == 0
}

func extractBasePathForImagesFromUrl() string {
	debugObject.CurrentPhase = "extractBasePathForImagesFromUrl"
	inputUrl, err := url.Parse(debugObject.Input)
	if err != nil {
		errorOut(err)
	}

	path := inputUrl.Path
	if path == "" {
		errorOut(fmt.Errorf("invalid input"))
	}
	r := regexp.MustCompile(`/(.*)*/`)
	match2 := r.FindStringSubmatch(path)
	if len(match2) == 0{
		errorOut(fmt.Errorf("no / in path"))
	}

	return fmt.Sprintf("%s://%s%s", inputUrl.Scheme, inputUrl.Host, match2[0])
}

func getInputName() string {
	debugObject.CurrentPhase = "getInputName"
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if scanner.Err() != nil {
		errorOut(scanner.Err())
	}
	debugObject.CurrentPhase = "getInputName-parse"
	input := scanner.Text()
	if scanner.Err() != nil {
		errorOut(scanner.Err())
	}
	
	if input == "" {
		input = getDefaultName()
	}
	return input
}

func getDefaultName() string {
	return time.Now().Format("2006-01-02T15-04-05")
}

func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}

func createDir(dirName string) {
	debugObject.CurrentPhase = "createDir"
	exists, err := exists(dirName)
	if err != nil {
		errorOut(err)
	}
	if exists {
		errorOut(fmt.Errorf("name already exists"))
	}
	os.Mkdir(dirName, os.ModePerm)
}

func getImageSuffix(imgUrl string) string {
	debugObject.CurrentPhase = "getImageSuffix"
	splitted := strings.Split(imgUrl, ".")
	if len(splitted) < 2 {
		errorOut(fmt.Errorf("cant extract img suffix from %s", imgUrl))
	}
	return splitted[len(splitted)-1]
}

func getHtmlResponse() {
	debugObject.CurrentPhase = "getHtmlResponse"
	fmt.Println("getting HTML...")
	resp, err := http.Get(debugObject.Input)
	if err != nil {
		errorOut(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errorOut(fmt.Errorf("resp status code %d", resp.StatusCode))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorOut(err)
	}

	debugObject.Html = string(bodyBytes)
}

func extractImgPathsFromHtml() []string {
	debugObject.CurrentPhase = "extractImgPathsFromHtml"
	fmt.Println("extracting img tags from HTML...")
	imgPaths := make([]string, 0)

	r := regexp.MustCompile(`<img src="(.*?)"`)
	match := r.FindAllStringSubmatch(debugObject.Html, -1)
	if len(match) < 1 {
		errorOut(fmt.Errorf("no tags found in html"))
	}
	for _, foundImgPath := range match {
		imgPathWithParams := string(foundImgPath[1])
		imgPath := strings.Split(imgPathWithParams, "?")[0]
		imgPaths = append(imgPaths, imgPath)
	}

	return imgPaths
}

func downloadFile(URL, filePath string) {
	debugObject.DownloadUrl = URL
	debugObject.CurrentPhase = "downloadFile"

	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		errorOut(err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		errorOut(errors.New("Received non 200 response code"))
	}
	//Create a empty file
	file, err := os.Create(filePath)
	if err != nil {
		errorOut(err)
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		errorOut(err)
	}
}

func doExit(statusCode int) {
	fmt.Println("shtok shtok")
	time.Sleep(1* time.Second)
	os.Exit(statusCode)
}

func errorOut(err error) {
	debugObject.Error = err.Error()
	ts := time.Now().Unix()
	fileName := fmt.Sprintf("error_%d.json", ts)

	file, _ := json.MarshalIndent(debugObject, "", " ")
	_ = ioutil.WriteFile(fileName, file, 0644)

	fmt.Println("Error occurred!")
	fmt.Println("Error details are in file:", fileName)
	doExit(1)
}