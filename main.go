package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func parseArguments(arguments []string) (string, string) {
	if len(arguments) != 3 {
		fmt.Println("Wrong arguments")
		os.Exit(1)
	}

	return arguments[1], arguments[2]
}

func getRequestUrl(project string, document string) string {
	return fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents/%s", project, document)
}

func createResultFile() *os.File {
	file, err := os.Create("result")
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
	return file
}

func main() {
	fmt.Printf("Downloading...")

	project, document := parseArguments(os.Args)
	requestUrl := getRequestUrl(project, document)
	requestUrl += "?pageSize=1"
	resultFile := createResultFile()

	nextPageToken := ""
	for {
		if nextPageToken != "" {
			requestUrl += "&pageToken=" + nextPageToken
		}

		response, err := http.Get(requestUrl)
		if err != nil {
			fmt.Println("Error making GET request:", err)
			os.Exit(1)
		}

		if response.StatusCode != http.StatusOK {
			fmt.Printf("Failed request: %d %s\n", response.StatusCode, http.StatusText(response.StatusCode))
			os.Exit(1)
		}

		var jsonResponse struct {
			NextPageToken string            `json:"nextPageToken"`
			Documents     []json.RawMessage `json:"documents"`
		}

		err = json.NewDecoder(response.Body).Decode(&jsonResponse)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
			os.Exit(1)
		}

		for _, doc := range jsonResponse.Documents {
			var compactDoc bytes.Buffer
			err = json.Compact(&compactDoc, doc)
			if err != nil {
				fmt.Println("Error compacting JSON:", err)
				os.Exit(1)
			}

			_, err = resultFile.Write(compactDoc.Bytes())
			if err != nil {
				fmt.Println("Error writing to file:", err)
				os.Exit(1)
			}

			_, err = resultFile.WriteString("\n")
			if err != nil {
				fmt.Println("Error writing newline to file:", err)
				os.Exit(1)
			}
		}
		fmt.Printf(".")

		nextPageToken = jsonResponse.NextPageToken
		if nextPageToken == "" {
			fmt.Printf("\n")
			break
		}
	}
}
