<!-- Write default READme -->

# README

## Description

This is a very simple example of what is a reverse proxy in Go.
Don't use this, it's just an example for learning purposes.

## Features

- Redirect all requests to the target server
- Add X-Forwarded-For header to keep track of the original IP address
- Cache the response of the target server by default for 1 minute

In the future, i may add more features like load balancing, security, etc.

## Usage

Define the demo of the target server in the `main.go` file in the demoUrl variable.

```go
var demoUrl = "http://localhost:8080"
```

Then run the server with the following command:

```bash
go run main.go
```




