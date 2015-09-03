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
