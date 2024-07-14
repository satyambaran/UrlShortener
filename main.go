package main

import (
    "context"
    "crypto/sha256"
    "encoding/base64"
    "errors"
    "fmt"
    "log"
    "math/rand"
    "os"
    "strconv"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

const (
    initialLength = 6
    maxRetries    = 8
    host          = "localhost"
    port          = 5432
    user          = "postgres"
    password      = "password"
    dbname        = "db"
    baseURL       = "http://localhost:3000/"
    ttl           = 3 * 24 * 60 * 60
)

var ctx = context.Background()

type URL struct {
    ID          uint   `gorm:"primaryKey;autoIncrement"`
    ShortURL    string `gorm:"unique;not null"`
    OriginalURL string `gorm:"not null"`
}

type URLShortener struct {
    db    *gorm.DB
    cache *redis.Client
    rng   *rand.Rand
}

func NewURLShortener() (*URLShortener, error) {
    dbUrl := "host=" + host + " port=" + strconv.Itoa(port) + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"
    redisUrl := os.Getenv("REDIS_URL")
    db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    db.AutoMigrate(&URL{})

    cache := redis.NewClient(&redis.Options{
        Addr: redisUrl,
    })
    SetEvictionPolicy(cache)

    rng := rand.New(rand.NewSource(time.Now().UnixNano()))

    return &URLShortener{db: db, cache: cache, rng: rng}, nil
}
func SetEvictionPolicy(cache *redis.Client) {
    policy := "allkeys-lfu"

    _, err := cache.ConfigSet(ctx, "maxmemory-policy", policy).Result()
    if err != nil {
        log.Fatal("Failed to set eviction policy:", err)
    }
    currentPolicy, err := cache.ConfigGet(ctx, "maxmemory-policy").Result()
    if err != nil {
        log.Fatal("Failed to get current eviction policy:", err)
    }
    fmt.Println("Current eviction policy:", currentPolicy)
}
func (s *URLShortener) generateShortURL(url string, length int) string {
    // hash := sha256.Sum256([]byte(url + time.Now().String()))
    hash := sha256.Sum256([]byte(url + string(rune(s.rng.Int63()))))
    return base64.URLEncoding.EncodeToString(hash[:length])
}

func (s *URLShortener) Shorten(url, shortURL string) (string, error) {
    if shortURL != "" {
        newURL := URL{ShortURL: shortURL, OriginalURL: url}
        result := s.db.Create(&newURL)
        if result.Error != nil && !errors.Is(result.Error, gorm.ErrDuplicatedKey) {
            return "", result.Error
        }
        if result.Error != nil && errors.Is(result.Error, gorm.ErrDuplicatedKey) {
            return "", errors.New("requested url is not available")
        }
        s.cache.Set(ctx, shortURL, url, ttl)
        return baseURL + shortURL, nil
    }
    length := initialLength
    for i := 0; i <= maxRetries; i++ { //imp <=
        shortURL = s.generateShortURL(url, length)
        newURL := URL{ShortURL: shortURL, OriginalURL: url}
        result := s.db.Create(&newURL)
        if result.Error == nil {
            s.cache.Set(ctx, shortURL, url, ttl)
            return baseURL + shortURL, nil
        }
        if result.Error != nil && !errors.Is(result.Error, gorm.ErrDuplicatedKey) {
            return "", result.Error
        }
        if i == maxRetries-1 {
            length++
        }
    }
    return "", errors.New("failed to generate a unique short URL after multiple attempts")
}
func (s *URLShortener) Resolve(shortURL string) (string, error) {
    originalURL, err := s.cache.Get(ctx, shortURL).Result()
    if err == nil {
        return originalURL, nil
    }
    // Fallback to database if not found in cache
    var url URL
    result := s.db.First(&url, "short_url = ?", shortURL)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return "", errors.New("URL not found")
    } else if result.Error != nil {
        return "", result.Error
    }
    // Update cache
    s.cache.Set(ctx, shortURL, url.OriginalURL, ttl)
    return url.OriginalURL, nil
}

func main() {
    urlShortener, err := NewURLShortener()
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    app := fiber.New()
    app.Use(logger.New())

    app.Post("/shorten", func(c *fiber.Ctx) error {
        type request struct {
            URL          string `json:"url"`
            RequestedURL string `json:"requested_url"`
        }
        var req request
        if err := c.BodyParser(&req); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
        }
        shortURL, err := urlShortener.Shorten(req.URL, req.RequestedURL)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
        }
        return c.JSON(fiber.Map{"short_url": shortURL})
    })

    app.Get("/:shortURL", func(c *fiber.Ctx) error {
        shortURL := c.Params("shortURL")
        originalURL, err := urlShortener.Resolve(shortURL)
        if err != nil {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
        }
        return c.Redirect(originalURL)
    })

    log.Fatal(app.Listen(":3000"))
}
