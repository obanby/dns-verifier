package dnsclient_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/miekg/dns"
	"testing"
	"verify/dnsclient"
)

func TestNewClient(t *testing.T) {
	t.Parallel()
	
	want := &dnsclient.Client{
		Domain:     "dns-exercise.dev",
		NameServer: "205.251.192.179:53",
		Port:       53,
	}

	got, err := dnsclient.New(
		dnsclient.WithPort(53),
		dnsclient.WithDomain("dns-exercise.dev"),
		dnsclient.WithNameServer("ns-179.awsdns-22.com"),
	)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestQuery(t *testing.T) {
	t.Parallel()

	client, err := dnsclient.New(
		dnsclient.WithDomain("spf.dns-exercise.dev"),
		dnsclient.WithNameServer("ns-179.awsdns-22.com"),
	)
	if err != nil {
		t.Fatal(err)
	}

	want := `spf.dns-exercise.dev.	600	IN	TXT	"v=spf1 ip4:192.168.0.1/16 -all"`

	answer, err := client.Query(dns.TypeTXT)
	if err != nil {
		t.Fatal(err)
	}

	if len(answer) < 1 {
		t.Fatal("expected at least one answer, got none")
	}

	got := answer[0].String()
	if !cmp.Equal(want, got) {
		t.Errorf(cmp.Diff(want, got))
	}
}
