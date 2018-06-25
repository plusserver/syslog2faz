// syslog to FortiAnalyzer
// dheilema 2018
// 2018 by Nexinto GmbH

package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"syslog2faz/asaparser"
)

var (
	port   string
	test   bool
	config string
	faz    string
)

func main() {
	flag.StringVar(&port, "p", "10514", "Listening port")
	flag.StringVar(&config, "c", "filter.list", "Configuration file")
	flag.StringVar(&faz, "f", "", "Name/IP of FortiAnalyzer")
	flag.BoolVar(&test, "t", false, "Run regex tests")
	flag.Parse()
	// init parser, optional testing
	err := asaparser.New(config, test)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if test {
		fmt.Println("Regex Test successful")
	}
	if faz == "" {
		fmt.Println("Missing faz ip/name. Use -f <faz>")
		os.Exit(1)
	}
	netaddr, _ := net.ResolveUDPAddr("udp", ":"+port)
	conn, err := net.ListenUDP("udp", netaddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("listening on " + port)
	fmgAddr, err := net.ResolveUDPAddr("udp", faz+":514")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	connout, err := net.DialUDP("udp", nil, fmgAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("sending to   " + faz)

	buf := make([]byte, 2048)
	for {
		numRead, srcAddr, _ := conn.ReadFrom(buf)
		in := string(buf[:numRead])
		source := strings.Split(srcAddr.String(), ":")

		// remove syslog facility/severity "<nnn>"
		offset := 0
		if in[0] == 60 {
			for i := 2; i < 6; i++ {
				if in[i] == 62 {
					offset = i + 1
				}
			}
		}
		l, err := asaparser.Parse(in, offset)
		if err != nil {
			fmt.Println("****", err, in)

		} else {
			l["devname"] = source[0]
			l["devid"] = "FGT40C" + strings.Replace(source[0], ".", "", -1)
			_, err = connout.Write([]byte("<188>" + l.String() + "\n"))
			if err != nil {
				fmt.Println(err)
				return
			}
		}

	}
}
