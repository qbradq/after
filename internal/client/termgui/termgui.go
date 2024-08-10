// Package termgui implements a text-mode client over termui and drivers.
package termgui

import (
	"os"
	"runtime/pprof"
)

var cpuProfile *os.File

// beginCPUProfile starts the CPU profiling.
func beginCPUProfile() {
	var err error
	cpuProfile, err = os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	if err = pprof.StartCPUProfile(cpuProfile); err != nil {
		panic(err)
	}
}

// endCPUProfile ends the CPU profiling.
func endCPUProfile() {
	if cpuProfile != nil {
		pprof.StopCPUProfile()
		cpuProfile.Close()
		cpuProfile = nil
	}
}
