package parse

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/Aman1143/reverse-proxy/src/configschema"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

func ParaseYAMLConfig(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("error in reading file", err)
	}
	defer file.Close()

	var content string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}

	var yamlData map[string]interface{}
	err = yaml.Unmarshal([]byte(content), &yamlData)
	if err != nil {
		log.Fatal("error in unmarshaling YAML:", err)
	}

	jsonBytes, err := json.Marshal(yamlData)
	if err != nil {
		log.Fatal("error in marshaling to JSON:", err)
	}

	return string(jsonBytes)
}

func ValidateConfig(configStr string) configschema.RootConfigSchema {
	var config configschema.RootConfigSchema

	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		log.Fatal("Error parsing JSON:", err)
	}

	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		log.Fatal("Validation errors:", err)
	}

	return config
}
