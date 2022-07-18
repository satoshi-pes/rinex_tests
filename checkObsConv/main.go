package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	//showrinex()
	showrinex2()
}

func showrinex() {
	fname1 := os.Args[1]
	//fmt.Println(fname1, fname2)

	f1, err := os.Open(fname1)
	if err != nil {
		log.Printf("cannot open file='%s'. err='%v'", fname1, err)
		return
	}
	defer f1.Close()
	s1 := bufio.NewScanner(f1)

	version, obsCodeList1, sysList1 := ReadRinexHeader(s1)
	_ = version

	if len(sysList1) == 0 {
		numCodes := len(obsCodeList1[0])
		obs1, err1 := ParseObsRinex2(s1, numCodes)

		if err1 != nil {
			log.Printf("failed to parse obs RINEX2: err='%v'\n", err1)
		}

		fmt.Printf("epoch: %v\n", obs1.TimeTag)
		for _, PRN := range obs1.PRNs {
			fmt.Printf("PRN: %v\n", PRN)
			d := obs1.ObsData[PRN]
			for i, code := range obsCodeList1[0] {
				fmt.Printf("%2d: code='%3s', val='%s'\n", i, code, d[i])
			}
		}
	} else {
		obsCodess := make(map[rune][]string)
		for i, sys := range sysList1 {
			obsCodess[sys] = obsCodeList1[i]
		}
		obs1, err1 := ParseObs(s1, obsCodess)

		if err1 != nil {
			log.Printf("failed to parse obs RINEX: err='%v'\n", err1)
		}

		fmt.Printf("epoch: %v\n", obs1.TimeTag)
		for _, PRN := range obs1.PRNs {
			fmt.Printf("PRN: %v\n", PRN)
			d := obs1.ObsData[PRN]
			obsCode := obsCodess[rune(PRN[0])]
			for i, code := range obsCode {
				fmt.Printf("%2d: code='%3s', val='%s'\n", i, code, d[i])
			}
		}
	}
}

func showrinex2() {
	fname1 := os.Args[1]
	fname2 := os.Args[2]

	compTwoRinex(fname1, fname2)
}

func compTwoRinex(filename1, filename2 string) {
	v1, o1, e1 := readRinexfile(filename1)
	v2, o2, e2 := readRinexfile(filename2)

	_, _ = e1, e2

	//fmt.Printf("file1: %s, %v, %v\n", v1, o1, e1)
	//fmt.Printf("file2: %s, %v, %v\n", v2, o2, e2)

	if !o1.TimeTag.Equal(o2.TimeTag) {
		fmt.Printf("the first epoch is different: %v", o2.TimeTag)
		return
	}

	// compare
	for _, PRN := range o1.PRNs {
		fmt.Printf("PRN: %v\n", PRN)

		d1 := o1.ObsData[PRN]
		d2, ok := o2.ObsData[PRN]
		if !ok {
			// skip
			continue
		}

		sys := PRN[:1]
		ss := []string{}
		for i, code := range o1.ObsCodes[sys] {
			s := fmt.Sprintf("%2d: code='%3s', val='%s'", i, code, d1[i])

			//code1new, ok1 := getObsCode(v1, sys, code)
			code1new, _ := getObsCode(v1, sys, code)

			//var data2 string
			for j, code2 := range o2.ObsCodes[sys] {
				//code2new, ok2 := getObsCode(v1, sys, code2)
				code2new, _ := getObsCode(v2, sys, code2)
				if code1new == code2new {
					//data2 = d2[j]
					s += fmt.Sprintf(" -- code='%3s', val='%s'", code2, d2[j])
					break
				}
			}
			ss = append(ss, s)
		}

		for _, s := range ss {
			fmt.Printf("%s\n", s)
		}
	}

	return
}

func getObsCode(v string, sys string, code string) (code3ch string, ok bool) {
	switch v {
	case "2", "2.10", "2.11":
		switch sys {
		case "G":
			code3ch, ok = ObsCodeTableGPS2xx[code]
		case "R":
			code3ch, ok = ObsCodeTableGLO2xx[code]
		case "S":
			code3ch, ok = ObsCodeTableSBAS2xx[code]
		case "E":
			code3ch, ok = ObsCodeTableGAL2xx[code]
		}
	case "2.12":
		switch sys {
		case "G":
			code3ch, ok = ObsCodeTableGPS212[code]
		case "R":
			code3ch, ok = ObsCodeTableGLO212[code]
		case "J":
			code3ch, ok = ObsCodeTableQZS212[code]
		case "C":
			code3ch, ok = ObsCodeTableBDS212[code]
		case "S":
			code3ch, ok = ObsCodeTableSBAS212[code]
		case "E":
			code3ch, ok = ObsCodeTableGAL212[code]
		}
	default:
		// ver >= 3
		return code, true
	}
	return
}

