package baostock

import (
	"fmt"
	"time"
)

// LoginResult 登录返回结果。
type LoginResult struct {
	ErrorCode string
	ErrorMsg  string
	UserID    string
}

func (r *LoginResult) Success() bool { return r.ErrorCode == ErrSuccess }

// Login 以匿名方式登录（默认 user_id=anonymous, password=123456）。
// 登录成功后保持 TCP 长连接，供后续查询使用。
func (c *Client) Login(userID, password string) (*LoginResult, error) {
	if userID == "" {
		userID = "anonymous"
	}
	if password == "" {
		password = "123456"
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	c.userID = userID

	// 如果设置了 apiKey，把它作为 options 字段传给服务端
	options := "0"
	if c.apiKey != "" {
		options = c.apiKey
	}
	msgBody := "login" + MessageSplit + userID + MessageSplit + password + MessageSplit + options
	resp, err := c.sendRecv(MsgTypeLoginReq, msgBody)
	if err != nil {
		return nil, err
	}

	arr, err := parseBodyArr(resp)
	if err != nil {
		return nil, err
	}
	if len(arr) < 2 {
		return nil, fmt.Errorf("baostock: login response fields < 2")
	}

	result := &LoginResult{
		ErrorCode: arr[0],
		ErrorMsg:  arr[1],
	}
	if result.Success() && len(arr) >= 4 {
		result.UserID = arr[3]
		c.userID = arr[3]
		fmt.Println("login success!")
	} else {
		fmt.Println("login failed:", result.ErrorMsg)
	}
	return result, nil
}

// Logout 登出并关闭连接。
type LogoutResult struct {
	ErrorCode string
	ErrorMsg  string
}

func (r *LogoutResult) Success() bool { return r.ErrorCode == ErrSuccess }

func (c *Client) Logout() (*LogoutResult, error) {
	if c.conn == nil {
		return &LogoutResult{ErrorCode: ErrNoLogin, ErrorMsg: "not connected"}, nil
	}

	now := time.Now().Format("20060102150405")
	msgBody := "logout" + MessageSplit + c.userID + MessageSplit + now

	resp, err := c.sendRecv(MsgTypeLogoutReq, msgBody)
	if err != nil {
		c.Close()
		return nil, err
	}

	arr, err := parseBodyArr(resp)
	if err != nil {
		c.Close()
		return nil, err
	}

	result := &LogoutResult{}
	if len(arr) >= 2 {
		result.ErrorCode = arr[0]
		result.ErrorMsg = arr[1]
	}
	if result.Success() {
		fmt.Println("logout success!")
	}

	c.Close()
	return result, nil
}
