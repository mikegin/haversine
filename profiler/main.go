package profiler

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

// refers to a single call to a function, used to handle nested/recursive calls
type ProfileBlock struct {
	StartTSC            uint64
	OldTSCElapsedAtRoot uint64
	Parent              *ProfileAnchor
}

type ProfileAnchor struct {
	TSCElapsed         uint64
	TSCElapsedChildren uint64
	TSCElapsedAtRoot   uint64
	HitCount           uint64
	Label              string
	Blocks             []ProfileBlock
}

type Profiler struct {
	ProfileAnchors [4096]*ProfileAnchor
	StartTSC       uint64
	EndTSC         uint64
}

var GlobalProfiler *Profiler = &Profiler{
	StartTSC: uint64(ReadCPUTimer()),
}
var GlobalProfilerParent *ProfileAnchor = nil

func GetCurrentFunctionFrame() runtime.Frame {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame
}

func (p *Profiler) TimeFunctionStart(name string, id int) {
	if os.Getenv("PROFILER") != "true" {
		return
	}
	if id >= len(p.ProfileAnchors) {
		fmt.Fprintf(os.Stderr, "Out of bounds id pos in %s: %d", GetCurrentFunctionFrame().Function, id)
	}

	if p.ProfileAnchors[id] == nil {
		p.ProfileAnchors[id] = &ProfileAnchor{}
	}
	pa := p.ProfileAnchors[id]
	pa.Label = name

	b := ProfileBlock{
		StartTSC:            uint64(ReadCPUTimer()),
		Parent:              GlobalProfilerParent,
		OldTSCElapsedAtRoot: pa.TSCElapsedAtRoot,
	}
	GlobalProfilerParent = pa
	pa.Blocks = append(pa.Blocks, b)
}

func (p *Profiler) TimeFunctionEnd(id int) {
	if os.Getenv("PROFILER") != "true" {
		return
	}
	if id >= len(p.ProfileAnchors) {
		fmt.Fprintf(os.Stderr, "Out of bounds id pos in %s: %d", GetCurrentFunctionFrame().Function, id)
	}

	pa := p.ProfileAnchors[id]
	endTSC := uint64(ReadCPUTimer())

	// pop off current block
	block := pa.Blocks[len(pa.Blocks)-1]
	pa.Blocks = pa.Blocks[:len(pa.Blocks)-1]

	elapsed := endTSC - block.StartTSC

	if block.Parent != nil {
		block.Parent.TSCElapsedChildren += elapsed
	}

	pa.TSCElapsed += elapsed

	pa.TSCElapsedAtRoot = block.OldTSCElapsedAtRoot + elapsed

	GlobalProfilerParent = block.Parent

	pa.HitCount += 1

}

func (p *Profiler) PrintProfile() {
	if os.Getenv("PROFILER") != "true" {
		return
	}
	totalTSCElapsed := p.EndTSC - p.StartTSC
	cpuFreq := EstimateCPUTimerFreq()

	if cpuFreq > 0 {
		fmt.Printf("\nTotal time: %0.4fms (CPU freq %d)\n", 1000.0*float64(totalTSCElapsed)/float64(cpuFreq), cpuFreq)
	}

	for _, anchor := range p.ProfileAnchors {
		if anchor != nil {
			elapsedSelf := anchor.TSCElapsed - anchor.TSCElapsedChildren
			percent := 100.0 * float64(elapsedSelf) / float64(totalTSCElapsed)
			fmt.Printf("  %s[%d]: %d (%.2f%%", anchor.Label, anchor.HitCount, elapsedSelf, percent)

			if anchor.TSCElapsedAtRoot != elapsedSelf {
				fmt.Printf(", %.2f%% w/children", 100.0*float64(anchor.TSCElapsedAtRoot)/float64(totalTSCElapsed))
			}

			fmt.Printf(")")

			if cpuFreq > 0 {
				fmt.Printf(" %0.4fms", 1000.0*float64(elapsedSelf)/float64(cpuFreq))
			}

			fmt.Printf("\n")
		}
	}
}

func GetOSTimerFreq() uint64 {
	return 1000000
}

// ReadOSTimer returns the current value (ticks) of the OS timer.
func ReadOSTimer() uint64 {
	var tv syscall.Timeval
	syscall.Gettimeofday(&tv)
	return GetOSTimerFreq()*uint64(tv.Sec) + uint64(tv.Usec)
}

func ReadCPUTimer() int64 // assembly function

// EstimateCPUTimerFreq estimates the CPU timer frequency.
func EstimateCPUTimerFreq() uint64 {
	millisecondsToWait := uint64(100)
	osFreq := GetOSTimerFreq()

	cpuStart := ReadCPUTimer()
	osStart := ReadOSTimer()
	osElapsed := uint64(0)
	osWaitTime := osFreq * millisecondsToWait / 1000

	for osElapsed < osWaitTime {
		osElapsed = ReadOSTimer() - osStart
	}

	cpuEnd := ReadCPUTimer()
	cpuElapsed := cpuEnd - cpuStart

	var cpuFreq uint64
	if osElapsed > 0 {
		cpuFreq = osFreq * uint64(cpuElapsed) / osElapsed
	}

	return cpuFreq
}
