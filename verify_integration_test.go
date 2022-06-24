//go:build integration
// +build integration

package verify_test

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/miekg/dns"
	"os"
	"testing"
	"verify"
	"verify/dnsclient"
)

func TestGetRecordFromResourceRecord(t *testing.T) {
	t.Parallel()

	want := verify.Record{600, "AAAA", "2601:644:500:e210:62f8:1dff:feb8:947a"}
	client, err := dnsclient.New(
		dnsclient.WithDomain("dns-exercise.dev"),
		dnsclient.WithNameServer("ns-179.awsdns-22.com"),
	)

	if err != nil {
		t.Fatal(err)
	}
	client.SetSubDomain("aaaa")
	answer, err := client.Query(dns.TypeAAAA)

	if err != nil {
		t.Fatal(err)
	}

	if len(answer) < 1 {
		t.Fatal("didn't get an answer for ns")
	}
	got, err := verify.GetRecordFromResourceRecord(answer[0])
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestDnsQueryForRecord(t *testing.T) {
	t.Parallel()

	records := verify.DnsRecords{
		"aaaa":    {600, "AAAA", "2601:644:500:e210:62f8:1dff:feb8:947a"},
		"cname":   {300, "CNAME", "githubtest.net."},
		"foo.sub": {60, "A", "2.2.3.3"},
		"ptr":     {300, "PTR", "foo.bar.com."},
		"spf":     {600, "TXT", `v=spf1 ip4:192.168.0.1/16 -all`},
		"www.sub": {300, "A", "2.2.3.6"},
	}

	client, err := dnsclient.New(
		dnsclient.WithDomain("dns-exercise.dev"),
		dnsclient.WithNameServer("ns-179.awsdns-22.com"),
	)

	if err != nil {
		t.Fatal(err)
	}

	for subDomain, record := range records {
		want := records[subDomain]
		recordType := verify.GetRecordTypeFromStr(record.DnsType)
		client.SetSubDomain(subDomain)
		answer, err := client.Query(recordType)
		if err != nil {
			t.Fatal(err)
		}

		if len(answer) < 1 {
			t.Fatal("expected at least one answer")
		}

		got, err := verify.GetRecordFromResourceRecord(answer[0])
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(want, got) {
			t.Error(cmp.Diff(want, got))
		}
	}
}

func TestHasChangedRecord(t *testing.T) {
	t.Parallel()

	records := verify.DnsRecords{
		"aaaa":    {600, "AAAA", "2601:644:500:e210:62f8:1dff:feb8:947a"},
		"cname":   {300, "CNAME", "githubtest.net."},
		"foo.sub": {60, "A", "2.2.3.3"},
		"ptr":     {300, "PTR", "foo.bar.com."},
		"spf":     {600, "TXT", `v=spf1 ip4:192.168.0.1/16 -all`},
		"www.sub": {300, "A", "2.2.3.6"},
		"www":     {300, "A", "2.2.3.7"},
		"txt":     {300, "TXT", "Twinkle twinkle little star"},
	}

	client, err := dnsclient.New(
		dnsclient.WithDomain("dns-exercise.dev"),
		dnsclient.WithNameServer("ns-179.awsdns-22.com"),
	)

	if err != nil {
		t.Fatal(err)
	}

	verifier := verify.NewVerifier(client)
	verifier.LoggDest = &bytes.Buffer{}

	want := verify.DnsRecords{
		"www": {300, "A", "2.2.3.6"},
		"txt": {600, "TXT", `Bah bah black sheep`},
	}

	got := verify.DnsRecords{}
	for subdomain, record := range records {
		newRecord, hasChanged, err := verifier.HasSubDomainChangeRecord(subdomain, record)
		if err != nil {
			t.Fatal(err)
		}
		if hasChanged {
			got[subdomain] = newRecord
		}
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestGenerateChangedRecordsFile(t *testing.T) {
	t.Parallel()

	file, err := os.ReadFile("testdata/config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	records, err := verify.UnMarshal(file)
	if err != nil {
		t.Fatal(err)
	}

	client, err := dnsclient.New(
		dnsclient.WithDomain("dns-exercise.dev"),
		dnsclient.WithNameServer("ns-179.awsdns-22.com"),
	)
	if err != nil {
		t.Fatal(err)
	}

	want := verify.DnsRecords{}
	verifier := verify.NewVerifier(client)
	verifier.LoggDest = &bytes.Buffer{}

	for subdomain, record := range records {
		newRecord, hasChanged, err := verifier.HasSubDomainChangeRecord(subdomain, record)
		if err != nil {
			t.Fatal(err)
		}
		if hasChanged {
			want[subdomain] = newRecord
		}
	}

	path := t.TempDir() + "/changed.yaml"
	outFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		t.Fatal(err)
	}

	data, err := verify.Marshal(want)
	_, err = outFile.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	outFile.Close()

	inFile, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got, err := verify.UnMarshal(inFile)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
