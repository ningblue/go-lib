package hertz_http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type Client struct {
	c        *client.Client
	apiToken string
	timeout  int
	BaseURL  string
	ctx      context.Context
	Headers  map[string]string
}

func NewClient(baseURL string, apiToken string, timeout int) (*Client, error) {
	c, err := client.NewClient()
	if err != nil {
		hlog.Error("NewClient error", err)
		return nil, err
	}
	return &Client{
		c:        c,
		apiToken: apiToken,
		timeout:  timeout,
		BaseURL:  baseURL,
		ctx:      context.Background(),
		Headers:  make(map[string]string),
	}, nil
}

func (c *Client) SetHeaders(headers map[string]string) {
	for k, v := range headers {
		c.Headers[k] = v
	}
}

func (c *Client) SetHeader(key, value string) {
	c.Headers[key] = value
}

// Get  发送 GET 请求
func (c *Client) Get(path string, body []byte) ([]byte, error) {
	return c.doRequest(consts.MethodGet, path, body)
}

// Post 发送 POST 请求
func (c *Client) Post(path string, body []byte) ([]byte, error) {
	return c.doRequest(consts.MethodPost, path, body)
}

// 发送 HTTP 请求
func (c *Client) doRequest(method, path string, body []byte) ([]byte, error) {
	// 构建完整 URL
	url := c.BaseURL + path

	// 创建 Hertz 请求
	req := &protocol.Request{}
	res := &protocol.Response{}
	req.SetMethod(method)
	req.SetHeaders(c.Headers)
	req.SetRequestURI(url)
	// 设置请求体
	if body != nil {
		req.SetBody(body)
	}
	// 发送请求
	err := c.c.Do(c.ctx, req, res)

	// 处理错误
	if err != nil {
		hlog.Error("do error", err)
		return nil, err
	}
	if res.StatusCode() != 200 {
		hlog.Error("status code error: ", res.StatusCode(), " body: ", string(res.Body()))
		return nil, fmt.Errorf("status code error: %d, msg:%s ", res.StatusCode(), string(res.Body()))
	}
	if res.Body() == nil {
		hlog.Error("body is nil")
		return nil, fmt.Errorf("body is nil")
	}
	var tmp map[string]interface{}
	err = json.Unmarshal(res.Body(), &tmp)
	if err != nil {
		hlog.Error("json unmarshal error", err)
		return nil, err
	}
	var errMsg string
	if _, ok := tmp["msg"]; ok {
		if _, ok = tmp["msg"].(string); !ok {
			hlog.Error("msg is not string", tmp["msg"])
			return nil, fmt.Errorf("msg is not string")
		}
		errMsg = tmp["msg"].(string)
	}
	if _, ok := tmp["code"]; ok {
		if _, ok = tmp["code"].(float64); !ok {
			hlog.Error("code is not float64", tmp["code"])
			return nil, fmt.Errorf("code is not float64,msg :%s", tmp["code"])
		}
		if tmp["code"].(float64) != 0 && tmp["code"].(float64) != 200 {
			hlog.Error("code is not 0 or 200", tmp["code"])
			return nil, fmt.Errorf("code is not 0 or 200,msg :%s", errMsg)
		}
	}

	// 返回响应内容
	return res.Body(), err
}
