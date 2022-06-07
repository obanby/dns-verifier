package dnsclient

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"net"
)

type Client struct {
	Domain     string
	SubDomain  string
	NameServer string
	Port       int
}

type option func(*Client) error

func WithDomain(d string) option {
	return func(c *Client) error {
		if _, isValid := dns.IsDomainName(d); !isValid {
			return errors.New("can't set domain to an empty string")
		}
		c.Domain = d
		return nil
	}
}

func WithNameServer(ns string) option {
	return func(c *Client) error {
		ips, err := net.LookupIP(ns)
		if err != nil {
			return err
		}

		if len(ips) < 1 {
			return fmt.Errorf("expected at least 1 ip address while resolving %s nameserver ip", ns)
		}

		c.NameServer = fmt.Sprintf("%s:%d", ips[0].String(), c.Port)
		return nil
	}
}

func WithPort(p int) option {
	return func(c *Client) error {
		minRange := 1024
		maxRange := 65535
		if (p < 1024 || p > 65535) && p != 53 {
			return errors.New(
				fmt.Sprintf("cant set %d port number must be higher than %d, and smaller that %d.",
					p,
					minRange,
					maxRange),
			)
		}
		c.Port = p
		return nil
	}
}

func New(opts ...option) (*Client, error) {
	c := &Client{
		Port:      53,
		SubDomain: "",
	}
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return &Client{}, err
		}
	}
	return c, nil
}

func (c *Client) Query(dnsType uint16) ([]dns.RR, error) {
	domain := c.FQDN()
	dnsClient := dns.Client{}
	message := dns.Msg{}
	message.SetQuestion(domain, dnsType)
	response, _, err := dnsClient.Exchange(&message, c.NameServer)
	if err != nil {
		return []dns.RR{}, err
	}

	if len(response.Answer) == 0 {
		return []dns.RR{}, fmt.Errorf("couldn't find nameserver records in nameserver: %s", c.NameServer)
	}

	return response.Answer, nil
}

func (c *Client) SetSubDomain(s string) {
	if s == "" {
		return
	}
	c.SubDomain = s
}

func (c *Client) FQDN() string {
	domain := c.Domain + "."
	if c.SubDomain != "" {
		return fmt.Sprintf("%s.%s", c.SubDomain, domain)
	}
	return domain
}
