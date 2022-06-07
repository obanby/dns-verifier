package verify_test

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/miekg/dns"
	"os"
	"testing"
	"verify"
)

func TestPrintUsage(t *testing.T) {
	t.Parallel()

	want := fmt.Sprintf(`usage: %s <config-file> <domain> <nameserver>`, os.Args[0])
	got := verify.Usage()

	if !cmp.Equal(want, got) {
		t.Errorf(cmp.Diff(want, got))
	}
}

func TestParsingConfigFile(t *testing.T) {
	t.Parallel()

	want := verify.DnsRecords{
		"aaaa":    {600, "AAAA", "2601:644:500:e210:62f8:1dff:feb8:947a"},
		"cname":   {300, "CNAME", "githubtest.net."},
		"foo.sub": {60, "A", "2.2.3.3"},
		"ptr":     {300, "PTR", "foo.bar.com."},
		"spf":     {600, "TXT", "v=spf1 ip4:192.168.0.1/16-all"},
		"txt":     {600, "TXT", "Twinkle twinkle little star"},
		"www":     {300, "A", "2.2.3.7"},
		"www.sub": {300, "A", "2.2.3.6"},
	}
	got, err := verify.ParseConfigFile("testdata/config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestGetRecordTypeFromString(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		arg  string
		want uint16
	}{
		{"AAAA", dns.TypeAAAA},
		{"A", dns.TypeA},
		{"CNAME", dns.TypeCNAME},
		{"TXT", dns.TypeTXT},
		{"PTR", dns.TypePTR},
		{"SPF", dns.TypeSPF},
		{"bogus", dns.TypeANY},
	}

	for _, tc := range testcases {
		got := verify.GetRecordTypeFromStr(tc.arg)

		if tc.want != got {
			t.Error(cmp.Diff(tc.want, got))
		}
	}
}

func TestMarshaling(t *testing.T) {
	t.Parallel()

	file, err := os.ReadFile("testdata/config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	want, err := verify.UnMarshal(file)
	if err != nil {
		t.Fatal(err)
	}

	data, err := verify.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := verify.UnMarshal(data)

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
