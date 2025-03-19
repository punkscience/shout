# Shout

A simple CLI application for posting messages to Bluesky social network via the AT Protocol.

## Description

"Shout" is a Go-based utility that allows you to post messages to your Bluesky account directly from the command line. It handles authentication with the AT Protocol, stores only the auth token for future use, and provides a straightforward interface for sharing content on Bluesky.

## Features

- Post messages to Bluesky from the command line
- Simple authentication flow for first-time users

## Installation

### Prerequisites

- Go 1.16 or higher
- A Bluesky account (register at [bsky.app](https://bsky.app) if you don't have one)


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

### First-time Setup

When you run the application for the first time, it will guide you through a one-time setup process:

1. You'll be prompted to enter your Bluesky username (typically your handle with or without the @)
2. You'll be asked to enter your password or app password (input will be hidden)
3. Your auth token will be stored in `~/.config/shout/config.json` (or the equivalent path on Windows)

```
$ ./shout auth bluesky
```

### Regular Usage

After the initial setup, simply provide your message as a command-line argument:

```
$ ./shout post "Hello Bluesky! This post was sent using a command-line tool."
```

## Configuration

The application stores the auth token in a JSON file located at:
- Linux/macOS: `~/.config/shout/config.json`
- Windows: `C:\Users\<username>\.config\shout\config.json`

If you need to update your credentials, simply delete this file and you'll be prompted to enter new credentials on the next run.

## Development

To contribute to this project:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

This project is free, completely open for any purpose to use, modify, pass on, etc. 
