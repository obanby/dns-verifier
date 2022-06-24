package verify

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/miekg/dns"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"regexp"
	"strings"
	"verify/dnsclient"
)

type DnsRecords map[string]Record

type Record struct {
	TTL     int
	DnsType string `yaml:"type"`
	Value   string
}

func UnMarshal(data []byte) (DnsRecords, error) {
	records := DnsRecords{}
	err := yaml.Unmarshal(data, records)
	if err != nil {
		return DnsRecords{}, err
	}
	return records, nil
}

func Marshal(dr DnsRecords) ([]byte, error) {
	data, err := yaml.Marshal(dr)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func ParseConfigFile(path string) (DnsRecords, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return DnsRecords{}, err
	}

	return UnMarshal(file)
}

func GetRecordTypeFromStr(s string) uint16 {
	s = strings.ToUpper(s)
	switch s {
	case "TXT":
		return dns.TypeTXT
	case "A":
		return dns.TypeA
	case "AAAA":
		return dns.TypeAAAA
	case "CNAME":
		return dns.TypeCNAME
	case "PTR":
		return dns.TypePTR
	case "SPF":
		return dns.TypeSPF
	default:
		return dns.TypeANY
	}
}

var typeRegex = regexp.MustCompile(`IN\s+([a-zA-Z]+)`)
var valueRegex = regexp.MustCompile(`IN\s+[a-zA-Z]+\s+(.*)`)

func GetRecordFromResourceRecord(record dns.RR) (Record, error) {
	recordStr := record.String()
	recordStr = strings.ReplaceAll(recordStr, "\"", "")
	buff := bytes.NewBufferString(recordStr).Bytes()
	typeMatches := typeRegex.FindSubmatch(buff)
	valueMatches := valueRegex.FindSubmatch(buff)

	// 2 represent the minimum capture group count for regex grouping
	if len(typeMatches) < 2 || len(valueMatches) < 2 {
		return Record{}, errors.New("failed to parse the resource record")
	}

	r := Record{
		TTL:     int(record.Header().Ttl),
		DnsType: string(typeMatches[1]),
		Value:   string(valueMatches[1]),
	}

	return r, nil
}

type verifier struct {
	client   *dnsclient.Client
	LoggDest io.Writer
}

func NewVerifier(c *dnsclient.Client) *verifier {
	return &verifier{
		client:   c,
		LoggDest: os.Stdout,
	}
}

func (v *verifier) HasSubDomainChangeRecord(s string, record Record) (newRecord Record, hasChanged bool, err error) {
	v.client.SetSubDomain(s)
	answer, err := v.client.Query(
		GetRecordTypeFromStr(record.DnsType),
	)
	if err != nil {
		return newRecord, false, err
	}

	if len(answer) < 1 {
		err = fmt.Errorf("expected at least one record answer from nameserver %s for subdomain %s",
			v.client.NameServer,
			v.client.SubDomain,
		)
		return newRecord, false, err
	}

	resultRecord, err := GetRecordFromResourceRecord(answer[0])

	if !cmp.Equal(record, resultRecord) {
		fmt.Fprintf(v.LoggDest, "[CHG] resorce %s record did not match\n", v.client.FQDN())
		return resultRecord, true, err
	}

	fmt.Fprintf(v.LoggDest, "[ok] resorce %s record matched\n", v.client.FQDN())
	return record, false, nil
}

func Usage() string {
	return fmt.Sprintf(`usage: %s <config-file> <domain> <nameserver>`, os.Args[0])
}

func RunCLI() {
	if len(os.Args) < 4 {
		fmt.Println(Usage())
		os.Exit(1)
	}

	config, domain, nameserver := os.Args[1], os.Args[2], os.Args[3]

	client, err := dnsclient.New(
		dnsclient.WithDomain(domain),
		dnsclient.WithNameServer(nameserver),
	)

	if err != nil {
		panicWithMessage("internal error:", err)
	}

	verifer := NewVerifier(client)

	inRecords, err := ParseConfigFile(config)
	if err != nil {
		panicWithMessage("error parsing config file", err)
	}

	changedRecords := DnsRecords{}
	for subdomain, record := range inRecords {
		newRecord, hasChanged, err := verifer.HasSubDomainChangeRecord(subdomain, record)
		if err != nil {
			panicWithMessage("error while verifying subdomains:", err)
		}
		if hasChanged {
			changedRecords[subdomain] = newRecord
		}
	}

	if len(changedRecords) == 0 {
		return
	}

	data, err := Marshal(changedRecords)
	if err != nil {
		panicWithMessage("error marshaling changed records", err)
	}

	outRecord, err := os.OpenFile("changes.yaml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		panicWithMessage("error creating new changes.yaml file", err)
	}

	_, err = outRecord.Write(data)
	if err != nil {
		panicWithMessage("error writing to file", err)
	}
	outRecord.Close()
}

func panicWithMessage(message string, err error) {
	panic(fmt.Sprintf("%s: %q", message, err))
}