func readRinexfile(filename string) (version string, obsd obsData, err error) {
	f1, err := os.Open(filename)
	if err != nil {
		log.Printf("cannot open file='%s'. err='%v'", filename, err)
		return
	}
	defer f1.Close()
	s1 := bufio.NewScanner(f1)

	fmt.Println(filename)
	version, obsCodeList1, sysList1 := ReadRinexHeader(s1)

	if len(sysList1) == 0 {
		numCodes := len(obsCodeList1[0])
		obsd, err = ParseObsRinex2(s1, numCodes)

		// obscodes
		obsd.ObsCodes = map[string][]string{}
		obsd.ObsCodes["G"] = obsCodeList1[0]
		obsd.ObsCodes["R"] = obsCodeList1[0]
		obsd.ObsCodes["S"] = obsCodeList1[0]
		obsd.ObsCodes["E"] = obsCodeList1[0]
		if version >= "2.12" {
			obsd.ObsCodes["J"] = obsCodeList1[0]
			obsd.ObsCodes["C"] = obsCodeList1[0]
		}

		if err != nil {
			log.Printf("failed to parse obs RINEX2: err='%v'\n", err)
			return
		}
	} else {
		obsCodess := make(map[rune][]string)
		for i, sys := range sysList1 {
			obsCodess[sys] = obsCodeList1[i]
		}
		obsd, err = ParseObs(s1, obsCodess)

		// obscodes
		obsd.ObsCodes = map[string][]string{}
		for i, sys := range sysList1 {
			obsd.ObsCodes[string(sys)] = obsCodeList1[i]
		}

		if err != nil {
			log.Printf("failed to parse obs RINEX: err='%v'\n", err)
			return
		}
	}

	return
}

func ReadRinexHeader(scanner *bufio.Scanner) (version string, obsCodesList [][]string, sysList []rune) {
	var buf string

	for scanner.Scan() {
		buf = scanner.Text()

		if len(buf) < 61 {
			// skip invalid header
			continue
		}

		label := strings.TrimSpace(buf[60:])

		// get obs types
		switch label {
		case "RINEX VERSION / TYPE":
			version = strings.TrimSpace(buf[:9])
		case "# / TYPES OF OBSERV":
			// version2
			obsCodes, err := ParseObsTypesRinex2Header(scanner, buf)
			if err != nil {
				log.Printf("error in reading '# / TYPES OF OBSERV'. err='%v'\n", err)
				continue
			}
			obsCodesList = append(obsCodesList, obsCodes)
		case "SYS / # / OBS TYPES":
			satSys, obsCodes, err := ParseObsTypesRinex3Header(scanner, buf)
			fmt.Println(satSys, string(satSys))
			if err != nil {
				log.Printf("error in reading 'SYS / # / OBS TYPES'. err='%v'\n", err)
				continue
			}
			sysList = append(sysList, satSys)
			obsCodesList = append(obsCodesList, obsCodes)
		case "END OF HEADER":
			return
		}
	}
	return
}

// ParseObsTypesRinex2Header parses obstypes in RINEX2 obs header "# / TYPES OF OBSERV"
func ParseObsTypesRinex2Header(scanner *bufio.Scanner, s string) (obsCodes []string, err error) {
	// number of codes
	var numCodes int

	if len(s) < 6 {
		err = fmt.Errorf("too short header, s='%s'", s)
		return
	}

	numCodes, err = strconv.Atoi(strings.TrimSpace(s[:6]))
	if err != nil {
		err = fmt.Errorf("failed to parse numCodes, s='%s', err=%v", s[:6], err)
		return
	}

	n := 0    // number of codes in the current line
	idx := 10 // index of the string
	for i := 0; i < numCodes; i++ {
		// check length
		if len(s) < idx+2 {
			err = fmt.Errorf("too short msg, s='%s'", s)
			return
		}

		obsCodes = append(obsCodes, s[idx:idx+2])

		n++
		idx += 6
		// line feed
		if n == 9 && i+1 < numCodes {
			scanner.Scan()
			s = scanner.Text() // move to the new line
			n, idx = 0, 10
		}
	}

	return
}

