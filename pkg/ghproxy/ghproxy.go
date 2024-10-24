package ghproxy

import (
	"embed"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	loggerPkg "github.com/nerdneilsfield/shlogin/pkg/logger"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const (
	sizeLimit int64 = 1024 * 1024 * 1024 * 999
	chunkSize int64 = 1024 * 10
)

var logger = loggerPkg.GetLogger()

//go:embed index.html
var indexHTMLFS embed.FS

//go:embed favicon.ico
var faviconFS embed.FS

var httpClient = &fasthttp.Client{
	MaxConnsPerHost:     1000,
	MaxIdleConnDuration: time.Minute * 5,
	ReadBufferSize:      64 * 1024, // 增加读取缓冲区大小到 64KB
	WriteBufferSize:     64 * 1024, // 同时增加写入缓冲区大小
}

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
	cfIP := c.Get("CF-Connecting-IP")
	trueClientIP := c.Get("True-Client-IP")

	if cfIP != "" {
		remoteIP = cfIP
	} else if trueClientIP != "" {
		remoteIP = trueClientIP
	}

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
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(targetURL)
	req.Header.SetMethod(c.Method())

	// 复制请求头
	c.Request().Header.VisitAll(func(key, value []byte) {
		if string(key) != "Host" {
			req.Header.Add(string(key), string(value))
		}
	})

	// 自定义重定向处理
	maxRedirects := 10
	for i := 0; i < maxRedirects; i++ {
		err := httpClient.Do(req, resp)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		// 检查内容长度
		if contentLength := resp.Header.Peek("Content-Length"); contentLength != nil {
			var size int64
			if _, err := fmt.Sscanf(string(contentLength), "%d", &size); err == nil && size > sizeLimit {
				logger.Warn("Content-Length exceeds limit", zap.Int64("size", size), zap.String("url", targetURL), zap.String("limit", fmt.Sprintf("%d", sizeLimit)))
				return c.Redirect(targetURL)
			}
		}

		// 检查是否为重定向状态码
		if resp.StatusCode() >= 300 && resp.StatusCode() < 400 && allowRedirects {
			location := resp.Header.Peek("Location")
			if location == nil {
				break
			}
			logger.Debug("Redirect to", zap.String("location", string(location)))
			req.SetRequestURI(string(location))
			resp.Reset()
		} else {
			break
		}
	}

	// 复制响应头和内容到客户端
	resp.Header.VisitAll(func(key, value []byte) {
		c.Response().Header.Add(string(key), string(value))
	})

	c.Status(resp.StatusCode())
	return c.Send(resp.Body())
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

	app.Get("/ip", func(c *fiber.Ctx) error {
		ip := c.IP()
		if cfIP := c.Get("CF-Connecting-IP"); cfIP != "" {
			ip = cfIP
		} else if trueIP := c.Get("True-Client-IP"); trueIP != "" {
			ip = trueIP
		} else if xFF := c.Get("X-Forwarded-For"); xFF != "" {
			ip = xFF
		}
		loggerClientInfo(c)
		logger.Debug("Client IP", zap.String("ip", ip))
		return c.SendString(ip)
	})

	app.Get("/ip/json", func(c *fiber.Ctx) error {
		loggerClientInfo(c)
		return c.JSON(fiber.Map{
			"ip":              c.IP(),
			"cf_ip":           c.Get("CF-Connecting-IP"),
			"true_ip":         c.Get("True-Client-IP"),
			"x_forwarded_for": c.Get("X-Forwarded-For"),
			"address":         c.Context().RemoteAddr().String(),
			"user_agent":      c.Get("User-Agent"),
			"referer":         c.Get("Referer"),
			"request_url":     c.OriginalURL(),
		})
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

		return proxy(path, true, c)
	})

	addr := fmt.Sprintf("%s:%d", host, port)
	logger.Info("Server starting on", zap.String("addr", addr))
	err := app.Listen(addr)
	if err != nil {
		logger.Fatal("Server start failed: ", zap.Error(err))
	}
}
