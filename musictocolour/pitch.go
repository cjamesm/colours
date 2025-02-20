package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/fogleman/gg"
	"github.com/gordonklaus/portaudio"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/mjibson/go-dsp/fft"
)

const sampleRate = 44100
const bufferSize = 2048 // Increased buffer size
const averagingPeriod = 500 * time.Millisecond

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	in := make([]float32, bufferSize)
	stream, err := portaudio.OpenDefaultStream(1, 0, sampleRate, len(in), in)
	if err != nil {
		log.Fatalf("Error opening default stream: %v", err)
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		log.Fatalf("Error starting stream: %v", err)
	}
	defer stream.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	fmt.Println("Recording... Press Ctrl+C to stop.")

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "index.html")
		})
		http.HandleFunc("/output/colour_box.png", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "output/colour_box.png")
		})
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	var pitches []float64
	var volumes []float64
	ticker := time.NewTicker(averagingPeriod)

	for {
		select {
		case <-sig:
			return
		case <-ticker.C:
			if len(pitches) > 0 {
				averagePitch := calculateAverage(pitches)
				averageVolume := calculateAverage(volumes)
				fmt.Printf("Average Pitch: %.2f Hz, Average Volume: %.2f\n", averagePitch, averageVolume)

				colour := pitchToColour(averagePitch, averageVolume)
				drawColourBox(colour)

				pitches = pitches[:0] // Reset pitches slice
				volumes = volumes[:0] // Reset volumes slice
			}
		default:
			if err := stream.Read(); err != nil {
				if paErr, ok := err.(portaudio.Error); ok && paErr == portaudio.InputOverflowed {
					log.Printf("Input overflowed: %v", err)
					continue
				}
				log.Printf("Error reading from stream: %v", err)
				continue
			}

			pitch := detectPitch(in)
			volume := calculateVolume(in)
			pitches = append(pitches, pitch)
			volumes = append(volumes, volume)

			time.Sleep(10 * time.Millisecond) // Reduced sleep duration
		}
	}
}

func detectPitch(samples []float32) float64 {
	complexSamples := make([]complex128, len(samples))
	for i, sample := range samples {
		complexSamples[i] = complex(float64(sample), 0)
	}

	fftResult := fft.FFT(complexSamples)
	magnitudes := make([]float64, len(fftResult))
	for i, c := range fftResult {
		magnitudes[i] = math.Sqrt(real(c)*real(c) + imag(c)*imag(c))
	}

	maxIndex := 0
	maxMagnitude := 0.0
	for i, magnitude := range magnitudes {
		if magnitude > maxMagnitude {
			maxMagnitude = magnitude
			maxIndex = i
		}
	}

	frequency := float64(maxIndex) * sampleRate / float64(len(samples))
	return frequency
}

func calculateVolume(samples []float32) float64 {
	sum := 0.0
	for _, sample := range samples {
		sum += math.Abs(float64(sample))
	}
	return sum / float64(len(samples))
}

func calculateAverage(values []float64) float64 {
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func pitchToColour(pitch, volume float64) colorful.Color {
	// Define the min and max pitch values for mapping
	const minPitch = 20.0     // Minimum pitch in Hz
	const maxPitch = 100000.0 // Maximum pitch in Hz

	// Clamp the pitch value to the min and max range
	clampedPitch := math.Max(minPitch, math.Min(maxPitch, pitch))

	// Map the clamped pitch to a value between 0 and 1
	normalizedPitch := (clampedPitch - minPitch) / (maxPitch - minPitch)

	// Map the normalized pitch to hue (0 to 360)
	hue := normalizedPitch * 360

	// Map the volume to saturation (0.5 to 1.0)
	const minVolume = 0.0
	const maxVolume = 1.0
	clampedVolume := math.Max(minVolume, math.Min(maxVolume, volume))
	saturation := 0.5 + (clampedVolume * 0.5)

	return colorful.Hsl(hue, saturation, 0.5)
}

func drawColourBox(colour colorful.Color) {
	const width = 200
	const height = 200

	dc := gg.NewContext(width, height)
	dc.SetColor(colour)
	dc.Clear()
	dc.SavePNG("output/colour_box.png")
}
