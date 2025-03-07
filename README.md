# Shout

A simple Go application that converts text to uppercase, making it look like you're SHOUTING!

## Description

"Shout" is a Go-based utility that takes input text and converts it to uppercase. It's a simple demonstration of string manipulation in Go, implemented as a module that can be used both as a standalone application and as a library in other Go projects.

## Features

- Convert any text to uppercase
- Simple command-line interface
- Can be used as a library in other Go applications

## Installation

### Prerequisites

- Go 1.16 or higher

### Building from source

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/shout.git
   cd shout
   ```

2. Build the application:
   ```
   go build
   ```

This will create an executable named `shout` in the project directory.

## Usage

### Command Line

Run the application directly:

```
./shout "hello world"
```

Output:
```
HELLO WORLD
```

### As a Library

You can import the package in your Go code:

```go
package main

import (
    "fmt"
    "shout"
)

func main() {
    input := "hello world"
    result := shout.ConvertToUppercase(input)
    fmt.Println(result) // Output: HELLO WORLD
}
```

## Development

To contribute to this project:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

