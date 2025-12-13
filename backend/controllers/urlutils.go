package controllers

import "strings"

// formatNotificationURL 格式化通知中的URL
// 如果 url 已经是完整URL（带协议），直接返回
// 否则拼接 host 和 url
func formatNotificationURL(host, url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	// 确保 url 以 / 开头
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	return "https://" + host + url
}
