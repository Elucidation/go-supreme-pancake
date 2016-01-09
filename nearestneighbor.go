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

// Globals
const defaultstyle = "fill:rgb(127,0,0)"

var port = flag.String("port", ":2003", "http service address")

const WindowWidth = 576
const WindowHeight = 350

var N = 50                    // number of bodies
var grid_cells = [2]int{9, 7} // number of cells in grid
var grid_data = [][][]int{}   // 2d grid with array of indices of bodies overlapping cell

var bodies [][]float64
var positions []float64

var cellsize = [2]float64{
	float64(WindowWidth) / float64(grid_cells[0]),
	float64(WindowHeight) / float64(grid_cells[1])}

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

	// Init grid data
	grid_data = make([][][]int, grid_cells[0])
	for i := 0; i < grid_cells[0]; i++ {
		grid_data[i] = make([][]int, grid_cells[1])
		for j := 0; j < grid_cells[1]; j++ {
			grid_data[i][j] = make([]int, 0)
		}
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

		var style_red = "fill:rgb(127,0,0)"
		var style_blue = "fill:rgb(0,0,127)"
		var cell = getCell(bodies[i][0], bodies[i][1])
		// fmt.Println(i, cell)
		var thestyle string
		if math.Mod(float64(cell[0]+cell[1]*grid_cells[0]), 2) == 0 {
			thestyle = style_red
		} else {
			thestyle = style_blue
		}

		s.Circle(
			int(bodies[i][0]),
			int(bodies[i][1]),
			int(bodies[i][2]),
			thestyle)
	}
}

// Grid draws a grid at the specified coordinate, dimensions, and spacing, with optional style.
func Grid2(s *svg.SVG, x int, y int, w int, h int, nx int, ny int, style ...string) {

	if len(style) > 0 {
		s.Gstyle(style[0])
	}
	for ix := x; ix <= x+w; ix += nx {
		s.Line(ix, y, ix, y+h)
	}

	for iy := y; iy <= y+h; iy += ny {
		s.Line(x, iy, x+w, iy)
	}
	if len(style) > 0 {
		s.Gend()
	}

}

func calcCells() {
	for i := 0; i < N; i++ {
		cell := getCell(bodies[i][0], bodies[i][1])
		// fmt.Println(i, cell, bodies[i])
		grid_data[cell[0]][cell[1]] = append(grid_data[cell[0]][cell[1]], i)
	}
}

func drawCells(s *svg.SVG) {
	for i := 0; i < grid_cells[0]; i++ {
		for j := 0; j < grid_cells[1]; j++ {
			// fmt.Println(i, j, grid_data[i][j])
			// TODO: draw rectangle in grid cell with color as # of objects
			p := getGridCoords(i, j)
			alpha := math.Min(1.0, float64(len(grid_data[i][j]))/10)
			thestyle := fmt.Sprintf("fill:rgba(0,0,0,%f)", alpha)
			s.Rect(p[0], p[1], int(cellsize[0]), int(cellsize[1]), thestyle)
		}
	}
}

func getCell(x float64, y float64) [2]int {
	return [2]int{
		int(x / cellsize[0]),
		int(y / cellsize[1])}
}

func getGridCoords(cx int, cy int) [2]int {
	return [2]int{
		int(float64(cx) * cellsize[0]),
		int(float64(cy) * cellsize[1]),
	}
}

func brute(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)

	s.Start(WindowWidth, WindowHeight)
	s.Title("Nearest Neighbor Brute Force")

	initSystem()

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
	calcCells()
	drawCells(s)
	Grid2(s, 0, 0, WindowWidth, WindowHeight, int(cellsize[0]), int(cellsize[1]), "stroke:rgba(100,100,100,0.5)")

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
