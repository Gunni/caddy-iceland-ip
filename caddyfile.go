package caddy_icelandic_ip

import (
	"bufio"
	"context"
	"net/http"
	"net/netip"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

const (
	ipv4 = "https://rix.is/is-net.txt"
	ipv6 = "https://rix.is/is-net6.txt"
)

func init() {
	caddy.RegisterModule(IcelandicIPRange{})
}

// IcelandicIPRange provides a range of IP address prefixes (CIDRs) retrieved from RIX.
type IcelandicIPRange struct {
	// refresh Interval
	Interval caddy.Duration `json:"interval,omitempty"`
	// request Timeout
	Timeout caddy.Duration `json:"timeout,omitempty"`

	// Holds the parsed CIDR ranges from Ranges.
	ranges []netip.Prefix

	ctx  caddy.Context
	lock *sync.RWMutex
}

// CaddyModule returns the Caddy module information.
func (IcelandicIPRange) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.ip_sources.icelandic",
		New: func() caddy.Module { return new(IcelandicIPRange) },
	}
}

// getContext returns a cancelable context, with a timeout if configured.
func (s *IcelandicIPRange) getContext() (context.Context, context.CancelFunc) {
	if s.Timeout > 0 {
		return context.WithTimeout(s.ctx, time.Duration(s.Timeout))
	}
	return context.WithCancel(s.ctx)
}

func (s *IcelandicIPRange) fetch(api string) ([]netip.Prefix, error) {
	ctx, cancel := s.getContext()
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var prefixes []netip.Prefix
	for scanner.Scan() {
		prefix, err := caddyhttp.CIDRExpressionToPrefix(scanner.Text())
		if err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

func (s *IcelandicIPRange) getPrefixes() ([]netip.Prefix, error) {
	var fullPrefixes []netip.Prefix
	// fetch ipv4 list
	prefixes, err := s.fetch(ipv4)
	if err != nil {
		return nil, err
	}
	fullPrefixes = append(fullPrefixes, prefixes...)

	// fetch ipv6 list
	prefixes, err = s.fetch(ipv6)
	if err != nil {
		return nil, err
	}
	fullPrefixes = append(fullPrefixes, prefixes...)

	return fullPrefixes, nil
}

func (s *IcelandicIPRange) Provision(ctx caddy.Context) error {
	s.ctx = ctx
	s.lock = new(sync.RWMutex)

	// update in background
	go s.refreshLoop()
	return nil
}

func (s *IcelandicIPRange) refreshLoop() {
	if s.Interval == 0 {
		s.Interval = caddy.Duration(time.Hour)
	}

	ticker := time.NewTicker(time.Duration(s.Interval))
	// first time update
	s.lock.Lock()
	// it's nil anyway if there is an error
	s.ranges, _ = s.getPrefixes()
	s.lock.Unlock()
	for {
		select {
		case <-ticker.C:
			fullPrefixes, err := s.getPrefixes()
			if err != nil {
				break
			}

			s.lock.Lock()
			s.ranges = fullPrefixes
			s.lock.Unlock()
		case <-s.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (s *IcelandicIPRange) GetIPRanges(_ *http.Request) []netip.Prefix {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ranges
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
//
//	icelandic {
//	   interval val
//	   timeout val
//	}
func (m *IcelandicIPRange) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // Skip module name.

	// No same-line options are supported
	if d.NextArg() {
		return d.ArgErr()
	}

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "interval":
			if !d.NextArg() {
				return d.ArgErr()
			}
			val, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return err
			}
			m.Interval = caddy.Duration(val)
		case "timeout":
			if !d.NextArg() {
				return d.ArgErr()
			}
			val, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return err
			}
			m.Timeout = caddy.Duration(val)
		default:
			return d.ArgErr()
		}
	}

	return nil
}

// interface guards
var (
	_ caddy.Module            = (*IcelandicIPRange)(nil)
	_ caddy.Provisioner       = (*IcelandicIPRange)(nil)
	_ caddyfile.Unmarshaler   = (*IcelandicIPRange)(nil)
	_ caddyhttp.IPRangeSource = (*IcelandicIPRange)(nil)
)
