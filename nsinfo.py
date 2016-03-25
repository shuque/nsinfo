#!/usr/bin/env python
#

"""
nsinfo.py

For a given DNS zone, print out some information about its authoritative
nameservers, specifically nameserver name, IP address, ASN originating
prefix covering the IP address, and reverse mapping of IP address to name.

Author: Shumon Huque <shuque@gmail.com>

Example run:

$ nsinfo.py complex.com
ns1.complex.com. 192.16.16.5 AS15133 ns1.edgecastdns.net.
ns1.complex.com. 2606:2800:3::5 AS15133 ns1.edgecastdns.net.
ns2.complex.com. 192.16.16.6 AS15133 ns2.edgecastdns.net.
ns2.complex.com. 2606:2800:3::6 AS15133 ns2.edgecastdns.net.
ns3.complex.com. 198.7.29.5 AS15133 ns3.edgecastdns.net.
ns3.complex.com. 2606:2800:c::5 AS15133 ns3.edgecastdns.net.
ns4.complex.com. 198.7.29.6 AS15133 ns4.edgecastdns.net.
ns4.complex.com. 2606:2800:c::6 AS15133 ns4.edgecastdns.net.

"""

import os.path
import sys
import getopt
import socket
import struct
import dns.resolver, dns.reversename

PYVERSION      = sys.version_info.major
PROGNAME       = os.path.basename(sys.argv[0])
VERSION        = "0.1"

IP2ASN_V4_SUFFIX = ".origin.asn.cymru.com."
IP2ASN_V6_SUFFIX = ".origin6.asn.cymru.com."


def usage():
    print("Usage: {} [-4|6] <domain>".format(PROGNAME))
    sys.exit(1)


def get_resolver(timeout=5, edns=False):
    """return initialized resolver object"""
    r = dns.resolver.Resolver()
    r.lifetime = timeout
    if edns:
        r.use_edns(edns=0, ednsflags=0, payload=4096)
    return r


def do_query(r, qname, qtype, qclass='IN', quiet_notfound=False):
    """Perform DNS query and return answer RRset object"""
    response = None
    try:
        answers = r.query(qname, qtype, qclass)
    except (dns.resolver.NoAnswer, dns.resolver.NXDOMAIN):
        if not quiet_notfound:
            print("{}/{}/{}: No records found.".format(qname, qtype, qclass))
    except (dns.exception.Timeout):
        print("{}/{}/{}: Query timed out.".format(qname, qtype, qclass))
    except Exception as e:
        print("{}/{}/{}: error: {}".format(qname, qtype, qclass, 
                                           type(e).__name__))
    else:
        response = answers.rrset

    return response


def reverse_octets(packedstring):
    if PYVERSION == 2:
        return ["%d" % ord(x) for x in packedstring]
    else:
        return ["%d" % x for x in packedstring]


def reverse_hexdigits(packedstring):
    if PYVERSION == 2:
        return ''.join(["%02x" % ord(x) for x in packedstring])
    else:
        return ''.join(["%02x" % x for x in packedstring])


def ip2asn(res, address):
    """
    TXT records queried return single strings of the form:
    ASN | IPprefix | CountryCode | RIR | date, e.g.
    '55 | 128.91.0.0/16 | US | arin | '
    '55 | 2607:f470::/32 | US | arin | 2008-05-01'
    """
    qname = None
    try:
        if address.find('.') != -1:
            packed = socket.inet_pton(socket.AF_INET, address)
            octetlist = reverse_octets(packed)
            qname = '.'.join(octetlist[::-1]) + IP2ASN_V4_SUFFIX
        elif address.find(':') != -1:
            packed = socket.inet_pton(socket.AF_INET6, address)
            hexlist = reverse_hexdigits(packed)
            qname = '.'.join(hexlist[::-1]) + IP2ASN_V6_SUFFIX
    except socket.error:
        pass
    if not qname:
        raise ValueError("%s isn't an IP address" % address)

    txt_rrset = do_query(res, qname, 'TXT')
    if txt_rrset:
        return "AS" + txt_rrset[0].strings[0].split('|')[0].rstrip(' ')
    else:
        return None


def ip2name(res, address):
    ptr_rrset = do_query(res, dns.reversename.from_address(address), 'PTR')
    if ptr_rrset:
        return ptr_rrset[0].target
    else:
        return None


if __name__ == '__main__':

    try:
        (options, args) = getopt.getopt(sys.argv[1:], '46')
    except getopt.GetoptError:
        usage()
    if len(args) != 1:
        usage()

    addrtypes = [ 'AAAA', 'A' ]
    for (opt, optval) in options:
        if opt == "-4":
            addrtypes = [ 'A' ]
        elif opt == "-6":
            addrtypes = [ 'AAAA' ]

    qname = args[0]

    res = get_resolver()

    ns_rrset = do_query(res, qname, 'NS')
    if not ns_rrset:
        sys.exit(1)

    hostlist = []
    for ns_rr in ns_rrset:
        for rrtype in addrtypes:
            ip_rrset = do_query(res, ns_rr.target, rrtype, quiet_notfound=True)
            if ip_rrset:
                for ip_rr in ip_rrset:
                    hostlist.append( (ns_rr.target, ip_rr.address) )

    for (ns, addr) in sorted(hostlist):
        print("{} {} {} {}".format(
            ns, addr, ip2asn(res, addr), ip2name(res, addr)))

