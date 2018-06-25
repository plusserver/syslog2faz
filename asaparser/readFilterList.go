// dheilema 2018
// 2018 by Nexinto GmbH

package asaparser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// read the filters from a file fname
// runs the test if requested
func readFilterList(fname string, runTests bool) error {
	dat, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	lines := strings.Split(string(dat), "\n")
	linenr := 0
	msgid := ""
	step := 0
	regex := ""
	test := ""
	paramList := []string{}
	extras := []string{}
	for _, line := range lines {
		linenr++
		if len(line) > 1 {
			line = line[0 : len(line)-1]
			switch line[0] {
			case 35:
				// comment
			case 9:
				// remove tab
				line = line[1:]
				switch step {
				case 0:
					// regex
					regex = line
				case 1:
					// paramList
					paramList = strings.Split(line, ";")
				case 2:
					// extras
					extras = strings.Split(line, ";")
					new := append(ml[msgid], LogFilter{paramList, extras, regexp.MustCompile(regex)})
					ml[msgid] = new
				case 3:
					// test syslog
					test = "%ASA-1-" + msgid + ": " + line
				case 4:
					// test expected parameters
					if runTests {
						log, err := Parse(test, 0)
						if err != nil {
							return errors.New("error while testing " + msgid + " line " + fmt.Sprintf("%d", linenr) + ". " + fmt.Sprint(err))
						}
						expected := strings.Split(line, ";")
						for i, key := range paramList {
							if log[key] != expected[i] {
								out := "error while testing " + msgid + " line " + fmt.Sprintf("%d", linenr) + "\n" + test + "\n"
								out = out + "parse result: " + key + "=\"" + log[key] + "\"\n"
								out = out + "expected    : " + key + "=\"" + expected[i] + "\""
								return errors.New(out)
							}
						}
					}
					// allow another test / go back to step 2
					step = 2
				}
				step++
			default:
				// msgid go to step 0
				msgid = line
				step = 0
			}
		}
	}
	return nil
}
