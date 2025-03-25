package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

var (
	localConfigurationDir = "configuration/"
	chartPath             = "chart/infra-server/Chart.yaml"
	staticPath            = "chart/infra-server/static"
	valuesPath            = "chart/infra-server/configuration/development-values-from-files.yaml"
)

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
	fileContent, err := os.ReadFile(path)
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

func determineValues() (map[string]interface{}, error) {
	values := map[string]interface{}{
		"Values": map[string]interface{}{
			"testMode": true,
		},
	}

	data := make(map[string]interface{})
	fileContent, err := os.ReadFile(chartPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(fileContent, &data)
	if err != nil {
		return nil, err
	}

	values["Chart"] = map[string]interface{}{
		"Annotations": data["annotations"],
	}

	return values, nil
}

func renderFile(path, content string, decodeString bool) error {
	configPath := fmt.Sprintf("%s%s", localConfigurationDir, path)
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
		return fmt.Errorf("error while parsing the template at %s: %v", path, err)
	}

	outputFile, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("error while creating the output file: %v", err)
	}
	defer outputFile.Close()

	values, err := determineValues()
	if err != nil {
		return fmt.Errorf("An error occurred while determining values: %v", err)
	}
	err = tmpl.Execute(outputFile, values)
	if err != nil {
		return fmt.Errorf("An error occurred while rendering the template: %v", err)
	}

	return nil
}

func renderFlavorList() error {
	file := "flavors.yaml"
	fileContent, err := os.ReadFile(fmt.Sprintf("%s/%s", staticPath, file))
	if err != nil {
		return err
	}
	return renderFile(file, string(fileContent), false)
}

func renderWorkflows() error {
	files, err := os.ReadDir(staticPath)
	if err != nil {
		return fmt.Errorf("error while looking for workflow files: %v", err)
	}
	for _, file := range files {
		fileName := file.Name()
		if (strings.HasPrefix(fileName, "test-") || strings.HasPrefix(fileName, "workflow-")) && strings.HasSuffix(fileName, ".yaml") {
			fileContent, err := os.ReadFile(fmt.Sprintf("%s/%s", staticPath, fileName))
			if err != nil {
				return err
			}
			if err := renderFile(fileName, string(fileContent), false); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	if err := createLocalConfigurationDir(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	data, err := readFileToMap(valuesPath)
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

	if err := renderFlavorList(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	} else {
		fmt.Println("Created flavors.yaml")
	}

	if err := renderWorkflows(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	} else {
		fmt.Println("Rendered workflows")
	}
}
