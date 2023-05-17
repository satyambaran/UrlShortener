# URL Shortener

URL Shortener is a simple web application implemented in Go using the Gin web framework. It allows you to shorten long URLs into shorter and more manageable ones. The shortened URLs can then be used to redirect users to the original long URLs.

## Features

- URL shortening: Convert long URLs into short URLs.
- URL redirection: Redirect users from a short URL to the original long URL.
- Customizable short URL length: Adjust the length of the generated short URLs.
- In-memory database: Store the shortened URLs and their corresponding long URLs in an in-memory database.

## Prerequisites

- Go 1.16 or higher
- Git

## Getting Started

1. Clone the repository:

   ```shell
   git clone https://github.com/satyambaran/UrlShortener.git
   ```
2. Change to the project directory:
  
  ```shell
   cd UrlShortener
   ```
3. Build and run the application:
  ```shell
  go run main.go
  ```
4. Access the application at http://localhost:8080.

## Usage
Provide the long URL in the body of the API request. You can use tools like curl to interact with the URL shortener API. Here's an example of how to shorten a URL using curl:

  ```shell
  curl -X POST -d "url=http://example.com" http://localhost:8080/ 
  ```

This will return a JSON response containing the shortened URL:

  ```shell
  {"shortURL":"http://localhost:8080/AyWZrg"}
  ```
You can then use the shortened URL to redirect to the original long URL in browser or try getting redirected using curl:

  ```shell
  curl http://localhost:8080/AyWZrg  
  ```
  
