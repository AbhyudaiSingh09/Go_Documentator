
# Go Interface Parser and Implementation Reporter

This Go project is designed to:
- Parse Go source files to extract interfaces and their methods.
- Walk through a directory to find all types that implement the detected interfaces.
- Send the results via an HTTP POST request to an API, using an API key for authentication.

## Features

- Parses Go files to detect all interfaces and their methods.
- Finds types that implement the detected interfaces within a specified directory.
- Sends the parsed interface and implementation data to the OpenAI API (or any other API) via a POST request.
- API authentication using a Bearer token (API key).

## Requirements

- Go 1.16 or higher
- OpenAI API key (or another API key if you change the destination API)
- YAML configuration file to specify the source file and directory

## Installation

1. Clone the repository or download the source code.
   
   ```bash
   git clone https://github.com/AbhyudaiSingh09/Go_Documentator.git
   cd go_parser

	2.	Install necessary Go packages.

go mod tidy



Configuration

You need to provide a configuration file (config.yaml) in the following format:

go_file_path: "path/to/your/sourcefile.go"
go_directory: "path/to/your/services_directory"

	•	go_file_path: The file that contains the interfaces you want to parse.
	•	go_directory: The directory where the code will search for implementations of these interfaces.

Example config.yaml

go_file_path: "services/access/access.go"
go_directory: "services"

Usage

	1.	Set your API key as an environment variable.
Export the API key (e.g., OpenAI API key) to your terminal session:

export API_KEY="your_openai_api_key"


	2.	Run the program.
Execute the program using the Go command:

go run main.go



How It Works

Parsing Interfaces

The program parses a Go file (specified in config.yaml) and detects all the interfaces and their associated methods. It collects this data into a map where each interface name maps to a slice of method names.

Finding Implementations

It walks through the specified directory, parses each Go file, and looks for types (e.g., structs) that implement the previously detected interfaces. For each type, it compares the methods to those defined by the interfaces to determine if it implements any of them.

Sending Data via API

The results (interfaces, methods, and implementations) are formatted into a message and sent via an HTTP POST request to an API endpoint (e.g., OpenAI API) using the Bearer token authorization method.

Example Output

Here are the interfaces and their implementations:
Interface: Service
Methods: [Actions ActionByID]
Implementations: [UserService OrderService]

API Integration

The sendData function sends the results to the OpenAI API (or any API you configure).

API Request Example

curl https://api.openai.com/v1/chat/completions \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $OPENAI_API_KEY" \
-d '{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "Write a haiku about AI."}
  ]
}'

Modifying the API Endpoint

You can change the sendData function to send the results to a different API by modifying the URL and request headers.

Error Handling

	•	If the API key is not set, the program will terminate with the error: API_KEY environment variable not set.
	•	If there is an error reading the Go files, or if the HTTP request fails, appropriate error messages will be logged.

Contributing

Feel free to submit pull requests or open issues if you find bugs or have suggestions for improvements.

License

This project is licensed under the MIT License.

---