// ParseObsTypesRinex3Header parses obstypes in RINEX3 obs header "SYS / # / OBS TYPES"
func ParseObsTypesRinex3Header(scanner *bufio.Scanner, s string) (satSys rune, obsCodes []string, err error) {
	var numCodes int

	if len(s) < 6 {
		err = fmt.Errorf("too short msg, s='%s'", s)
		return
	}

	// parse satsys code
	satSys = rune(s[0]) // 'G', 'R', 'J', 'E', 'C'
	numCodes, err = strconv.Atoi(strings.TrimSpace(s[3:6]))
	if err != nil {
		err = fmt.Errorf("cannot parse numCodes, err=%w", err)
		return
	}

	n := 0   // number of codes in the current line
	idx := 7 // index of the string
	for i := 0; i < numCodes; i++ {
		if len(s) < idx+3 {
			err = fmt.Errorf("too short msg, s='%s'", s)
			return
		}
		obsCodes = append(obsCodes, s[idx:idx+3])

		n++
		idx += 4
		if n == 13 && i+1 < numCodes {
			scanner.Scan()
			s = scanner.Text() // move to the new line
			n, idx = 0, 7
		}
	}

	return
}

// obsData stores rinexobs data for an epoch
type obsData struct {
	TimeTag  time.Time
	PRNs     []string
	ObsCodes map[string][]string // ObsCodes[sys] = []string{codes...}
	ObsData  map[string][]string // ObsData[sys] = []{data...}
}

func ParseObsRinex2(scanner *bufio.Scanner, numCodes int) (obs obsData, err error) {
	var buf string

	for scanner.Scan() {
		buf = scanner.Text()

		// parse timetag, etc
		timeTag, epochFlag, numSat, PRNs, e := parseEpochHeaderRinex2(scanner, buf)
		if e != nil {
			err = e
			return
		}

		// skip special event
		if epochFlag > 0 {
			// skip special event
			for i := 0; i < numSat; i++ {
				scanner.Scan()
			}
			continue
		}

		d := obsData{
			TimeTag: timeTag,
			PRNs:    PRNs,
			ObsData: map[string][]string{},
		}

		// read observation data
		for _, PRN := range PRNs {
			_ = PRN
			d.ObsData[PRN] = make([]string, numCodes)
			obs := make([]string, numCodes)

			// parse a line
			scanner.Scan()
			buf = scanner.Text()

			n, idx := 0, 0
			for i := 0; i < numCodes; i++ {
				// data
				switch {
				case len(buf) > idx+15:
					obs[i] = buf[idx : idx+16] // with SS
				case len(buf) > idx+14:
					obs[i] = buf[idx : idx+15] // with LLI
				case len(buf) > idx+13:
					obs[i] = buf[idx : idx+14] // only Data
				}

				n++
				idx += 16

				// continuation line
				if n == 5 && i+1 < numCodes {
					scanner.Scan()
					buf = scanner.Text()
					n, idx = 0, 0
				}
			}

			d.ObsData[PRN] = obs
		}

		return d, nil
	}

	return
}

func parseEpochHeaderRinex2(scanner *bufio.Scanner, buf string) (timeTag time.Time, epochFlag, numSat int, PRNs []string, err error) {
	// read timetag
	yy, err1 := strconv.Atoi(strings.TrimSpace(buf[:3]))
	mm, err2 := strconv.Atoi(strings.TrimSpace(buf[4:6]))
	dd, err3 := strconv.Atoi(strings.TrimSpace(buf[7:9]))
	HH, err4 := strconv.Atoi(strings.TrimSpace(buf[10:12]))
	MM, err5 := strconv.Atoi(strings.TrimSpace(buf[13:15]))
	ss, err6 := strconv.ParseFloat(strings.TrimSpace(buf[15:26]), 64)
	epochFlag, err7 := strconv.Atoi(strings.TrimSpace(buf[28:29]))
	numSat, err8 := strconv.Atoi(strings.TrimSpace(buf[29:32]))

	// error check
	for i, e := range []error{err1, err2, err3, err4, err5, err6, err7, err8} {
		if e != nil {
			log.Printf("error at %d\n", i)
			err = e
			return
		}
	}

	if epochFlag > 1 {
		return
	}

	// timetag
	if yy >= 80 {
		yy += 1900
	} else {
		yy += 2000
	}
	sec := int(ss)
	nsec := int((ss - float64(sec)) * 1.e9)

	timeTag = time.Date(yy, time.Month(mm), dd, HH, MM, sec, nsec, time.UTC)

	// get list of satellites
	n, idx := 0, 32
	for i := 0; i < numSat; i++ {
		if len(buf) < idx+3 {
			err = fmt.Errorf("failed to parse PRN: s='%s'", buf)
			return
		}
		PRNs = append(PRNs, buf[idx:idx+3])

		n++
		idx += 3
		// continuation line
		if n == 12 && i+1 < numSat {
			scanner.Scan()
			buf = scanner.Text()
			n, idx = 0, 32
		}
	}

	return
}

