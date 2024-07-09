package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

	"math/rand"

	"github.com/mikegin/utils"
)

func main() {
	args := os.Args[1:]

	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s [uniform/cluster] [random seed] [number of coordinate pairs to generate]\n", os.Args[0])
		return
	}

	distributionValue := args[0]

	if distributionValue != "uniform" && distributionValue != "cluster" {
		fmt.Fprintf(os.Stderr, "Invalid distribution type")
		return
	} else if distributionValue != "uniform" {
		fmt.Fprintf(os.Stderr, "Cluster distributino not yet supported")
		return
	}

	seed := args[1]
	randomSeed, err := strconv.ParseInt(seed, 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error covnerting seed to number %d\n", randomSeed)
		return
	}

	r := rand.New(rand.NewSource(randomSeed))

	maxPairCount := int64(1 << 34)
	pairCount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing pair count %s\n", args[2])
		return
	}

	if pairCount >= maxPairCount {
		fmt.Fprintf(os.Stderr, "To avoid accidentally generating massive files, number of pairs must be less than %d.\n", maxPairCount)
		return
	}

	fmt.Println("Distribution:", distributionValue)
	fmt.Println("Random seed:", randomSeed)
	fmt.Println("Number of coordinate pairs:", pairCount)

	fpairs, err := os.Create(fmt.Sprintf("data_%d_%s.%s", pairCount, "pairs", "json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pairs json file")
		return
	}
	defer fpairs.Close()

	fanswers, err := os.Create(fmt.Sprintf("data_%d_%s.%s", pairCount, "haveranswers", "f64"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating haveranswers f64 file")
		return
	}
	defer fanswers.Close()

	fpairs.Write([]byte("{\n\t\"pairs\": ["))

	sum := float64(0)
	sumCoef := 1.0 / float64(pairCount)

	for i := int64(0); i < pairCount; i++ {
		x0 := r.Float64() * 180
		y0 := r.Float64() * 90
		x1 := r.Float64() * 180
		y1 := r.Float64() * 90

		haversineDistance := utils.ReferenceHaversine(x0, y0, x1, y1, utils.EARTH_RADIUS)

		sum += sumCoef * haversineDistance

		ending := ""
		if i < pairCount-1 {
			ending = ","
		}

		s := fmt.Sprintf("\n\t\t{ \"x0\": %f, \"y0\": %f, \"x1\": %f, \"y1\": %f}%s", x0, y0, x1, y1, ending)
		fpairs.Write([]byte(s))
	}

	fpairs.Write([]byte("\n\t]\n}"))

	var buf bytes.Buffer
	err = binary.Write(&buf, binary.LittleEndian, sum)
	if err != nil {
		fmt.Fprintf(os.Stderr, "binary.Write failed for haversine sum: %v", err)
	}
	fanswers.Write([]byte(buf.Bytes()))

	fmt.Println("Expected sum:", sum)

}
