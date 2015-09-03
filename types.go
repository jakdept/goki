package main

import ()

type dnsRec struct {
	TTL int
	value string
	String() string
}

type aRec struct {
	dnsRec
}

type cnameRec struct {
	dnsRec
}

type mxRec struct {
	dnsRec
	priority int
}

type txtRec struct {
	dnsRec
}

type nsRec struct {
	cnameRec
}

type ptrRec struct {
	aRec
}

type soaRec struct {

}

type fqdn struct {
	parentPart string
	localPart string
	records []dnsRec
	subdomains []fqdn
}

type zone struct {
	soa soaRec
	defaultTTL int
	tld fqdn
}