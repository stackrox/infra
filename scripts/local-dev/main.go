package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

var localConfigurationDir = "configuration/"

var values = map[string]interface{}{
	"Charts": map[string]interface{}{
		"Annotations": map[string]interface{}{
			"ocpCredentialsMode": "Passthrough",
			"acsDemoVersion":     "4.3.1",
		},
	},
	"Values": map[string]interface{}{
		"testMode": true,
	},
}

func createLocalConfigurationDir() error {
	// remove previous configurations
	err := os.RemoveAll(localConfigurationDir)
	if err != nil {
		return fmt.Errorf("error while deleting the local configuration directory: %v", err)
	}

	// create clean configuration directory
	err = os.MkdirAll(localConfigurationDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error occurred creating the local configuration directory: %v", err)
	}

	return nil
}

func readFileToMap(path string) (map[string]string, error) {
	data := make(map[string]string)
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(fileContent, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getPathFromKey(key string) string {
	// replace __ with / for directories
	filepath := strings.Join(strings.Split(key, "__"), "/")
	// replace last _ with . for file ending
	index := strings.LastIndex(filepath, "_")
	if index != -1 {
		return filepath[:index] + "." + filepath[index+1:]
	}

	return filepath
}

func renderFile(path, content string, decodeString bool) error {
	configPath := fmt.Sprintf("configuration/%s", path)
	dir := filepath.Dir(configPath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error while creating directories: %v", err)
	}

	if decodeString {
		decodedContent, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return fmt.Errorf("error while decoding base64 content: %v", err)
		}
		content = string(decodedContent)
	}

	// Use a simple template engine to render the template
	tmpl, err := template.New("template").Parse(content)
	if err != nil {
		return fmt.Errorf("error while parsing the template: %v", err)
	}

	outputFile, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("error while creating the output file: %v", err)
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, values)
	if err != nil {
		return fmt.Errorf("An error occurred while rendering the template: %v", err)
	}

	return nil
}

func renderFlavors() error {
	path := "flavors.yaml"
	fileContent, err := ioutil.ReadFile(fmt.Sprintf("chart/infra-server/static/%s", path))
	if err != nil {
		return err
	}
	return renderFile(path, string(fileContent), false)
}

func main() {
	if err := createLocalConfigurationDir(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	data, err := readFileToMap("chart/infra-server/configuration/development-values-from-files.yaml")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for key, content := range data {
		filepath := getPathFromKey(key)
		err := renderFile(filepath, content, true)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", filepath, err)
			return
		} else {
			fmt.Println("Created", filepath)
		}
	}

	if err := renderFlavors(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	} else {
		fmt.Println("Created flavors.yaml")
	}
}
