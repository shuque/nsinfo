package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Defaults
var (
	defaultTimeout = 3
	defaultRetries = 3
)

// Globals
var resolver net.IP
var qopts = QueryOptions{rdflag: true, payload: 1460,
	timeout: time.Second * time.Duration(defaultTimeout),
	retries: defaultRetries}

//
// Print usage string
//
func usage() {
	fmt.Printf("Usage: nsinfo <zonename>\n")
}

//
// Make domain name fully qualified
//
func mkFqdn(qname string) string {
	if strings.HasSuffix(qname, ".") {
		return qname
	}
	return qname + "."
}

//
// Obtain list of IPv4 and IPv6 addresses for hostname
//
func getIPAddresses(resolver net.IP, hostname string) []net.IP {

	var ipList []net.IP
	var q Query

	var rrTypes = []uint16{
		dns.TypeAAAA,
		dns.TypeA,
	}

	for _, rrtype := range rrTypes {
		q = getQuery(hostname, rrtype, dns.ClassINET)
		response, err := sendQuery(q, resolver, qopts)
		if err != nil {
			break
		}
		for _, rr := range response.Answer {
			if rr.Header().Rrtype == rrtype {
				if rrtype == dns.TypeAAAA {
					ipList = append(ipList, rr.(*dns.AAAA).AAAA)
				} else if rrtype == dns.TypeA {
					ipList = append(ipList, rr.(*dns.A).A)
				}
			}
		}
	}

	return ipList

}

//
// Return reverse DNS lookup of given IP address
//
func reverseLookup(resolver net.IP, ipaddr net.IP) string {

	arpaname, err := dns.ReverseAddr(ipaddr.String())
	if err != nil {
		return ""
	}
	q := getQuery(arpaname, dns.TypePTR, dns.ClassINET)
	response, err := sendQuery(q, resolver, qopts)
	if err != nil || len(response.Answer) < 1 {
		return "NO-PTR"
	}
	ptrRr := response.Answer[0].(*dns.PTR)
	return ptrRr.Ptr
}

const hexDigit = "0123456789abcdef"

var ip2asnSuffixV4 = "origin.asn.cymru.com."
var ip2asnSuffixV6 = "origin6.asn.cymru.com."

func ip2asn(resolver net.IP, ip net.IP) string {

	var h1, h2 byte
	var qname string

	if ip.To4() != nil {
		// IPv4 address
		qname = fmt.Sprintf("%d.%d.%d.%d.%s",
			int(ip[3]),
			int(ip[2]),
			int(ip[1]),
			int(ip[0]),
			ip2asnSuffixV4)
	} else {
		// IPv6 address
		buf6 := make([]byte, 0, len(ip)*4)
		for i := len(ip) - 1; i >= 0; i-- {
			h1 = ip[i] & 0xf
			h2 = ip[i] >> 4
			buf6 = append(buf6, hexDigit[h1])
			buf6 = append(buf6, '.')
			buf6 = append(buf6, hexDigit[h2])
			buf6 = append(buf6, '.')
		}
		qname = string(buf6) + ip2asnSuffixV6
	}

	q := getQuery(qname, dns.TypeTXT, dns.ClassINET)
	response, _ := sendQuery(q, resolver, qopts)
	if response == nil {
		fmt.Printf("No ASN found\n")
		return ""
	}

	txtRr := response.Answer[0].(*dns.TXT)
	return "AS[" + strings.TrimSuffix(strings.Split(txtRr.Txt[0], "|")[0], " ") + "]"

}

//
// main()
//
func main() {

	var err error

	if len(os.Args) != 2 {
		usage()
		return
	}

	zone := mkFqdn(os.Args[1])
	resolver, err = getResolver()

	if err != nil {
		fmt.Printf("Error obtaining resolver address: %s", err.Error())
	}

	q := getQuery(zone, dns.TypeNS, dns.ClassINET)
	response, err := sendQuery(q, resolver, qopts)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	var nsRr *dns.NS
	var nsNames []string

	for _, rr := range response.Answer {
		if rr.Header().Rrtype == dns.TypeNS {
			nsRr = rr.(*dns.NS)
			nsNames = append(nsNames, nsRr.Ns)
		}
	}

	// TODO: sort in DNS canonical order
	sort.Strings(nsNames)

	var ipList []net.IP

	for _, nsName := range nsNames {
		ipList = getIPAddresses(resolver, nsName)
		for _, ip := range ipList {
			fmt.Printf("%s %s %s %s\n", nsName, ip, ip2asn(resolver, ip),
				reverseLookup(resolver, ip))
		}
	}

}
