package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/mikegin/gjson"
	"github.com/mikegin/profiler"
	"github.com/mikegin/utils"
)

var p *profiler.Profiler

func GetHaversinePairs(data []byte) []gjson.Result {
	p.TimeFunctionStart(profiler.GetCurrentFunctionFrame().Function, 0)
	defer p.TimeFunctionEnd(0)

	return gjson.Get(string(data), "pairs").Array()
}

func main() {

	p = &profiler.Profiler{
		StartTSC: uint64(profiler.ReadCPUTimer()),
	}

	args := os.Args

	if len(args) != 2 && len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s [haversine_input.json]\n", args[0])
		fmt.Fprintf(os.Stderr, "       %s [haversine_input.json] [answers.f64]\n", args[0])
		return
	}

	p.TimeFunctionStart("Haversine File Read", 1)
	file, err := os.Open(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %v", err)
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file stat %v", err)
		return
	}

	b := make([]byte, fi.Size())

	for {
		_, err := file.Read(b)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
	}
	p.TimeFunctionEnd(1)

	pairs := GetHaversinePairs(b)

	p.TimeFunctionStart("Haversine Sum", 4)
	sum := float64(0)
	sumCoef := 1 / float64(len(pairs))
	for _, p := range pairs {
		p := p.Value().(map[string]interface{})
		sum += sumCoef * utils.ReferenceHaversine(p["x0"].(float64), p["y0"].(float64), p["x1"].(float64), p["y1"].(float64), utils.EARTH_RADIUS)
	}
	p.TimeFunctionEnd(4)

	fmt.Println("Input size:", fi.Size())
	fmt.Println("Pair count:", len(pairs))
	fmt.Println("Haversine sum:", sum)

	// validation
	if len(args) == 3 {
		p.TimeFunctionStart("Haversine Answer Read", 2)
		validateFile, err := os.Open(args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening answers f64 file: %s", args[2])
		}

		b := make([]byte, 8)

		for {
			_, err := validateFile.Read(b)
			if err != nil {
				if err != io.EOF {
					fmt.Println(err)
				}
				break
			}
		}

		bits := binary.LittleEndian.Uint64(b)
		refSum := math.Float64frombits(bits)

		fmt.Println("Reference sum:", refSum)
		fmt.Println("Difference:", refSum-sum)
		p.TimeFunctionEnd(2)
	}

	p.EndTSC = uint64(profiler.ReadCPUTimer())
	p.PrintProfile()
}
