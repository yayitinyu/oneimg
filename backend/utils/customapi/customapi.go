package customapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Config 自定义API配置
type Config struct {
	ApiUrl   string
	ApiKey   string
	DeleteUrl string
}

// NewCustomApiUploader 创建上传器
func NewCustomApiUploader(url, key, deleteUrl string) *Config {
	// 去除末尾斜杠
	url = strings.TrimRight(url, "/")
	return &Config{
		ApiUrl:    url,
		ApiKey:    key,
		DeleteUrl: deleteUrl,
	}
}

// Upload 上传图片
func (c *Config) Upload(fileBytes []byte, filename string) (*UploadResponse, error) {
	if c.ApiUrl == "" {
		return nil, fmt.Errorf("API地址未配置")
	}

	uploadUrl := c.ApiUrl + "/api/upload"

	// 准备表单数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加图片文件 (注意：字段名根据NodeSeek文档为 "image")
	part, err := writer.CreateFormFile("image", filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("创建表单文件失败: %v", err)
	}
	_, err = io.Copy(part, bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("写入文件内容失败: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("关闭Writer失败: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", uploadUrl, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.ApiKey != "" {
		req.Header.Set("X-API-Key", c.ApiKey)
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

    // DEBUG: Log the response body
    fmt.Printf("Custom API Response Body: %s\n", string(respBody))

	// 解析响应
	var result UploadResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		// 尝试作为文本返回错误
		return nil, fmt.Errorf("解析响应失败: %v, body: %s", err, string(respBody))
	}

	// 检查HTTP状态码 (虽然API可能返回200但success=false，但也要防备非200)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d, Message: %s", resp.StatusCode, result.Message)
	}

	if !result.Success && result.Code != 200 { // 兼容 code=200 表示成功的情况
		return nil, fmt.Errorf("上传失败: %s", result.Message)
	}

	return &result, nil
}

// Delete 删除图片
// 假设API为: DELETE /api/image/{id_or_hash_or_url_segment}
// NodeSeek文档： DELETE /api/image/{image_id}
// 这里我们需要确定 image_id 是什么。通常是 hash 或 database id。
// 我们在上传时最好保存这个 ID。
func (c *Config) Delete(imageIdentifier string) error {
	if c.ApiUrl == "" {
		return fmt.Errorf("API地址未配置")
	}

	// 构造删除URL
    var deleteUrl string
    if c.DeleteUrl != "" {
        deleteUrl = strings.ReplaceAll(c.DeleteUrl, "{id}", imageIdentifier)
        // 支持 {hash} 别名
        deleteUrl = strings.ReplaceAll(deleteUrl, "{hash}", imageIdentifier)
    } else {
	    // 默认 fallback
	    deleteUrl = fmt.Sprintf("%s/api/image/%s", c.ApiUrl, imageIdentifier)
    }

	req, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	if c.ApiKey != "" {
		req.Header.Set("X-API-Key", c.ApiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取body看是否有错误信息
		body, _ := io.ReadAll(resp.Body)
        fmt.Printf("Delete Failed: URL=%s Code=%d Body=%s\n", deleteUrl, resp.StatusCode, string(body))
		return fmt.Errorf("删除失败 HTTP %d: %s", resp.StatusCode, string(body))
	}
    
    // DEBUG LOG
    fmt.Printf("Delete Request Success: URL=%s\n", deleteUrl)

	// 解析可选
	var result DeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		if !result.Success && result.Code != 200 {
			return fmt.Errorf("删除API返回失败: %s", result.Message)
		}
	}

	return nil
}
