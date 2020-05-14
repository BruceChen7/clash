package http

import (
	"bufio"
	"encoding/base64"
	"net"
	"net/http"
	"strings"
	"time"

	adapters "github.com/Dreamacro/clash/adapters/inbound"
	"github.com/Dreamacro/clash/common/cache"
	"github.com/Dreamacro/clash/component/auth"
	"github.com/Dreamacro/clash/log"
	authStore "github.com/Dreamacro/clash/proxy/auth"
	"github.com/Dreamacro/clash/tunnel"
)

type HttpListener struct {
	net.Listener
	address string
    // 是否关闭
	closed  bool
	cache   *cache.Cache
}

func NewHttpProxy(addr string) (*HttpListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
    // 创建一个监听器
	hl := &HttpListener{l, addr, false, cache.New(30 * time.Second)}

	go func() {
		log.Infoln("HTTP proxy listening at: %s", addr)

		for {
			c, err := hl.Accept()
			if err != nil {
				if hl.closed {
					break
				}
				continue
			}
            // 处理连接
			go handleConn(c, hl.cache)
		}
	}()

	return hl, nil
}

func (l *HttpListener) Close() {
	l.closed = true
	l.Listener.Close()
}

func (l *HttpListener) Address() string {
	return l.address
}

func canActivate(loginStr string, authenticator auth.Authenticator, cache *cache.Cache) (ret bool) {
	if result := cache.Get(loginStr); result != nil {
		ret = result.(bool)
	}
    // base64接出来
	loginData, err := base64.StdEncoding.DecodeString(loginStr)
    // 采用basic鉴权
	login := strings.Split(string(loginData), ":")
	ret = err == nil && len(login) == 2 && authenticator.Verify(login[0], login[1])

    // 保存起来
	cache.Put(loginStr, ret, time.Minute)
	return
}

func handleConn(conn net.Conn, cache *cache.Cache) {
	br := bufio.NewReader(conn)
	request, err := http.ReadRequest(br)
	if err != nil || request.URL.Host == "" {
        // 连接关闭
		conn.Close()
		return
	}

    // 生成一个校验器
	authenticator := authStore.Authenticator()
	if authenticator != nil {
		if authStrings := strings.Split(request.Header.Get("Proxy-Authorization"), " "); len(authStrings) != 2 {
			_, err = conn.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic\r\n\r\n"))
			conn.Close()
			return
		} else if !canActivate(authStrings[1], authenticator, cache) {
			conn.Write([]byte("HTTP/1.1 403 Forbidden\r\n\r\n"))
			log.Infoln("Auth failed from %s", conn.RemoteAddr().String())
			conn.Close()
			return
		}
	}

    // 是https的请求，代理https
	if request.Method == http.MethodConnect {
		_, err := conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		if err != nil {
			return
		}
		tunnel.Add(adapters.NewHTTPS(request, conn))
		return
	}

	tunnel.Add(adapters.NewHTTP(request, conn))
}
