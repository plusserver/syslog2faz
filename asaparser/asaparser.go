// the parser part for Cisco ASA logs
// dheilema 2018
// 2018 by Nexinto GmbH

package asaparser

import (
	"errors"
	"regexp"
	"strings"
	"syslog2faz/faz"
)

type (
	// LogFilter contains the necessary infos to parse a log line
	LogFilter struct {
		pList    []string
		extra    []string
		compiled *regexp.Regexp
	}
	// lfList is an array of all LogFilters of the same message id
	lfList  []LogFilter
	msgList map[string]lfList
)

var (
	ml msgList
)

// read file <fname> into msgList.
// Set <runTests> to true to run the regex tests
func New(fname string, runTests bool) error {
	ml = make(map[string]lfList)
	return readFilterList(fname, runTests)
}

// parse a verbose log line and return faz log structure
func Parse(in string, offset int) (faz.Log, error) {
	l := make(map[string]string)
	// fix possible offset problem
	if len(in) > offset {
		in = in[offset:]
	} else {
		return l, errors.New("offset > len()")
	}
	// remove CR
	if (in[len(in)-1]) == 10 {
		in = in[:len(in)-1]
	}
	// check if the ASA message ID is present
	msgid := strings.Split(in[1:13], "-")
	if len(msgid) == 3 {
		if msgid[0] != "ASA" {
			return l, errors.New("Possible offset problem! (" + in[1:13] + ") is not a Message ID (ASA-n-nnnnnn)")
		}
	} else {
		// nothing to parse, stay silent
		return l, nil
	}
	// cut the message part
	msg := in[15:]

	// see if we can parse this
	lfList, exists := ml[msgid[2]]
	if !exists {
		return l, errors.New("no parser for msgid " + msgid[2])
	}

	// we might have multiple filter for one message ID
	found := false
	for _, LogFilter := range lfList {
		if LogFilter.compiled.Match([]byte(msg)) {
			found = true
			matches := LogFilter.compiled.FindStringSubmatch(msg)
			if len(LogFilter.pList)+1 == len(matches) {
				for i, name := range LogFilter.pList {
					l[name] = matches[i+1]
				}
			} else {
				return l, errors.New("error in parser: difference between number of parameters and regexes")
			}
			// add extras
			for _, value := range LogFilter.extra {
				kv := strings.Split(value, "=")
				if len(kv) == 2 {
					l[kv[0]] = kv[1]
				}
			}
			l["level"] = faz.VerboseLogLevel(msgid[1])
		}
	}
	if found == false {
		return l, errors.New("error in parser: the filter for " + msgid[2] + " doesn't match")
	}
	return l, nil
}
