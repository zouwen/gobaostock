package baostock

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"strings"
	"time"
)

// Client 是 baostock 客户端，持有 TCP 连接和登录会话信息。
type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	userID string
	apiKey string // 可选：通过 SetAPIKey 设置
}

// New 创建一个新的客户端实例（未连接）。
func New() *Client {
	return &Client{}
}

// SetAPIKey 在登录前设置 API Key（付费用户使用，免费用户无需调用）。
// 对应 Python 的 bs.set_API_key(apiKey)。
func (c *Client) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

// connect 建立 TCP 连接。
func (c *Client) connect() error {
	addr := fmt.Sprintf("%s:%d", ServerHost, ServerPort)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("baostock: connect failed: %w", err)
	}
	c.conn = conn
	c.reader = bufio.NewReaderSize(conn, 1<<20) // 1MB 缓冲
	return nil
}

// Close 关闭底层 TCP 连接。
func (c *Client) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

// ---------------------------------------------------------------------------
// 消息头构造
// ---------------------------------------------------------------------------

// buildHeader 构造 21 字节消息头：version(7) + \x01 + msgType(2) + \x01 + bodyLen(10)
func buildHeader(msgType string, bodyLen int) string {
	lenStr := fmt.Sprintf("%010d", bodyLen)
	return ClientVersion + MessageSplit + msgType + MessageSplit + lenStr
}

// buildCRC32 计算 headBody 的 CRC32（IEEE 多项式），与 Python zlib.crc32 一致。
func buildCRC32(headBody string) uint32 {
	// Python zlib.crc32 使用 IEEE CRC-32，但返回有符号 int，Go 用 uint32 等效
	return crc32.ChecksumIEEE([]byte(headBody))
}

// sendRecv 发送一条消息并读取完整响应。
// 请求格式：header + body + \x01 + crc32 + \n
// 响应结束标志：
//   - 压缩消息（K 线 Plus）：<![CDATA[]]>\n
//   - 非压缩消息（其他接口）：以 \n 结尾（通过读取一行）
func (c *Client) sendRecv(msgType, msgBody string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("baostock: not connected")
	}

	header := buildHeader(msgType, len(msgBody))
	headBody := header + msgBody
	crc := buildCRC32(headBody)
	raw := headBody + MessageSplit + fmt.Sprintf("%d", crc) + MessageDelim

	_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := fmt.Fprint(c.conn, raw); err != nil {
		return "", fmt.Errorf("baostock: send failed: %w", err)
	}

	_ = c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// K 线 Plus 响应是压缩格式，结束标志为 <![CDATA[]]>\n
	if msgType == MsgTypeKDataPlusReq {
		endMarker := []byte(EndMarker)
		var buf []byte
		tmp := make([]byte, 65536)
		for {
			n, err := c.conn.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
				if bytes.HasSuffix(buf, endMarker) {
					break
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return "", fmt.Errorf("baostock: recv failed: %w", err)
			}
		}
		return c.parseCompressedResponse(buf)
	}

	// 其他接口：非压缩，读到换行符为止
	line, err := c.reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", fmt.Errorf("baostock: recv failed: %w", err)
	}
	return strings.TrimRight(line, "\n"), nil
}

// parseCompressedResponse 解析压缩响应（K 线 Plus）：21 字节头 + zlib 压缩体
func (c *Client) parseCompressedResponse(raw []byte) (string, error) {
	if len(raw) < HeaderLength {
		return "", fmt.Errorf("baostock: response too short (%d bytes)", len(raw))
	}

	headBytes := raw[:HeaderLength]
	headStr := string(headBytes)
	headParts := strings.SplitN(headStr, MessageSplit, 3)
	if len(headParts) < 3 {
		return "", fmt.Errorf("baostock: bad header: %q", headStr)
	}

	innerLenStr := strings.TrimSpace(headParts[2])
	innerLen := 0
	fmt.Sscanf(innerLenStr, "%d", &innerLen)

	compressedBody := raw[HeaderLength : HeaderLength+innerLen]
	r, err := zlib.NewReader(bytes.NewReader(compressedBody))
	if err != nil {
		return "", fmt.Errorf("baostock: zlib open failed: %w", err)
	}
	defer r.Close()
	decompressed, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("baostock: zlib decompress failed: %w", err)
	}
	return headStr + string(decompressed), nil
}

// ---------------------------------------------------------------------------
// 响应解析
// ---------------------------------------------------------------------------

// parseBodyArr 将 header(21字节) + body 的字符串切分成 body 的字段数组
func parseBodyArr(response string) ([]string, error) {
	if len(response) < HeaderLength {
		return nil, fmt.Errorf("baostock: response too short")
	}
	body := response[HeaderLength:]
	// 去掉末尾可能的 \n
	body = strings.TrimRight(body, "\n")
	return strings.Split(body, MessageSplit), nil
}
