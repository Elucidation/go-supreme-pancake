// web server displays SVG image of:
// 2D grid with cirlces in it, with lines between those within a certain range
// where range is calculated using brute force and grid-based nearest neighbor

package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/ajstarks/svgo"
)

const defaultstyle = "fill:rgb(127,0,0)"

var port = flag.String("port", ":2003", "http service address")

const WindowWidth = 500
const WindowHeight = 500

var N = 100 // number of bodies
var bodies [][]float64
var positions []float64

func main() {
	rand.Seed(time.Now().Unix())

	flag.Parse()
	http.Handle("/brute/", http.HandlerFunc(brute))
	http.Handle("/nn/", http.HandlerFunc(nn))
	fmt.Printf("Starting Server on http://localhost%s\n", *port)
	fmt.Printf("  Try going to http://localhost%s/brute\n", *port)
	err := http.ListenAndServe(*port, nil)
	if err != nil {
		log.Println("ListenAndServe:", err)
	}
}

// For now, units in pixels
func initSystem() {
	bodies = make([][]float64, N)
	positions = make([]float64, 3*N)
	// x y radius

	// body radius (in pixels)
	minradius := 10
	maxradius := 20

	for i := 0; i < N; i++ {
		idx := i * 3
		positions[idx] = rand.Float64() * WindowWidth
		positions[idx+1] = rand.Float64() * WindowHeight
		positions[idx+2] = float64(minradius) + rand.Float64()*float64(maxradius-minradius)

		bodies[i] = positions[idx : idx+3]

		// fmt.Println("#", i, ":", bodies[i])
	}
}

func shapestyle(path string) string {
	i := strings.LastIndex(path, "/") + 1
	if i > 0 && len(path[i:]) > 0 {
		return "fill:" + path[i:]
	}
	return defaultstyle
}

func getR2(a []float64, b []float64) float64 {
	return math.Pow(b[0]-a[0], 2) + math.Pow(b[1]-a[1], 2)
}

// O(N^2) number of checks needed
func bruteNearest() []int {
	// indices of nearest neighbors for each body
	nearest := make([]int, N)
	dists := make([]float64, N)

	// For each body
	for i := 0; i < N; i++ {
		dists[i] = WindowWidth * WindowHeight // set maxdist

		// check every other body, and set index to closest
		for j := 0; j < N; j++ {
			if i != j {
				r2 := getR2(bodies[i], bodies[j])
				if r2 < dists[i] {
					dists[i] = r2
					nearest[i] = j
				}
			}
		}
	}
	return nearest
}

func drawSystem(s *svg.SVG, style string) {
	for i := 0; i < N; i++ {
		s.Circle(
			int(bodies[i][0]),
			int(bodies[i][1]),
			int(bodies[i][2]),
			style)
	}
}

func brute(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)

	s.Start(WindowWidth, WindowHeight)
	s.Title("Nearest Neighbor Brute Force")

	initSystem()

	cellsize := WindowWidth / 10

	// Plot circles
	fmt.Println("Plotting brute force")
	drawSystem(s, shapestyle(req.URL.Path))

	// Get nearest neighbors using brute force
	nearest := bruteNearest()

	// Draw green line between nearest neighbors
	for i := 0; i < N; i++ {
		a := bodies[i]
		b := bodies[nearest[i]]
		s.Line(int(a[0]), int(a[1]),
			int(b[0]), int(b[1]),
			"stroke:green;stroke-width:3px")

	}

	// Draw grid last
	s.Grid(0, 0, WindowWidth, WindowHeight, cellsize, "stroke:rgba(100,100,100,0.5)")

	s.End()
}

func nn(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)
	s.Start(WindowWidth, WindowHeight)
	s.Title("Nearest Neighbor Grid-based")
	s.Circle(250, 250, 125, shapestyle(req.URL.Path))
	s.End()
}