func ParseObs(scanner *bufio.Scanner, obsCodess map[rune][]string) (obs obsData, err error) {
	var buf string

	for scanner.Scan() {
		buf = scanner.Text()

		if strings.HasPrefix(buf, ">") {
			// found new observation block
			timeTag, epochFlag, numSat, e := ParseEpochHeader(scanner, buf)
			if e != nil {
				log.Printf("failed to read timetag: s='%s', err='%v'", buf, e)
				continue
			}

			// check epoch flag
			if epochFlag > 1 {
				// skip special event
				for i := 0; i < numSat; i++ {
					scanner.Scan()
				}
				continue
			}

			obs = obsData{
				TimeTag: timeTag,
				PRNs:    []string{},
				ObsData: map[string][]string{},
			}

			// parse timetag
			for i := 0; i < numSat; i++ {

				// parse a line
				scanner.Scan()
				buf = scanner.Text()

				PRN, obsx, e := decodeObs(buf, obsCodess)
				if e != nil {
					err = fmt.Errorf("failed to read obs: '%s', err='%w'", buf, e)
					return
				}
				obs.PRNs = append(obs.PRNs, PRN)
				obs.ObsData[PRN] = obsx
			}

			return
		}
	}

	return
}

func ParseEpochHeader(scanner *bufio.Scanner, buf string) (timeTag time.Time, epochFlag, numSat int, err error) {
	// time layout of timetag
	dateLayout := "2006  1  2 15  4  5" // YYYY mm dd HH MM SS

	// date
	timeTag, err = time.Parse(dateLayout, buf[2:29])
	if err != nil {
		err = fmt.Errorf("failed to read timetag: '%s', err='%w'", buf[2:29], err)
		return
	}

	// epoch flag
	epochFlag, err = strconv.Atoi(buf[31:32])
	if err != nil {
		err = fmt.Errorf("failed to read epochflag: '%s', err='%w'", buf[31:32], err)
		return
	}
	if epochFlag > 1 {
		return
	}

	// number of satellites
	numSat, err = strconv.Atoi(strings.TrimSpace(buf[32:35]))
	if err != nil {
		err = fmt.Errorf("failed to read numsat: '%s', err='%w'", buf[32:35], err)
		return
	}

	return
}

func decodeObs(buf string, obsCodess map[rune][]string) (PRN string, obs []string, err error) {
	if len(buf) < 3 {
		err = fmt.Errorf("invalid observation data: '%s'", buf)
		return
	}

	PRN = buf[:3]
	codes, ok := obsCodess[rune(PRN[0])]
	if !ok {
		return
	}
	numCodes := len(codes)
	obs = make([]string, numCodes)
	for i, offset := 0, 3; i < numCodes; i, offset = i+1, offset+16 {
		switch {
		case len(buf) > offset+15:
			//obs[i] = strings.TrimSpace(buf[offset : offset+16])
			obs[i] = buf[offset : offset+16]
		case len(buf) > offset+14:
			//obs[i] = strings.TrimSpace(buf[offset : offset+15])
			obs[i] = buf[offset : offset+15]
		case len(buf) > offset+13:
			//obs[i] = strings.TrimSpace(buf[offset : offset+14])
			obs[i] = buf[offset : offset+14]
		}
	}
	return
}

// obscodes
// for version 2.00, 2.10, 2.11
var ObsCodeTableGPS2xx map[string]string = map[string]string{
	"C1": "C1C", // GPS L1C/A
	"P1": "C1P", // GPS L1P
	"L1": "L1C", // GPS L1C/A...?
	"D1": "D1C",
	"S1": "S1C",
	"C2": "C2C", // GPS L2C/A
	"P2": "C2P", // GPS L2P
	"L2": "L2P",
	"D2": "D2P",
	"S2": "S2P",
	"C5": "C5X", // GPS L5(I+Q)
	"L5": "L5X",
	"D5": "D5X",
	"S5": "S5X",
}

var ObsCodeTableGLO2xx map[string]string = map[string]string{
	"C1": "C1C", // GLO G1C/A
	"P1": "C1P", // GLO G1P
	"L1": "L1P", // GLO G1P...?
	"D1": "D1P",
	"S1": "S1P",
	"C2": "C2C", // GLO G2C/A
	"P2": "C2P", // GLO G2P
	"L2": "L2P", // GLO G2P...?
	"D2": "D2P",
	"S2": "S2P",
}

