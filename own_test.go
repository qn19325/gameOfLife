package main

import (
    "fmt"
    "os"
    "testing"
 
    "uk.ac.bris.cs/gameoflife/gol"
)
 
func BenchmarkDistributor(b *testing.B){
	params := []gol.Params{
		{ImageWidth: 16, ImageHeight: 16},
		{ImageWidth: 64, ImageHeight: 64},
		{ImageWidth: 512, ImageHeight: 512},
	}
	for _, p := range params {
		for _, turns := range []int{100,200,500} {
            p.Turns = turns
            os.Stdout = nil
			for threads := 1; threads <= 16; threads++ {
                p.Threads = threads
                testName := fmt.Sprintf("%dx%dx%dx%d benchmark", p.ImageHeight, p.ImageWidth, p.Turns, p.Threads)
                b.Run(testName, func(b *testing.B) {
                    for i := 0; i < b.N; i++ {
                        events := make(chan gol.Event)
                        gol.Run(p, events, nil)
                        
                    }
                })
            }
        }  
    }		
}
