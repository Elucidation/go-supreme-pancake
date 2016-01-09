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

const WindowWidth = 500
const WindowHeight = 400

var N = 50                      // number of bodies
var grid_cells = [2]int{15, 15} // number of cells in grid
var grid_data = [][][]int{}     // 2d grid with array of indices of bodies overlapping cell

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
	maxradius := math.Max(cellsize[0], cellsize[1]) / 2
	minradius := maxradius / 2

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
	count := 0

	// For each body
	for i := 0; i < N; i++ {
		dists[i] = WindowWidth * WindowHeight // set maxdist

		// check every other body, and set index to closest
		for j := 0; j < N; j++ {
			if i != j {
				count++
				r2 := getR2(bodies[i], bodies[j])
				if r2 < dists[i] {
					dists[i] = r2
					nearest[i] = j
				}
			}
		}
	}
	fmt.Println("Brute count:", count)
	return nearest
}

func clamp(val int, minval int, maxval int) int {
	if val < minval {
		return minval
	} else if val > maxval {
		return maxval
	}
	return val
}

func imin(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
func imax(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// O(n*d) number of checks for uniform density
func gridNearest() []int {
	// indices of nearest neighbors for each body
	nearest := make([]int, N)
	dists := make([]float64, N)
	count := 0

	// Initialize dists to max possible value
	for i := 0; i < N; i++ {
		nearest[i] = -1
		dists[i] = WindowWidth * WindowHeight // set maxdist
	}

	// For each grid cell
	for i := 0; i < grid_cells[0]; i++ {
		for j := 0; j < grid_cells[1]; j++ {

			// Get neighbors
			ra := imax(0, i-1)
			rb := imin(i+1, grid_cells[0]-1) + 1
			ca := imax(0, j-1)
			cb := imin(j+1, grid_cells[1]-1) + 1 // range include

			// For each body in this cell
			for _, bodyidx := range grid_data[i][j] {
				// Golang slices don't work like matlab slices, so no slice columns
				// http://stackoverflow.com/questions/15618155/go-lang-slice-columns-from-2d-array
				// fmt.Printf("\nIterating [%d:%d][%d:%d]\n", ra, rb, ca, cb)
				// For each neighbor in cell and adjacent cells
				for _, r := range grid_data[ra:rb] {
					for _, c := range r[ca:cb] {
						for _, otheridx := range c {
							// If not same body
							if bodyidx != otheridx {
								// Update if distance is less than last best
								count++
								r2 := getR2(bodies[bodyidx], bodies[otheridx])
								if r2 < dists[bodyidx] {
									dists[bodyidx] = r2
									nearest[bodyidx] = otheridx
								}
							}
						}
					}
				}
			}
		}
	}

	// Check those that aren't within a cell of others
	// default to brute force check for these
	fmt.Println(nearest)
	fmt.Println(dists)
	for i, v := range nearest {
		if v == -1 {
			// check every other body, and set index to closest
			for j := 0; j < N; j++ {
				if i != j {
					count++
					r2 := getR2(bodies[i], bodies[j])
					if r2 < dists[i] {
						dists[i] = r2
						nearest[i] = j
					}
				}
			}
		}
	}

	fmt.Println("Grid count:", count)
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
func Grid2(s *svg.SVG, x float64, y float64, w float64, h float64, nx float64, ny float64, style ...string) {

	if len(style) > 0 {
		s.Gstyle(style[0])
	}
	for ix := x; ix <= x+w; ix += nx {
		s.Line(int(ix), int(y), int(ix), int(y+h))
	}

	for iy := y; iy <= y+h; iy += ny {
		s.Line(int(x), int(iy), int(x+w), int(iy))
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

	// Draw system
	fmt.Println("Drawing system of", N, "bodies")
	drawSystem(s, shapestyle(req.URL.Path))

	// Get nearest neighbors using brute force
	nearestBrute := bruteNearest()

	calcCells()
	nearestGrid := gridNearest()

	mismatch_count := 0
	for i, v := range nearestBrute {
		if v != nearestGrid[i] {
			mismatch_count++
			fmt.Println(mismatch_count, "MISMATCH", i, v, nearestGrid[i])
		}
	}
	if mismatch_count > 0 {
		fmt.Println("Warning, brute force & grid-based have mismatch")
		// fmt.Println("brute", nearestBrute)
		// fmt.Println("grid", nearestGrid)
	}
	nearest := nearestGrid

	// Draw green line between nearest neighbors
	for i := 0; i < N; i++ {
		if nearestGrid[i] != nearestBrute[i] {
			// draw mismatch line
			a := bodies[i]
			b := bodies[nearestBrute[i]]
			c := bodies[nearestGrid[i]]
			s.Line(int(a[0]), int(a[1]),
				int(b[0]), int(b[1]),
				"stroke:red;stroke-width:5px")
			s.Line(int(a[0]), int(a[1]),
				int(c[0]), int(c[1]),
				"stroke:orange;stroke-width:5px")
		}
		if nearest[i] >= 0 {
			a := bodies[i]
			b := bodies[nearest[i]]
			s.Line(int(a[0]), int(a[1]),
				int(b[0]), int(b[1]),
				"stroke:green;stroke-width:3px")
		}

	}

	// Draw grid last
	drawCells(s)
	Grid2(s, 0, 0, WindowWidth, WindowHeight, cellsize[0], cellsize[1], "stroke:rgba(100,100,100,0.5)")

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
