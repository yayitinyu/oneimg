package controllers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxRemoteImageRedirects = 5

var errRemoteImageTooLarge = errors.New("remote image exceeds size limit")

type remoteImage struct {
	Data        []byte
	ContentType string
	URL         *url.URL
}

func downloadRemoteImage(ctx context.Context, rawURL string, maxSize int64, allowedTypes []string) (*remoteImage, error) {
	if maxSize <= 0 {
		return nil, errors.New("invalid image size limit")
	}

	parsedURL, err := validateRemoteURL(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		Proxy:                 nil,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		DialContext: func(dialCtx context.Context, network, address string) (net.Conn, error) {
			return dialPublicAddress(dialCtx, dialer, network, address)
		},
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRemoteImageRedirects {
				return errors.New("too many redirects")
			}
			_, err := validateRemoteURL(req.Context(), req.URL.String())
			return err
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create remote image request: %w", err)
	}
	req.Header.Set("User-Agent", "oneimg/1.0")
	req.Header.Set("Accept", "image/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download remote image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote server returned status %d", resp.StatusCode)
	}
	if resp.ContentLength > maxSize {
		return nil, errRemoteImageTooLarge
	}

	contentType := normalizeContentType(resp.Header.Get("Content-Type"))
	if !isAllowedContentType(contentType, allowedTypes) {
		return nil, fmt.Errorf("unsupported remote content type %q", contentType)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("read remote image: %w", err)
	}
	if int64(len(data)) > maxSize {
		return nil, errRemoteImageTooLarge
	}
	if len(data) == 0 {
		return nil, errors.New("remote image is empty")
	}

	return &remoteImage{Data: data, ContentType: contentType, URL: resp.Request.URL}, nil
}

func validateRemoteURL(ctx context.Context, rawURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, errors.New("only http and https URLs are supported")
	}
	if parsedURL.Hostname() == "" || parsedURL.User != nil {
		return nil, errors.New("URL host is invalid")
	}
	if _, err := resolvePublicIPs(ctx, parsedURL.Hostname()); err != nil {
		return nil, err
	}
	return parsedURL, nil
}

func resolvePublicIPs(ctx context.Context, host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return nil, errors.New("private or local network addresses are not allowed")
		}
		return []net.IP{ip}, nil
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve remote host: %w", err)
	}
	if len(ips) == 0 {
		return nil, errors.New("remote host has no IP address")
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return nil, errors.New("remote host resolves to a private or local address")
		}
	}
	return ips, nil
}

func isPublicIP(ip net.IP) bool {
	return ip != nil && ip.IsGlobalUnicast() && !ip.IsPrivate() &&
		!ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast()
}

func dialPublicAddress(ctx context.Context, dialer *net.Dialer, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid remote address: %w", err)
	}
	ips, err := resolvePublicIPs(ctx, host)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, ip := range ips {
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	return nil, fmt.Errorf("connect to remote host: %w", lastErr)
}

func normalizeContentType(contentType string) string {
	if index := strings.IndexByte(contentType, ';'); index >= 0 {
		contentType = contentType[:index]
	}
	return strings.ToLower(strings.TrimSpace(contentType))
}

func isAllowedContentType(contentType string, allowedTypes []string) bool {
	for _, allowedType := range allowedTypes {
		if normalizeContentType(allowedType) == contentType {
			return true
		}
	}
	return false
}
