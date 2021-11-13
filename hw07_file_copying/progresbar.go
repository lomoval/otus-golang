package main

import (
	"fmt"
	"strconv"
)

const (
	progressBarLength      = 50
	progressBarSymbol      = byte('*')
	progressBarEmptySymbol = byte('-')
)

type progressBar struct {
	fullProgress    int64
	currentProgress int64
	currentPercent  int
	progressBar     []byte // [progress]
	fillLength      int    // [*filled*-non-filled-]
}

func (pb progressBar) start(progress <-chan int64, fullProgress int64) <-chan struct{} {
	pb.progressBar = make([]byte, progressBarLength+2) // [progress length]
	pb.progressBar[0] = '['
	pb.progressBar[progressBarLength+1] = ']'
	for i := 1; i <= progressBarLength; i++ {
		pb.progressBar[i] = progressBarEmptySymbol
	}

	pb.currentProgress = 0
	pb.currentPercent = 0
	pb.fillLength = 0
	pb.fullProgress = fullProgress

	pb.calcPercent()
	pb.updateProgressBytes()
	pb.printProgress()

	terminated := make(chan struct{})
	go func() {
		defer close(terminated)
		for p := range progress {
			if pb.currentProgress < pb.fullProgress {
				pb.currentProgress = p
				pb.calcPercent()
				pb.updateProgressBytes()
				pb.printProgress()
			}
		}
		fmt.Println()
	}()
	return terminated
}

func (pb progressBar) printProgress() {
	fmt.Printf("\r%s%s%%", pb.progressBar, strconv.Itoa(pb.currentPercent))
}

func (pb *progressBar) updateProgressBytes() {
	if pb.fillLength == progressBarLength {
		return
	}
	percent := int(float32(progressBarLength) * float32(pb.currentPercent) / 100)
	for ; pb.fillLength < percent && pb.fillLength < progressBarLength; pb.fillLength++ {
		pb.progressBar[pb.fillLength+1] = progressBarSymbol
	}
}

func (pb *progressBar) calcPercent() {
	if pb.currentPercent == 100 {
		return
	}
	if pb.fullProgress == 0 {
		pb.currentPercent = 100
		return
	}
	pb.currentPercent = int(float32(pb.currentProgress) / float32(pb.fullProgress) * 100)
	if pb.currentPercent > 100 {
		pb.currentPercent = 100
		return
	}
}
