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

func PlotLatenciesCDF(plotPath string, sortedLatencies []float64, payloadSize int) {
	plotInstance := plot.New()

	plotInstance.Title.Text = fmt.Sprintf("Latency CDF for %dKiB requests", payloadSize)
	plotInstance.Y.Label.Text = "Portion of requests"
	plotInstance.X.Label.Text = "Latency (ms)"

	// Uncomment below for hard X limit
	//var maxIndexKept int
	//for maxIndexKept = 0; maxIndexKept < len(sortedLatencies) && sortedLatencies[maxIndexKept] <= plotInstance.X.Max; maxIndexKept++ {
	//}
	//sortedLatencies = sortedLatencies[:maxIndexKept]

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
	if err := plotInstance.Save(5*vg.Inch, 5*vg.Inch, plotPath); err != nil {
		log.Errorf("[sub-experiment %dKiB] Could not save CDF plot: %s", payloadSize, err.Error())
	}
}

func PlotPercentile(sortedLatencyMap map[int][]float64) {
	plotInstance := plot.New()

	plotInstance.Title.Text = fmt.Sprintf("Latency (50th percentile) as a function of payload size")
	plotInstance.Y.Label.Text = "Latency (ms)"
	plotInstance.Y.Scale = plot.LogScale{}
	plotInstance.Y.Tick.Marker = plot.LogTicks{}
	plotInstance.X.Label.Text = "Size of payload (KiB)"
	plotInstance.X.Scale = plot.LogScale{}
	plotInstance.X.Tick.Marker = plot.LogTicks{}

	// Uncomment below for hard X limit
	//var maxIndexKept int
	//for maxIndexKept = 0; maxIndexKept < len(sortedLatencies) && sortedLatencies[maxIndexKept] <= plotInstance.X.Max; maxIndexKept++ {
	//}
	//sortedLatencies = sortedLatencies[:maxIndexKept]

	percentiles := []int{50}

	for _, percentile := range percentiles {
		point_labels := []string{}
		latenciesToPlot := make(plotter.XYs, len(sortedLatencyMap))
		i := 0
		for payloadSize, sorted_latency := range sortedLatencyMap {
			latenciesToPlot[i].X = float64(payloadSize)
			// log.Printf("inserting %dth of sorted, which is %f", int(percentile*len(sorted_latency)/100), sorted_latency[int(percentile*len(sorted_latency)/100)])
			latenciesToPlot[i].Y = sorted_latency[int(percentile*len(sorted_latency)/100)] / 1000.0
			point_labels = append(point_labels, fmt.Sprintf("%.2f", latenciesToPlot[i].Y))
			i += 1
		}

		labels, err := plotter.NewLabels(plotter.XYLabels{
			XYs:    latenciesToPlot,
			Labels: point_labels,
		},
		)
		if err != nil {
			log.Fatalf("could not creates labels plotter: %+v", err)
		}

		err = plotutil.AddLinePoints(plotInstance, strconv.Itoa(percentile)+" Percentile", latenciesToPlot)
		if err != nil {
			log.Errorf("Could not add line points to CDF plot: %s", err.Error())
		}

		// lpLine, lpPoints, err := plotter.NewLinePoints(latenciesToPlot)
		// if err != nil {
		// 	panic(err)
		// }
		// lpLine.Color = color.RGBA{G: uint8(percentile * 255 / 100), A: 255}
		// lpPoints.Shape = draw.PyramidGlyph{}
		// lpPoints.Color = color.RGBA{R: 255, A: 255}

		// Add the plotters to the plot, with a legend
		// entry for each
		// plotInstance.Add(lpLine, lpPoints)
		plotInstance.Add(labels)

		// plotInstance.Legend.Add(strconv.Itoa(percentile)+" Percentile", lpLine, lpPoints)
	}

	// Save the plot to a PNG file.
	if err := plotInstance.Save(5*vg.Inch, 5*vg.Inch, "percentile_plot.png"); err != nil {
		log.Errorf("Could not save percentile plot: %s", err.Error())
	}
}

func PlotBW(sortedLatencyMap map[int][]float64) {
	plotInstance := plot.New()

	plotInstance.Title.Text = fmt.Sprintf("Bandwidth as a function of a payload size")
	plotInstance.Y.Label.Text = "BW in MiB/s"
	plotInstance.X.Label.Text = "Size of payload (KiB)"
	plotInstance.X.Scale = plot.LogScale{}
	plotInstance.X.Tick.Marker = plot.LogTicks{}

	// Uncomment below for hard X limit
	//var maxIndexKept int
	//for maxIndexKept = 0; maxIndexKept < len(sortedLatencies) && sortedLatencies[maxIndexKept] <= plotInstance.X.Max; maxIndexKept++ {
	//}
	//sortedLatencies = sortedLatencies[:maxIndexKept]

	latenciesToPlot := make(plotter.XYs, len(sortedLatencyMap))
	point_labels := []string{}
	i := 0

	for payloadSize, sorted_latency := range sortedLatencyMap {
		latenciesToPlot[i].X = float64(payloadSize)
		// log.Printf("inserting %dth of sorted, which is %f", int(percentile*len(sorted_latency)/100), sorted_latency[int(percentile*len(sorted_latency)/100)])
		latenciesToPlot[i].Y = float64(payloadSize) * 1024 / (sorted_latency[int(50*len(sorted_latency)/100)])
		point_labels = append(point_labels, fmt.Sprintf("%.2f", latenciesToPlot[i].Y))
		i += 1
	}

	labels, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    latenciesToPlot,
		Labels: point_labels,
	},
	)
	if err != nil {
		log.Fatalf("could not creates labels plotter: %+v", err)
	}

	// err := plotutil.AddLinePoints(plotInstance, strconv.Itoa(percentile)+" Percentile", latenciesToPlot)
	// if err != nil {
	// 	log.Errorf("[sub-experiment %dMB] Could not add line points to CDF plot: %s", percentile, err.Error())
	// }

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
	if err := plotInstance.Save(5*vg.Inch, 5*vg.Inch, "BW_plot.png"); err != nil {
		log.Errorf("Could not save BW plot: %s", err.Error())
	}
}
