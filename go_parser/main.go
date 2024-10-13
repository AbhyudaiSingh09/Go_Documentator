

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"gopkg.in/yaml.v2"
	"io"
	"strings"
)

type Config struct {
	GoFilePath  string `yaml:"go_file_path"`
	GoDirectory string `yaml:"go_directory"`
	APIKey      string // This will hold the API key from the environment
}

type InterfaceDetails struct {
	InterfaceName   string   `json:"interface_name"`
	Methods         []string `json:"methods"`
	Implementations []string `json:"implementations"`
}

func main() {
	// Get the API key from the environment
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable not set")
	}

	// Read the YAML configuration
	config, err := readConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Assign the API key to the config struct
	config.APIKey = apiKey

	// Parse the file to find all interfaces and their methods
	interfaces := findInterfaces(config.GoFilePath)

	// Walk the services directory to find implementations of these interfaces
	result := findImplementations(config.GoDirectory, interfaces)

	// Send the data via HTTP to an API
	sendData(config.APIKey, result)
}

// Function to find all interfaces in a given Go file
func findInterfaces(filePath string) map[string][]string {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Error parsing Go file: %v", err)
	}

	interfaceMethods := make(map[string][]string)

	// Traverse the AST to find interface declarations
	ast.Inspect(node, func(n ast.Node) bool {
		if iface, ok := n.(*ast.TypeSpec); ok {
			if interfaceType, ok := iface.Type.(*ast.InterfaceType); ok {
				var methods []string
				for _, method := range interfaceType.Methods.List {
					if len(method.Names) > 0 { // Make sure the method has a name
						methods = append(methods, method.Names[0].Name)
					}
				}
				interfaceMethods[iface.Name.Name] = methods
			}
		}
		return true
	})

	return interfaceMethods
}

// Function to find all types in a directory that implement the detected interfaces
func findImplementations(dirPath string, interfaceMethods map[string][]string) []InterfaceDetails {
	var results []InterfaceDetails

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process Go files
		if strings.HasSuffix(info.Name(), ".go") {
			fset := token.NewFileSet()

			node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				log.Printf("Error parsing Go file %s: %v", path, err)
				return nil
			}

			// Traverse the file to find type declarations and methods
			ast.Inspect(node, func(n ast.Node) bool {
				if typeSpec, ok := n.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						typeName := typeSpec.Name.Name
						methods := getMethodsForType(node, typeName)

						// Check if this type implements any interface
						for iface, ifaceMethods := range interfaceMethods {
							if implementsInterface(ifaceMethods, methods) {
								// Add the implementation to the result
								found := false
								for i, detail := range results {
									if detail.InterfaceName == iface {
										results[i].Implementations = append(results[i].Implementations, typeName)
										found = true
										break
									}
								}
								if !found {
									results = append(results, InterfaceDetails{
										InterfaceName:   iface,
										Methods:         ifaceMethods,
										Implementations: []string{typeName},
									})
								}
							}
						}
					}
				}
				return true
			})
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}

	return results
}

// Function to get methods for a specific type (e.g., a struct)
func getMethodsForType(file *ast.File, typeName string) []string {
	var methods []string

	// Traverse the file and collect methods for the given type
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			// Check if the method has a receiver
			if fn.Recv != nil {
				for _, field := range fn.Recv.List {
					// Get the type name of the receiver (pointer or non-pointer)
					if starExpr, ok := field.Type.(*ast.StarExpr); ok {
						if ident, ok := starExpr.X.(*ast.Ident); ok && ident.Name == typeName {
							methods = append(methods, fn.Name.Name)
						}
					} else if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == typeName {
						methods = append(methods, fn.Name.Name)
					}
				}
			}
		}
		return true
	})

	return methods
}

// Function to check if a type implements an interface
func implementsInterface(ifaceMethods, typeMethods []string) bool {
	methodSet := make(map[string]bool)
	for _, method := range typeMethods {
		methodSet[method] = true
	}

	for _, ifaceMethod := range ifaceMethods {
		if !methodSet[ifaceMethod] {
			return false
		}
	}
	return true
}
// Function to send the data via HTTP to OpenAI API
func sendData(apiKey string, results []InterfaceDetails) {
	// Convert the results to a user message
	userMessageContent := formatResultsForMessage(results)

	// Construct the JSON payload for the API
	payload := map[string]interface{}{
		"model": "gpt-4", // You can adjust the model if needed
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": userMessageContent,
			},
		},
	}

	// Convert the payload to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshaling results: %v", err)
	}

	// Define the API endpoint to which you will send the data
	url := "https://api.openai.com/v1/chat/completions"

	// Prepare the HTTP request with the API key in the headers
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Data sent successfully!")
	} else {
		fmt.Printf("Failed to send data. Status code: %d\n", resp.StatusCode)
	}
}

// Helper function to format the results as a message for OpenAI API
func formatResultsForMessage(results []InterfaceDetails) string {
	message := "Here are the interfaces and their implementations:\n"
	for _, result := range results {
		message += fmt.Sprintf("Interface: %s\nMethods: %v\nImplementations: %v\n\n", result.InterfaceName, result.Methods, result.Implementations)
	}
	return message
}
// Function to read YAML config
func readConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}