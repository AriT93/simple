# AI Joke Agent

A terminal-based joke assistant that fetches jokes from an external API based on natural language requests.

## Overview

This project is a simple terminal UI application built with Go that allows users to request jokes using natural language. The application parses the user's request, extracts parameters like joke category and type, and fetches appropriate jokes from the JokeAPI.

### Features

- Natural language joke requests
- Support for different joke categories (programming, misc, dark, pun, etc.)
- Support for joke types (single, twopart)
- Content filtering with blacklist flags
- Terminal UI with spinner for loading states
- Comprehensive test suite using Ginkgo and Gomega

## Project Structure

```
.
├── ai-agent.go           # Main application entry point
├── jokeclient/           # Joke API client package
│   ├── client.go         # Client implementation
│   └── client_test.go    # Tests for client
├── model/                # Data models
│   ├── joke.go           # Joke response model
│   └── joke_test.go      # Tests for model
├── utils/                # Utility functions
│   ├── text.go           # Text processing utilities
│   └── text_test.go      # Tests for utilities
└── integration/          # Integration tests
    └── joke_api_test.go  # API integration tests
```

## Getting Started

### Prerequisites

- Go 1.18 or higher
- Git

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/ai-agent.git
   cd ai-agent
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the application:
   ```bash
   go build
   ```

4. Run the application:
   ```bash
   ./ai-agent
   ```

## Usage

Once the application is running, you can:

- Type natural language requests like "Tell me a programming joke"
- Specify joke types: "Tell me a twopart joke"
- Request specific categories: "Tell me a Christmas joke"
- Filter content: "Tell me a joke but nothing nsfw or political"
- Type "help" to see usage instructions
- Type "quit" or press ESC to exit

## Development

### Running Tests

Run all tests:
```bash
ginkgo -r
```

Run specific test suite:
```bash
ginkgo -v jokeclient
ginkgo -v utils
ginkgo -v model
```

Run integration tests:
```bash
ginkgo -v integration
```

### Adding New Features

1. Create a new branch for your feature
2. Implement the feature with appropriate tests
3. Run tests to ensure everything passes
4. Submit a pull request

## Contributing

Contributions are welcome! Here's how you can help:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure to update tests as appropriate and follow the existing code style.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
