package main

import (
	"fmt"
	"math/rand"
	"time"

	"satoshi-pes/ringo/common"
	"satoshi-pes/ringo/utils"
)

func main() {

	codes := []string{
		"C1",
		"P1",
		"L1",
		"D1",
		"S1",
		"C2",
		"P2",
		"L2",
		"D2",
		"S2",
		"C5",
		"L5",
		"D5",
		"S5",
	}

	codesBad := []string{
		"C1",
		"P1",
		"L1",
		"D1",
		"S1",
		"C2",
		"XX",
		"P2",
		"L2",
		"D2",
		"ZX",
		"S2",
		"C5",
		"L5",
		"D5",
		"S5",
	}

	// shuffle
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(codes), func(i, j int) {
		codes[i], codes[j] = codes[j], codes[i]
	})
	rand.Shuffle(len(codesBad), func(i, j int) {
		codesBad[i], codesBad[j] = codesBad[j], codesBad[i]
	})

	// convert common.ObsCode
	table := common.ObsCodeTableGPS2xx

	// test sort obscodes
	codesTest := []common.ObsCode{}
	for _, codeStr := range codes {
		codesTest = append(codesTest, table[codeStr])
	}

	fmt.Println("test util.SortObsCodesStrArgsorted...")
	sorted2, sortedIdx2 := utils.SortObsCodesStrArgsorted(codesBad)
	for i, codeStr := range codesBad {
		fmt.Printf("%d: %s - %s %d\n", i, codeStr, sorted2[i], sortedIdx2[i])
	}

	fmt.Println("test util.SortObsCodesArgsorted...")
	sortedCodes, sortedIndex := utils.SortObsCodesArgsorted(codesTest)
	for i, code := range codesTest {
		fmt.Printf("%d: %s - %s %d\n", i, code.Str, sortedCodes[i].Str, sortedIndex[i])
	}
}
