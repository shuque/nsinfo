package main

import (
	"errors"
	"os"
	"fmt"
	"strings"
	"net"
	"sort"
	"github.com/miekg/dns"
)


func usage() {
	fmt.Printf("Usage: nsinfo <zonename>\n")
}


func mk_fqdn(qname string) (string) {
	if strings.HasSuffix(qname, ".") {
		return qname
	} else {
		return qname + "."
	}
}


func do_query(qname string, qtype uint16) (response *dns.Msg, err error) {

	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	// m.SetEdns0(4096, true)
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{qname, qtype, dns.ClassINET}

	c := new(dns.Client)
	response, _, err = c.Exchange(m, "127.0.0.1:53")
	if err != nil {
		return nil, err
	}
	switch response.MsgHdr.Rcode {
	case dns.RcodeSuccess:
		break
	case dns.RcodeNameError:
		return nil, errors.New(fmt.Sprintf("NXDOMAIN: %s: name doesn't exist\n", qname))
	default:
		return nil, errors.New(fmt.Sprintf("Error: Response code: %s\n", 
			dns.RcodeToString[response.MsgHdr.Rcode]))
	}

	var rrcount int
	for _, rr := range response.Answer {
		if rr.Header().Rrtype == qtype {
			rrcount += 1
		}
        }
	if rrcount == 0 {
		return nil, errors.New(fmt.Sprintf("NODATA: %s/%s\n", qname, qtype))

	}

	return response, err
}


func getIPAddresses(hostname string, rrtype uint16) ([]net.IP) {

	var ip_list []net.IP

	switch rrtype {

	case dns.TypeAAAA, dns.TypeA:
		response, err := do_query(hostname, rrtype)
		if err == nil && response != nil {
			for _, rr := range response.Answer {
				if rr.Header().Rrtype == rrtype {
					if rrtype == dns.TypeAAAA {
						ip_list = append(ip_list, rr.(*dns.AAAA).AAAA)
					} else if rrtype == dns.TypeA {
						ip_list = append(ip_list, rr.(*dns.A).A)
					}
				}
			}
		
		}
	default:
		fmt.Printf("getIPAddresses: %d: invalid rrtype\n", rrtype)
	}

	return ip_list

}


func reverseLookup(ipaddr net.IP) (string) {

	arpaname, err := dns.ReverseAddr(ipaddr.String())
	if err != nil {
		return ""
	}
	response, err := do_query(arpaname, dns.TypePTR)
	if response == nil {
		return "NO-PTR"
	}
	if len(response.Answer) < 1 {
		return "NO-PTR"
	}
	ptr_rr := response.Answer[0].(*dns.PTR)
	return ptr_rr.Ptr
}


const hexDigit = "0123456789abcdef"
var ip2asnSuffixV4 = "origin.asn.cymru.com."
var ip2asnSuffixV6 = "origin6.asn.cymru.com."

func ip2asn(ip net.IP) (string) {

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
		for i := len(ip)-1; i >= 0; i-- {
			h1 = ip[i] & 0xf
			h2 = ip[i] >> 4
			buf6 = append(buf6, hexDigit[h1])
			buf6 = append(buf6, '.')
			buf6 = append(buf6, hexDigit[h2])
			buf6 = append(buf6, '.')
		}
		qname = string(buf6) + ip2asnSuffixV6
	}

	response, _ := do_query(qname, dns.TypeTXT)
	if response == nil {
		fmt.Printf("No ASN found\n")
		return ""
	}

	txt_rr := response.Answer[0].(*dns.TXT)
	return "AS[" + strings.TrimSuffix(strings.Split(txt_rr.Txt[0], "|")[0], " ") + "]"

}


func main() {

	if len(os.Args) != 2 {
		usage()
		return
	}
	zone := mk_fqdn(os.Args[1])
	response, _ := do_query(zone, dns.TypeNS)
	if response == nil {
		return
	}

	var ns_rr *dns.NS
	var ns_name_list []string

	for _, rr := range response.Answer {
		if rr.Header().Rrtype == dns.TypeNS {
			ns_rr = rr.(*dns.NS)
			ns_name_list = append(ns_name_list, ns_rr.Ns)
		}
	}

	// should really be sorting in DNS canonical order here .. //
	sort.Strings(ns_name_list)

	var v4_list []net.IP
	var v6_list []net.IP

	for _, ns_name := range ns_name_list {
		v4_list = nil
		v6_list = nil

		v6_list = getIPAddresses(ns_name, dns.TypeAAAA)
		for _, ip := range v6_list {
			fmt.Printf("%s %s %s %s\n", ns_name, ip, ip2asn(ip), reverseLookup(ip))
		}


		v4_list = getIPAddresses(ns_name, dns.TypeA)
		for _, ip := range v4_list {
			fmt.Printf("%s %s %s %s\n", ns_name, ip, ip2asn(ip), reverseLookup(ip))
		}

	}

}
