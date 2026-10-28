package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GatewayHealthOptions chứa thông tin cấu hình phục vụ tiền kiểm tra sức khỏe Gateway.
type GatewayHealthOptions struct {
	BaseURL string
	APIKey  string
	Headers map[string]string
	Model   string
	Timeout time.Duration
	Client  *http.Client
}

var errCloudflareAccessBlocked = errors.New("cloudflare zero trust access blocked")

// CheckGatewayHealth thực hiện kiểm tra kết nối tiền kiểm (pre-flight health check) tới LLM gateway trước khi bắt đầu
// các tác vụ sáng tác dài hạn.
// Kiểm tra GET <base_url>/models, fallback GET <base_url>/health.
// Phát hiện sớm lỗi sập tunnel (Cloudflare 502/503), Cloudflare Zero Trust Access 302 login redirect,
// HTML error page, hoặc sai API key.
func CheckGatewayHealth(ctx context.Context, opts GatewayHealthOptions) error {
	baseURL := strings.TrimRight(strings.TrimSpace(opts.BaseURL), "/")
	if baseURL == "" {
		return nil
	}
	// Bỏ qua các provider test/mock đặc biệt
	if baseURL == "mock://" || strings.HasPrefix(baseURL, "mock") || baseURL == "test://" {
		return nil
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	probeClient := buildHealthProbeClient(opts.Client, timeout)

	// 1. Thử GET <baseURL>/models
	reqURL := baseURL + "/models"
	err := doHealthRequest(ctx, probeClient, reqURL, opts.APIKey, opts.Headers)
	if err == nil {
		return nil
	}
	if errors.Is(err, errCloudflareAccessBlocked) {
		return fmt.Errorf("Không thể kết nối tới LLM Gateway tại %s: %w", baseURL, err)
	}

	// 2. Fallback: GET <baseURL>/health
	healthURL := baseURL + "/health"
	healthErr := doHealthRequest(ctx, probeClient, healthURL, opts.APIKey, opts.Headers)
	if healthErr == nil {
		return nil
	}

	// Trả về thông báo lỗi rõ ràng, mang tính hành động
	return fmt.Errorf("Không thể kết nối tới LLM Gateway tại %s: %w. Vui lòng kiểm tra notebook Colab/Kaggle hoặc đường truyền mạng.", baseURL, err)
}

func buildHealthProbeClient(base *http.Client, timeout time.Duration) *http.Client {
	var cloned http.Client
	if base != nil {
		cloned = *base
	} else {
		cloned.Timeout = timeout
	}
	origCheckRedirect := cloned.CheckRedirect
	cloned.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if req.Response != nil && isCloudflareAccessResponse(req.Response) {
			return http.ErrUseLastResponse
		}
		if req.URL != nil && isCloudflareAccessURL(req.URL.String()) {
			return http.ErrUseLastResponse
		}
		if origCheckRedirect != nil {
			return origCheckRedirect(req, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
	return &cloned
}

func isCloudflareAccessURL(rawURL string) bool {
	lower := strings.ToLower(strings.TrimSpace(rawURL))
	return strings.Contains(lower, "cloudflareaccess.com") ||
		strings.Contains(lower, "/cdn-cgi/access/login")
}

func isCloudflareAccessResponse(resp *http.Response) bool {
	if resp == nil {
		return false
	}
	if strings.Contains(strings.ToLower(resp.Header.Get("Www-Authenticate")), "cloudflare-access") {
		return true
	}
	if isCloudflareAccessURL(resp.Header.Get("Location")) {
		return true
	}
	if resp.Request != nil && resp.Request.URL != nil && isCloudflareAccessURL(resp.Request.URL.String()) {
		return true
	}
	return false
}

func doHealthRequest(ctx context.Context, client *http.Client, targetURL, apiKey string, headers map[string]string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		if strings.TrimSpace(k) != "" && strings.TrimSpace(v) != "" {
			req.Header.Set(k, v)
		}
	}
	if apiKey != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if isCloudflareAccessResponse(resp) {
		return fmt.Errorf("%w: miền đang bật Cloudflare Zero Trust Access (HTTP %d chuyển hướng sang trang đăng nhập Cloudflare Access). Hãy tắt Access Policy cho miền này trên Cloudflare Zero Trust Dashboard hoặc cấu hình Service Token (CF-Access-Client-Id / CF-Access-Client-Secret) trong providers.<tên>.extra.headers", errCloudflareAccessBlocked, resp.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	trimmedBody := strings.TrimSpace(string(bodyBytes))
	lowerPrefix := strings.ToLower(trimmedBody)

	// Phát hiện Cloudflare Tunnel / Web Server HTML Error Pages
	if strings.Contains(contentType, "text/html") ||
		strings.HasPrefix(lowerPrefix, "<!doctype") ||
		strings.HasPrefix(lowerPrefix, "<html") {
		if resp.StatusCode == http.StatusBadGateway || resp.StatusCode == http.StatusServiceUnavailable {
			return fmt.Errorf("Cloudflare Tunnel báo lỗi HTTP %d (Bad Gateway / Service Unavailable) — máy chủ chưa sẵn sàng", resp.StatusCode)
		}
		return fmt.Errorf("nhận được trang web HTML thay vì JSON API (HTTP %d)", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("xác thực thất bại (HTTP %d): API key không hợp lệ hoặc thiếu quyền", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("máy chủ trả về mã lỗi HTTP %d: %s", resp.StatusCode, trimmedBody)
	}

	return nil
}
