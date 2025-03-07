# Shout

A simple CLI application for posting messages to Bluesky social network via the AT Protocol.

## Description

"Shout" is a Go-based utility that allows you to post messages to your Bluesky account directly from the command line. It handles authentication with the AT Protocol, securely stores your credentials, and provides a straightforward interface for sharing content on Bluesky.

## Features

- Post messages to Bluesky from the command line
- Secure credential storage in a local config file
- Simple authentication flow for first-time users
- Easy-to-use command-line interface

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
3. Your credentials will be securely stored in `~/.config/shout/config.json` (or the equivalent path on Windows)

```
$ ./shout "My first post from the command line!"
No config file found. Let's set up your Bluesky credentials.
Enter your Bluesky username: yourusername.bsky.social
Enter your Bluesky password: 
Credentials saved to ~/.config/shout/config.json
Posting to Bluesky as @yourusername.bsky.social...
Post successful! Your message has been posted to Bluesky.
```

### Regular Usage

After the initial setup, simply provide your message as a command-line argument:

```
$ ./shout "Hello Bluesky! This post was sent using a command-line tool."
Posting to Bluesky as @yourusername.bsky.social...
Post successful! Your message has been posted to Bluesky.
```

### Examples

Post a simple message:
```
$ ./shout "Just testing my new command-line tool for Bluesky!"
```

Post a message with hashtags:
```
$ ./shout "Learning to build tools with Go and the AT Protocol #developers #atproto"
```

## Configuration

The application stores credentials in a JSON file located at:
- Linux/macOS: `~/.config/shout/config.json`
- Windows: `C:\Users\<username>\.config\shout\config.json`

If you need to update your credentials, simply delete this file and you'll be prompted to enter new credentials on the next run.

### Security Considerations

- Your password is stored in plain text in the config file. For better security, consider using an app-specific password if Bluesky supports them.
- Ensure the config directory has appropriate permissions to protect your credentials.

## Troubleshooting

Common issues:
- **Authentication Failed**: Double-check your username and password. Delete the config file to reset credentials.
- **Post Failed**: Ensure you have an internet connection and your Bluesky account is in good standing.
- **Config File Issues**: Ensure the `.config/shout` directory exists and is writable.

## Development

To contribute to this project:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

This project is free, completely open for any purpose to use, modify, pass on, etc. 
