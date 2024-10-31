# Epublib API
Epublib API is a Golang server for the Epublib app.

## Getting Started

### Prerequisites

- Go 1.16 or higher
- Git
- Docker (optional)

### Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/fumui/epublib-api.git
    cd epublib-api
    ```

2. Install dependencies:

    ```sh
    go mod tidy
    ```

3. Build the server:
    - Native build:
    ```sh
    go build -o epublib-api ./cmd/api/main.go
    ```
    - Docker build:
    ```sh
    docker compose build api
    ```
4. Configure the server using environment variables in sample.env file (and rename it to .env file for docker build/run).
5. Run the server:
    - Native run:
    ```sh
    ./epublib-api
    ```
    - Docker run:
    ```sh
    docker compose up api
    ```

### API Endpoints

The server provides swagger documentation at `api/v1/swagger/index.html`.