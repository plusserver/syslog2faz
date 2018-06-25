// dheilema 2018
// 2018 by Nexinto GmbH
//
// basic functions to create the FAZ compatible log

package faz

import "time"

type (
	Log map[string]string
)

// create formatted string from map
func (l Log) String() string {
	now := time.Now().String()
	l["date"] = string(now[:10])
	l["time"] = string(now[11:19])

	out := ""
	for name, value := range l {
		if out != "" {
			out += " "
		}
		if needsQuoting(name) {
			value = "\"" + value + "\""
		}
		if name == "proto" {
			value = fixProto(value)
		}
		out += name + "=" + value
	}
	return out
}

// convert syslog severity as number (as string) to it's verbose name
func VerboseLogLevel(in string) (out string) {
	switch in {
	case "1":
		out = "emerg"
	case "2":
		out = "alert"
	case "3":
		out = "crit"
	case "4":
		out = "err"
	case "5":
		out = "warning"
	case "6":
		out = "notice"
	case "7":
		out = "info"
	case "8":
		out = "debug"
	}
	return
}

// some values need quotes
func needsQuoting(name string) bool {
	switch name {
	case "msg", "srcintf", "dstintf":
		return true
	default:
		return false
	}
}

// some proto needs sanitizing
func fixProto(name string) string {
	switch name {
	case "tcp", "TCP":
		return "6"
	case "udp", "UDP":
		return "17"
	case "icmp", "ICMP":
		return "1"
	case "ipv6-icmp", "IPv6-ICMP":
		return "58"
	}
	return name
}
