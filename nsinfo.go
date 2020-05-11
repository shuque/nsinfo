package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/miekg/dns"
)

func usage() {
	fmt.Printf("Usage: nsinfo <zonename>\n")
}

func mkFqdn(qname string) string {
	if strings.HasSuffix(qname, ".") {
		return qname
	}
	return qname + "."
}

func doQuery(qname string, qtype uint16) (response *dns.Msg, err error) {

	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	// m.SetEdns0(4096, true)
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{Name: qname, Qtype: qtype, Qclass: dns.ClassINET}

	c := new(dns.Client)
	response, _, err = c.Exchange(m, "127.0.0.1:53")
	if err != nil {
		return nil, err
	}
	switch response.MsgHdr.Rcode {
	case dns.RcodeSuccess:
		break
	case dns.RcodeNameError:
		return nil, fmt.Errorf("NXDOMAIN: %s doesn't exist", qname)
	default:
		return nil, fmt.Errorf("Error: Response code: %s",
			dns.RcodeToString[response.MsgHdr.Rcode])
	}

	var rrcount int
	for _, rr := range response.Answer {
		if rr.Header().Rrtype == qtype {
			rrcount++
		}
	}
	if rrcount == 0 {
		return nil, fmt.Errorf("NODATA: %s/%d", qname, qtype)

	}

	return response, err
}

func getIPAddresses(hostname string) []net.IP {

	var ipList []net.IP

	var rrTypes = []uint16{
		dns.TypeAAAA,
		dns.TypeA,
	}

	for _, rrtype := range rrTypes {

		response, err := doQuery(hostname, rrtype)
		if err != nil {
			break
		}
		if response == nil {
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

func reverseLookup(ipaddr net.IP) string {

	arpaname, err := dns.ReverseAddr(ipaddr.String())
	if err != nil {
		return ""
	}
	response, err := doQuery(arpaname, dns.TypePTR)
	if err != nil || response == nil || len(response.Answer) < 1 {
		return "NO-PTR"
	}
	ptrRr := response.Answer[0].(*dns.PTR)
	return ptrRr.Ptr
}

const hexDigit = "0123456789abcdef"

var ip2asnSuffixV4 = "origin.asn.cymru.com."
var ip2asnSuffixV6 = "origin6.asn.cymru.com."

func ip2asn(ip net.IP) string {

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

	response, _ := doQuery(qname, dns.TypeTXT)
	if response == nil {
		fmt.Printf("No ASN found\n")
		return ""
	}

	txtRr := response.Answer[0].(*dns.TXT)
	return "AS[" + strings.TrimSuffix(strings.Split(txtRr.Txt[0], "|")[0], " ") + "]"

}

func main() {

	var err error

	if len(os.Args) != 2 {
		usage()
		return
	}
	zone := mkFqdn(os.Args[1])

	response, err := doQuery(zone, dns.TypeNS)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
	if response == nil {
		fmt.Printf("Error: null DNS response to NS query")
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

	// TODO: sort this in DNS canonical order
	sort.Strings(nsNames)

	var ipList []net.IP

	for _, nsName := range nsNames {
		ipList = getIPAddresses(nsName)
		for _, ip := range ipList {
			fmt.Printf("%s %s %s %s\n", nsName, ip, ip2asn(ip),
				reverseLookup(ip))
		}
	}

}
