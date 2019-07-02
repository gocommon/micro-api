package api

import (
	"net/url"
	"strings"

	api "github.com/micro/micro/api/proto"
)

// Reqeust Reqeust
type Reqeust struct {
	req *api.Request
	url *url.URL
}

// NewRequest NewRequest
func NewRequest(req *api.Request) *Reqeust {
	return &Reqeust{
		req: req,
	}
}

// Header Header
func (p *Reqeust) Header(key string) string {
	if h, has := p.req.Header[key]; has {
		if len(h.Values) > 0 {
			return h.Values[0]
		}
	}

	return ""
}

// Query Query
func (p *Reqeust) Query(key string) string {
	if h, has := p.req.Get[key]; has {
		if len(h.Values) > 0 {
			return h.Values[0]
		}
	}

	return ""
}

// Post Post
func (p *Reqeust) Post(key string) string {
	if h, has := p.req.Post[key]; has {
		if len(h.Values) > 0 {
			return h.Values[0]
		}
	}

	return ""
}

// Form Form
func (p *Reqeust) Form(key string) string {
	if v := p.Post(key); len(v) == 0 {
		return v
	}

	return p.Query(key)
}

// Method Method
func (p *Reqeust) Method() string {
	return p.req.Method
}

// Path Path
func (p *Reqeust) Path() string {
	return p.req.Path
}

// URL URL
func (p *Reqeust) URL() *url.URL {
	if p.url == nil {
		p.url, _ = url.Parse(p.req.Url)
	}
	return p.url
}

// Body Body
func (p *Reqeust) Body() string {
	return p.req.Body
}

// Host Host
func (p *Reqeust) Host() string {
	return p.Header("Host")
}

// ContentType ContentType
func (p *Reqeust) ContentType() string {
	return p.Header("Content-Type")
}

// RemoteAddr 可能是逗号分隔的多个ip，有多个ip只取一个
func (p *Reqeust) RemoteAddr() string {
	ips := p.Header("X-Forwarded-For")
	if idx := strings.LastIndex(ips, ","); idx > -1 {
		// 只取最后一个
		return string(ips[idx+1:])
	}

	return ips
}

// Userinfo Userinfo

// User User
func (p *Reqeust) User() *url.Userinfo {
	if u := p.URL(); u != nil {
		return u.User
	}
	return nil
}

// RealIP RealIP
func (p *Reqeust) RealIP() string {

	ip := ""

	if ipPair := p.req.Header["X-Forwarded-For"]; ipPair != nil && len(ipPair.Values) > 0 {
		ipStr := ipPair.Values[0]

		ips := strings.Split(ipStr, ",")
		if len(ips) > 0 {
			ip = ips[0]
		}
	} else if ipPair := p.req.Header["X-Real-IP"]; ipPair != nil && len(ipPair.Values) > 0 {
		ip = ipPair.Values[0]
	}

	return ip
}
