// MIT License
//
// Copyright (c) 2021 Shyam Jesalpura and EASE lab
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package plotter

import (
	"fmt"
	"image/color"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const (
	plotLocationPrefix = "./"
)

func PlotLatenciesCDF(sortedLatencies []float64, payloadSize int) {
	plotInstance := plot.New()

	plotInstance.Title.Text = fmt.Sprintf("Latency CDF for %dKiB requests", payloadSize)
	plotInstance.Y.Label.Text = "Portion of requests"
	plotInstance.X.Label.Text = "Latency (ms)"

	latenciesToPlot := make(plotter.XYs, len(sortedLatencies))
	for i := 0; i < len(sortedLatencies); i++ {
		latenciesToPlot[i].X = sortedLatencies[i] / 1000.0
		latenciesToPlot[i].Y = stat.CDF(
			sortedLatencies[i],
			stat.Empirical,
			sortedLatencies,
			nil,
		)
	}

	err := plotutil.AddLinePoints(plotInstance, latenciesToPlot)
	if err != nil {
		log.Errorf("[sub-experiment %dKiB] Could not add line points to CDF plot: %s", payloadSize, err.Error())
	}

	// Save the plot to a PNG file.
	if err := plotInstance.Save(5*vg.Inch, 5*vg.Inch, plotLocationPrefix+"cdf_"+strconv.Itoa(payloadSize)+"KiB.png"); err != nil {
		log.Errorf("[sub-experiment %dKiB] Could not save CDF plot: %s", payloadSize, err.Error())
	}
}

func PlotPercentile(sortedLatencyMap map[int][]float64) {
	plotInstance := plot.New()

	plotInstance.Title.Text = "Latency (50th percentile) as a function of payload size"
	plotInstance.Y.Label.Text = "Latency (ms)"
	plotInstance.Y.Scale = plot.LogScale{}
	plotInstance.Y.Tick.Marker = plot.LogTicks{}
	plotInstance.X.Label.Text = "Size of payload (KiB)"
	plotInstance.X.Scale = plot.LogScale{}
	plotInstance.X.Tick.Marker = plot.LogTicks{}

	percentiles := []int{50}

	for _, percentile := range percentiles {
		var pointLabels []string
		latenciesToPlot := make(plotter.XYs, len(sortedLatencyMap))
		i := 0
		for payloadSize, sortedLatency := range sortedLatencyMap {
			latenciesToPlot[i].X = float64(payloadSize)
			latenciesToPlot[i].Y = sortedLatency[percentile*len(sortedLatency)/100] / 1000.0
			pointLabels = append(pointLabels, fmt.Sprintf("%.2f", latenciesToPlot[i].Y))
			i += 1
		}

		labels, err := plotter.NewLabels(plotter.XYLabels{
			XYs:    latenciesToPlot,
			Labels: pointLabels,
		},
		)
		if err != nil {
			log.Fatalf("could not creates labels plotter: %+v", err)
		}

		err = plotutil.AddLinePoints(plotInstance, strconv.Itoa(percentile)+" Percentile", latenciesToPlot)
		if err != nil {
			log.Errorf("Could not add line points to CDF plot: %s", err.Error())
		}

		plotInstance.Add(labels)

	}

	// Save the plot to a PNG file.
	if err := plotInstance.Save(5*vg.Inch, 5*vg.Inch, plotLocationPrefix+"percentile_plot.png"); err != nil {
		log.Errorf("Could not save percentile plot: %s", err.Error())
	}
}

func PlotBW(sortedLatencyMap map[int][]float64) {
	plotInstance := plot.New()

	plotInstance.Title.Text = "Bandwidth as a function of a payload size"
	plotInstance.Y.Label.Text = "BW in MiB/s"
	plotInstance.X.Label.Text = "Size of payload (KiB)"
	plotInstance.X.Scale = plot.LogScale{}
	plotInstance.X.Tick.Marker = plot.LogTicks{}

	latenciesToPlot := make(plotter.XYs, len(sortedLatencyMap))
	var pointLabels []string
	i := 0

	for payloadSize, sortedLatency := range sortedLatencyMap {
		latenciesToPlot[i].X = float64(payloadSize)
		latenciesToPlot[i].Y = float64(payloadSize) * 1024 / (sortedLatency[50*len(sortedLatency)/100])
		pointLabels = append(pointLabels, fmt.Sprintf("%.2f", latenciesToPlot[i].Y))
		i += 1
	}

	labels, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    latenciesToPlot,
		Labels: pointLabels,
	},
	)
	if err != nil {
		log.Fatalf("could not creates labels plotter: %+v", err)
	}

	lpLine, lpPoints, err := plotter.NewLinePoints(latenciesToPlot)
	if err != nil {
		panic(err)
	}
	lpLine.Color = color.RGBA{G: 255, A: 255}
	lpPoints.Shape = draw.PyramidGlyph{}
	lpPoints.Color = color.RGBA{R: 255, A: 255}

	// Add the plotters to the plot, with a legend
	// entry for each
	plotInstance.Add(lpLine, lpPoints, labels)

	// Save the plot to a PNG file.
	if err := plotInstance.Save(5*vg.Inch, 5*vg.Inch, plotLocationPrefix+"BW_plot.png"); err != nil {
		log.Errorf("Could not save BW plot: %s", err.Error())
	}
}
