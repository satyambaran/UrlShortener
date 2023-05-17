package main

import (
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

var (
    database           map[string]string
    databaseReverseMap map[string]string // wont be needing this in real database
    currentLength      = 6
    maxRetriesAllowed  = 6
    letters            = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    baseURL            = "http://localhost:8080/"
    allowedURLs        = map[string][]string{
        "GET":  {"/:shortURL"},
        "POST": {"/"},
    }
)

func main() {
    createDB()
    rand.Seed(time.Now().UnixNano())

    router := gin.Default()
    // router.Use(middleWare)
    router.GET("/:shortURL", urlRedirect)
    router.POST("/", urlShortner)
    router.NoRoute(noRoute)
    if err := router.Run(":8080"); err != nil {
        log.Fatal(err)
    }
}
func noRoute(c *gin.Context) {
    c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
}

// func middleWare(c *gin.Context) {
//     requestedURL := c.Request.URL.Path
//     method := c.Request.Method
//     isAllowed := isURLAllowed(requestedURL, method)
//     if isAllowed {
//         c.Next()
//     } else {
//         c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
//         c.Abort()
//     }
// }
// func isURLAllowed(url, method string) bool {
//     allowedURLsForThisMethod, ok := allowedURLs[method]
//     if !ok {
//         return false
//     }
//     for _, allowedURL := range allowedURLsForThisMethod {
//         if allowedURL == url {
//             return true
//         }
//     }
//     return false
// }
func urlShortner(c *gin.Context) {
    url := c.PostForm("url")
    if url == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
        return
    }

    val, exist := checkIfExistInDBRM(url)
    if exist {
        c.JSON(http.StatusOK, gin.H{"shortURL": val})
    }

    shortURL := generateShortURL()
    addInDB(shortURL, url)

    responseURL := fmt.Sprintf("%s%s", baseURL, shortURL)
    c.JSON(http.StatusOK, gin.H{"shortURL": responseURL})
}

func urlRedirect(c *gin.Context) {
    shortURL := c.Param("shortURL")
    url, ok := database[shortURL]
    if !ok {
        c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
        return
    }
    c.Redirect(http.StatusFound, url)
}

func generateShortURL() string {
    for i := 0; i < maxRetriesAllowed; i++ {
        var sb strings.Builder
        for i := 0; i < currentLength; i++ {
            sb.WriteByte(letters[rand.Intn(len(letters))])
        }
        shortURL := sb.String()
        exist := checkIfExistInDB(shortURL)
        if !exist {
            return shortURL
        }
    }
    currentLength++
    var sb strings.Builder
    for i := 0; i < currentLength; i++ {
        sb.WriteByte(letters[rand.Intn(len(letters))])
    }
    shortURL := sb.String() //since we just increased the length, there is no chance to have a retry
    return shortURL
}

func createDB() {
    database = make(map[string]string)
    databaseReverseMap = make(map[string]string)
}
func checkIfExistInDB(shortURL string) bool {
    _, ok := database[shortURL]
    return ok
}
func checkIfExistInDBRM(URL string) (string, bool) {
    val, ok := databaseReverseMap[URL]
    return val, ok
}
func addInDB(shortURL, url string) {
    database[shortURL] = url
    databaseReverseMap[url] = shortURL
}

/*
curl -X POST -d "url=http://1example.com" http://localhost:8080/ &&
curl -X POST -d "url=http://e1xample.com" http://localhost:8080/ &&
curl -X POST -d "url=http://ex1ample.com" http://localhost:8080/ &&
curl -X POST -d "url=http://exa1mple.com" http://localhost:8080/ &&
curl -X POST -d "url=http://exam1ple.com" http://localhost:8080/ &&
curl -X POST -d "url=http://examp1le.com" http://localhost:8080/ &&
curl -X POST -d "url=http://exampl1e.com" http://localhost:8080/ &&
curl -X POST -d "url=http://example1.com" http://localhost:8080/ &&
curl -X POST -d "url=http://2example.com" http://localhost:8080/ &&
curl -X POST -d "url=http://e2xample.com" http://localhost:8080/ &&
curl -X POST -d "url=http://ex2ample.com" http://localhost:8080/ &&
curl -X POST -d "url=http://exa2mple.com" http://localhost:8080/ &&
curl -X POST -d "url=http://exam2ple.com" http://localhost:8080/ &&
curl -X POST -d "url=http://examp2le.com" http://localhost:8080/ &&
curl -X POST -d "url=http://exampl2e.com" http://localhost:8080/ &&
curl -X POST -d "url=http://example2.com" http://localhost:8080/ &&
curl -X POST -d "url=http://ex1ample.com" http://localhost:8080/ &&
curl -X POST -d "url=http://examp2le.com" http://localhost:8080/
*/
