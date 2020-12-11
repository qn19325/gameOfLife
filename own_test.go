
package main
 
import (
    "fmt"
    "os"
    "testing"
 
    "uk.ac.bris.cs/gameoflife/gol"
)
 
func Benchmark(b *testing.B) {

    tests := []gol.Params{
        {ImageWidth: 16, ImageHeight: 16},
        {ImageWidth: 128, ImageHeight: 128},
        {ImageWidth: 512, ImageHeight: 512},
    }
    for _, p := range tests {
        for i:=0; i<3; i++{
            p.Threads = 16
            p.Turns = 100
            os.Stdout = nil
            testName := fmt.Sprintf("Testing Benchmark For: %dx%d Turns: %d  Threads: %d ", p.ImageHeight, p.ImageWidth, p.Turns, p.Threads)
            b.Run(testName, func(b *testing.B) {
                for i := 0; i < b.N; i++ {
                    events := make(chan gol.Event)
                    gol.Run(p, events, nil)
                    for range events{

                    }
                    
                }
            })
        }
    }
}