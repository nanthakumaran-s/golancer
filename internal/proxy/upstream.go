package proxy

import (
	"fmt"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL *url.URL
}

type UpstreamPool struct {
	Backends []*Backend
	next     uint64
}

func NewUpstreamPool(urls []string) (*UpstreamPool, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no upstream URLs found")
	}

	bs := make([]*Backend, 0, len(urls))

	for _, s := range urls {
		url, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the URL %s", s)
		}
		bs = append(bs, &Backend{URL: url})
	}

	return &UpstreamPool{Backends: bs}, nil
}

func (u *UpstreamPool) Next() *Backend {
	if len(u.Backends) == 0 {
		return nil
	}

	n := atomic.AddUint64(&u.next, 1)
	idx := int((n - 1) % uint64(len(u.Backends)))
	return u.Backends[idx]
}