var ObsCodeTableGAL2xx map[string]string = map[string]string{
	"C1": "C1X", // E1(B+C)
	"L1": "L1X",
	"D1": "D1X",
	"S1": "S1X",
	"C5": "C5X", // E5a
	"L5": "L5X",
	"D5": "D5X",
	"S5": "S5X",
	"C7": "C7X", // E5b
	"L7": "L7X",
	"D7": "D7X",
	"S7": "S7X",
	"C8": "C8X", // E5a+b
	"L8": "L8X",
	"D8": "D8X",
	"S8": "S8X",
	"C6": "C6X", // E6(B+C)
	"L6": "L6X",
	"D6": "D6X",
	"S6": "S6X",
}

var ObsCodeTableSBAS2xx map[string]string = map[string]string{
	"C1": "C1C",
	"L1": "L1C",
	"D1": "D1C",
	"S1": "S1C",
	"C5": "C5X",
	"L5": "L5X",
	"D5": "D5X",
	"S5": "S5X",
}

var ObsCodeTableGPS212 map[string]string = map[string]string{
	"P1": "C1P", // GPS L1P
	"L1": "L1P",
	"D1": "D1P",
	"S1": "S1P",
	"CA": "C1C", // GPS L1C/A
	"LA": "L1C",
	"DA": "D1C",
	"SA": "S1C",
	"CB": "C1X", // GPS L1C(D+P)
	"LB": "L1X",
	"DB": "D1X",
	"SB": "S1X",
	"C2": "C2D", // GPS L2C/A (L1C/A + (P2-P1))
	"P2": "C2P", // GPS L2P
	"L2": "L2P",
	"D2": "D2P",
	"S2": "S2P",
	"CC": "C2X",
	"LC": "L2X",
	"DC": "D2X",
	"SC": "S2X",
	"C5": "C5X", // GPS L5(I+Q)
	"L5": "L5X",
	"D5": "D5X",
	"S5": "S5X",
}

var ObsCodeTableGLO212 map[string]string = map[string]string{
	"P1": "C1P", // GLO G1 HA (G1P)
	"L1": "L1P",
	"D1": "D1P",
	"S1": "S1P",
	"CA": "C1C", // GLO G1 SA (G1C/A)
	"LA": "L1C",
	"DA": "D1C",
	"SA": "S1C",
	"P2": "C2P", // GLO G2 HA (G2P)
	"L2": "L2P",
	"D2": "D2P",
	"S2": "S2P",
	"CD": "C2C", // GLO G2 SA (G2C/A)
	"LD": "L2C",
	"DD": "D2C",
	"SD": "S2C",
}

var ObsCodeTableGAL212 map[string]string = ObsCodeTableGAL2xx
var ObsCodeTableSBAS212 map[string]string = ObsCodeTableSBAS2xx

var ObsCodeTableBDS212 map[string]string = map[string]string{
	// Compass E2 = BDS B1
	// Compass E5b = BDS B2
	// Compass E6 = BDS B3
	/*
		"C2": "C2X", // Compass E2 I/Q = BDS B1
		"L2": "L2X",
		"D2": "D2X",
		"S2": "S2X",
		"C7": "C7X", // Compass E5b I/Q = BDS B2
		"L7": "L7X",
		"D7": "D7X",
		"S7": "S7X",
		"C6": "C6X", // Compass E6 I/Q = BDS B3
		"L6": "L6X",
		"D6": "D6X",
		"S6": "S6X",
	*/
	"C2": "C2I", // Compass E2 I/Q = BDS B1
	"L2": "L2I",
	"D2": "D2I",
	"S2": "S2I",
	"C7": "C7I", // Compass E5b I/Q = BDS B2
	"L7": "L7I",
	"D7": "D7I",
	"S7": "S7I",
	"C6": "C6I", // Compass E6 I/Q = BDS B3
	"L6": "L6I",
	"D6": "D6I",
	"S6": "S6I",
}

var ObsCodeTableQZS212 map[string]string = map[string]string{
	"CA": "C1C", // L1C/A
	"LA": "L1C",
	"DA": "D1C",
	"SA": "S1C",
	"CB": "C1X", // L1C
	"LB": "L1X",
	"DB": "D1X",
	"SB": "S1X",
	"CC": "C2X", // L2C
	"LC": "L2X",
	"DC": "D2X",
	"SC": "S2X",
	"C5": "C5X", // L5C
	"L5": "L5X",
	"D5": "D5X",
	"S5": "S5X",
	"C6": "C6X", // LEX(L6)
	"L6": "L6X",
	"D6": "D6X",
	"S6": "S6X",
}
