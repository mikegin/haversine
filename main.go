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

func GetHaversinePairs(data []byte) []gjson.Result {
	profiler.GlobalProfiler.TimeFunctionStart(profiler.GetCurrentFunctionFrame().Function, 0)
	defer profiler.GlobalProfiler.TimeFunctionEnd(0)

	return gjson.Get(string(data), "pairs").Array()
}

func main() {
	args := os.Args

	if len(args) != 2 && len(args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s [haversine_input.json]\n", args[0])
		fmt.Fprintf(os.Stderr, "       %s [haversine_input.json] [answers.f64]\n", args[0])
		return
	}

	profiler.GlobalProfiler.TimeFunctionStart("Haversine Pairs File Read", 1)
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
	profiler.GlobalProfiler.TimeFunctionEnd(1)

	pairs := GetHaversinePairs(b)

	sum := float64(0)
	sumCoef := 1 / float64(len(pairs))
	func() {
		profiler.GlobalProfiler.TimeFunctionStart("Haversine Sum", 4)
		defer profiler.GlobalProfiler.TimeFunctionEnd(4)
		for _, p := range pairs {
			p := p.Value().(map[string]interface{})
			sum += sumCoef * utils.ReferenceHaversine(p["x0"].(float64), p["y0"].(float64), p["x1"].(float64), p["y1"].(float64), utils.EARTH_RADIUS)
		}
	}()

	fmt.Println("Input size:", fi.Size())
	fmt.Println("Pair count:", len(pairs))
	fmt.Println("Haversine sum:", sum)

	// validation
	if len(args) == 3 {
		profiler.GlobalProfiler.TimeFunctionStart("Haversine Answer File Read", 2)
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
		profiler.GlobalProfiler.TimeFunctionEnd(2)
	}

	profiler.GlobalProfiler.EndTSC = uint64(profiler.ReadCPUTimer())
	profiler.GlobalProfiler.PrintProfile()
}
