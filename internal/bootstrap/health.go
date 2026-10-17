package bootstrap

import (
	"context"
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
	Model   string
	Timeout time.Duration
	Client  *http.Client
}

// CheckGatewayHealth thực hiện kiểm tra kết nối tiền kiểm (pre-flight health check) tới LLM gateway trước khi bắt đầu
// các tác vụ sáng tác dài hạn.
// Kiểm tra GET <base_url>/models, fallback GET <base_url>/health.
// Phát hiện sớm lỗi sập tunnel (Cloudflare 502/503), HTML error page, hoặc sai API key.
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

	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	// 1. Thử GET <baseURL>/models
	reqURL := baseURL + "/models"
	err := doHealthRequest(ctx, client, reqURL, opts.APIKey)
	if err == nil {
		return nil
	}

	// 2. Fallback: GET <baseURL>/health
	healthURL := baseURL + "/health"
	healthErr := doHealthRequest(ctx, client, healthURL, opts.APIKey)
	if healthErr == nil {
		return nil
	}

	// Trả về thông báo lỗi rõ ràng, mang tính hành động
	return fmt.Errorf("Không thể kết nối tới LLM Gateway tại %s: %w. Vui lòng kiểm tra notebook Colab/Kaggle hoặc đường truyền mạng.", baseURL, err)
}

func doHealthRequest(ctx context.Context, client *http.Client, targetURL, apiKey string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))

	// Phát hiện Cloudflare Tunnel / Web Server HTML Error Pages
	if strings.Contains(contentType, "text/html") ||
		strings.HasPrefix(strings.TrimSpace(string(bodyBytes)), "<!DOCTYPE") ||
		strings.HasPrefix(strings.TrimSpace(string(bodyBytes)), "<html") {
		if resp.StatusCode == http.StatusBadGateway || resp.StatusCode == http.StatusServiceUnavailable {
			return fmt.Errorf("Cloudflare Tunnel báo lỗi HTTP %d (Bad Gateway / Service Unavailable) — máy chủ chưa sẵn sàng", resp.StatusCode)
		}
		return fmt.Errorf("nhận được trang web HTML thay vì JSON API (HTTP %d)", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("xác thực thất bại (HTTP %d): API key không hợp lệ hoặc thiếu quyền", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("máy chủ trả về mã lỗi HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	return nil
}
