package ghproxy

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	loggerPkg "github.com/nerdneilsfield/shlogin/pkg/logger"
	"go.uber.org/zap"
)

const (
	sizeLimit int64 = 1024 * 1024 * 1024 * 999
	assetURL        = "https://hunshcn.github.io/gh-proxy"
	chunkSize int64 = 1024 * 10
)

var logger = loggerPkg.GetLogger()

//go:embed index.html
var indexHTMLFS embed.FS

//go:embed favicon.ico
var faviconFS embed.FS

var (
	whiteList = []string{}
	blackList = []string{}
	passList  = []string{}

	exp1 = regexp.MustCompile(`^(?:https?://)?github\.com/(?P<author>.+?)/(?P<repo>.+?)/(?:releases|archive)/.*$`)
	exp2 = regexp.MustCompile(`^(?:https?://)?github\.com/(?P<author>.+?)/(?P<repo>.+?)/(?:blob|raw)/.*$`)
	exp3 = regexp.MustCompile(`^(?:https?://)?github\.com/(?P<author>.+?)/(?P<repo>.+?)/(?:info|git-).*$`)
	exp4 = regexp.MustCompile(`^(?:https?://)?raw\.(?:githubusercontent|github)\.com/(?P<author>.+?)/(?P<repo>.+?)/.+?/.+$`)
	exp5 = regexp.MustCompile(`^(?:https?://)?gist\.(?:githubusercontent|github)\.com/(?P<author>.+?)/.+?/.+$`)
	exp6 = regexp.MustCompile(`^(?:https?://)?github\.com/(?P<author>.+?)\.keys$`)
)

func loggerClientInfo(c *fiber.Ctx) {
	remoteIP := c.IP()
	remoteAddr := c.Context().RemoteAddr().String()
	xForwardedFor := c.Get("X-Forwarded-For")
	requestURL := c.OriginalURL()

	logger.Debug("Client info", zap.String("remote_ip", remoteIP), zap.String("remote_addr", remoteAddr), zap.String("x_forwarded_for", xForwardedFor), zap.String("request_url", requestURL))
}

func checkURL(u string) []string {
	exps := []*regexp.Regexp{exp1, exp2, exp3, exp4, exp5, exp6}
	for _, exp := range exps {
		if matches := exp.FindStringSubmatch(u); matches != nil {
			return matches[1:] // Return author and repo
		}
	}
	return nil
}

func proxy(targetURL string, allowRedirects bool, c *fiber.Ctx) error {
	logger.Debug("Proxy to", zap.String("target_url", targetURL))
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !allowRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// Create new request
	proxyReq, err := http.NewRequest(c.Method(), targetURL, bytes.NewReader(c.Body()))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		if string(key) != "Host" {
			proxyReq.Header.Add(string(key), string(value))
		}
	})

	// Make request
	resp, err := client.Do(proxyReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer resp.Body.Close()

	// Check size limit
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		var size int64
		if _, err := fmt.Sscanf(contentLength, "%d", &size); err == nil && size > sizeLimit {
			return c.Redirect(targetURL)
		}
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Response().Header.Add(key, value)
		}
	}

	// Handle redirects
	if location := resp.Header.Get("Location"); location != "" {
		if matches := checkURL(location); matches != nil {
			c.Response().Header.Set("Location", "/"+location)
		} else {
			return proxy(location, true, c)
		}
	}

	c.Status(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return c.Send(body)
}

func Run(host string, port int, proxyJsDelivr bool) {
	app := fiber.New(fiber.Config{
		ProxyHeader: fiber.HeaderXForwardedFor,
	})

	// Index route
	app.Get("/", func(c *fiber.Ctx) error {
		loggerClientInfo(c)
		if q := c.Query("q"); q != "" {
			return c.Redirect("/" + q)
		}

		indexHTML, err := indexHTMLFS.ReadFile("index.html")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		// clone the header
		c.Response().Header.Set("Content-Type", "text/html; charset=utf-8")
		c.Response().Header.Set("Content-Length", fmt.Sprintf("%d", len(indexHTML)))
		c.Response().Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Response().Header.Set("Pragma", "no-cache")
		c.Response().Header.Set("Expires", "0")
		return c.Send(indexHTML)
	})

	// Favicon route
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		favicon, err := faviconFS.ReadFile("favicon.ico")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		c.Response().Header.Set("Content-Type", "image/x-icon")
		c.Response().Header.Set("Content-Length", fmt.Sprintf("%d", len(favicon)))
		return c.Send(favicon)
	})

	// Main handler
	app.All("/*", func(c *fiber.Ctx) error {
		loggerClientInfo(c)
		path := c.Params("*")
		path = strings.Replace(path, "/", "", 1)
		if !strings.HasPrefix(path, "http") {
			path = "https://" + path
		}

		if !strings.HasPrefix(path, "://") {
			path = strings.Replace(path, "s:/", "s://", 1)
		}

		matches := checkURL(path)
		if matches == nil {
			return c.Status(fiber.StatusForbidden).SendString("Invalid input.")
		}

		// Check white/black/pass lists here
		// Implementation omitted for brevity

		if (proxyJsDelivr || false) && exp2.MatchString(path) {
			newPath := strings.Replace(path, "/blob/", "@", 1)
			newPath = strings.Replace(newPath, "github.com", "cdn.jsdelivr.net/gh", 1)
			return c.Redirect(newPath)
		}

		if exp2.MatchString(path) {
			path = strings.Replace(path, "/blob/", "/raw/", 1)
		}

		return proxy(path, false, c)
	})

	addr := fmt.Sprintf("%s:%d", host, port)
	logger.Info("Server starting on", zap.String("addr", addr))
	err := app.Listen(addr)
	if err != nil {
		logger.Fatal("Server start failed: ", zap.Error(err))
	}
}
