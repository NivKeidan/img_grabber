package main

// TODO:
// add parrallelism for image downloading

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
	"regexp"
	"strings"
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

func resetDebugObject() {
	debugObject = &DebugObject{
		Input: "",
		CurrentPhase: "",
		Html: "",
		DownloadUrl: "",
		Error: "",
	}
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

func doExit(statusCode int) {
	fmt.Println("shtok shtok")
	time.Sleep(1* time.Second)
	os.Exit(statusCode)
}

func main() {
	for {
		resetDebugObject()
		getInput()

		if inputIsExit() {
			doExit(0)
		}

		t := time.Now()
		fileNamePrefix := fmt.Sprintf("%d_%d_%d", t.Hour(), t.Minute(), t.Second())

		basePath := extractBasePathForImagesFromUrl()
		getHtmlResponse()

		imgPaths := extractImgPathsFromHtml()

		fmt.Printf("found %d images\n", len(imgPaths))

		for i, imgPath := range imgPaths {
			fileName := fmt.Sprintf("%s-%d", fileNamePrefix, i+1)
			finalPath := fmt.Sprintf("%s%s", basePath, imgPath)
			downloadFile(finalPath, fileName)
			fmt.Printf("downloaded image %d as %s\n", i+1, fileName)
		}
		fmt.Println()
	}
	
}

func downloadFile(URL, fileName string) {
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
	file, err := os.Create(fileName)
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