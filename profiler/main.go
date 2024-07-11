package profiler

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

type ProfileAnchor struct {
	TSCElapsed           uint64
	HitCount             uint64
	CurrentBlockStartTSC uint64
	Label                string
}

type Profiler struct {
	ProfileAnchors [4096]*ProfileAnchor
	StartTSC       uint64
	EndTSC         uint64
}

func GetCurrentFunctionFrame() runtime.Frame {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame
}

func (p *Profiler) TimeFunctionStart(name string, id int) {
	if id >= len(p.ProfileAnchors) {
		fmt.Fprintf(os.Stderr, "Out of bounds id pos in %s: %d", GetCurrentFunctionFrame().Function, id)
	}

	p.ProfileAnchors[id] = &ProfileAnchor{}
	pa := p.ProfileAnchors[id]
	pa.Label = name
	pa.CurrentBlockStartTSC = uint64(ReadCPUTimer())
}

func (p *Profiler) TimeFunctionEnd(id int) {
	if id >= len(p.ProfileAnchors) {
		fmt.Fprintf(os.Stderr, "Out of bounds id pos in %s: %d", GetCurrentFunctionFrame().Function, id)
	}

	pa := p.ProfileAnchors[id]
	pa.TSCElapsed += uint64(ReadCPUTimer()) - pa.CurrentBlockStartTSC
	pa.HitCount += 1

}

func (p *Profiler) PrintProfile() {
	totalCPUElapsed := p.EndTSC - p.StartTSC
	cpuFreq := EstimateCPUTimerFreq()

	if cpuFreq > 0 {
		fmt.Printf("\nTotal time: %0.4fms (CPU freq %d)\n", 1000.0*float64(totalCPUElapsed)/float64(cpuFreq), cpuFreq)
	}

	for _, anchor := range p.ProfileAnchors {
		if anchor != nil {
			elapsed := anchor.TSCElapsed
			if elapsed > 0 {
				percent := 100.0 * float64(elapsed) / float64(totalCPUElapsed)
				fmt.Printf("  %s[%d]: %d (%.2f%%)\n", anchor.Label, anchor.HitCount, elapsed, percent)
			}
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
