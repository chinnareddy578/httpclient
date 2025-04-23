# httpclient

`httpclient` is a lightweight and efficient HTTP client library written in Go. It simplifies making HTTP requests by providing an easy-to-use interface for GET, POST, PUT, DELETE, and other HTTP methods. This library is designed to handle common use cases like setting headers, query parameters, and handling JSON payloads.

## Installation

To use `httpclient` in your project, you can install it using `go get`:

```bash
go get github.com/chinnareddy578/httpclient
```

## Importing

After installation, you can import the library into your Go project:

```go
import "github.com/chinnareddy578/httpclient"
```

## Usage

Here is an example of how to use `httpclient`:

```go
package main

import (
	"fmt"
	"github.com/chinnareddy578/httpclient"
)

func main() {
	client := httpclient.New()

	// Example GET request
	response, err := client.Get("https://api.example.com/data", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Response:", string(response.Body))

	// Example POST request with JSON payload
	payload := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}
	response, err = client.Post("https://api.example.com/users", payload, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Response:", string(response.Body))
}
```

## Contribution Guide

We welcome contributions to improve `httpclient`. To contribute, please follow these steps:

1. Fork the repository on GitHub.
2. Create a new branch for your feature or bug fix:
   ```bash
   git checkout -b feature-name
   ```
3. Make your changes and commit them with clear and concise messages.
4. Push your changes to your forked repository:
   ```bash
   git push origin feature-name
   ```
5. Open a pull request on the main repository.

### Guidelines

- Ensure your code follows Go best practices.
- Write tests for any new features or bug fixes.
- Update the documentation if necessary.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

