# go-supreme-pancake - Nearest Neighbor using svgo 
Go webserver serving SVG images demonstrating grid-based nearest neighbor calculation using [ajstarks/svgo](https://github.com/ajstarks/svgo) to generate SVG images.

Example with 100 bodies with green lines between nearest neighbors, An easy way to find the nearest neighbors is using brute force `O(n^2)` calculations by just checking every possible combination. We can do better than O(n^2) for nearest neighbor calculation. One way is to grid up the space into cells, in one `O(n)` pass populate the cells with the indices of bodies overlapping that cell.

This image shows [one example SVG](nearest_neighbors.svg) generated for 100 random bodies.

![100 bodies with green lines to nearest neighbors](http://rawgithub.com/Elucidation/go-supreme-pancake/master/nearest_neighbors.svg)

program output used when generating that image:

    Starting Server on http://localhost:2003
      Try going to http://localhost:2003/brute
    Warning, brute force & grid-based have mismatch
    Drawing system of 100 bodies
    Brute count: 9900
    Grid count: 733
    1 MISMATCH 56 66 64
    Warning, brute force & grid-based have mismatch

Notice how Brute force for `N=100` is exactly `O(N^2) = 100*100 - 100 = 9900` where `-100` is ignoring self checks for the 100 bodies. In comparison the grid based measurements is `733` which is a roughly of `N * average cell density + mismatch*N`.



The cell size is determined as the max radius of a body, meaning a body will never overlap more than 2 cells in any direction. 

Then, in another `O(n)` pass we check each body with the adjacent cells, roughly`n*d` calculations where d is the average density of the space. From this in the average and best case we approach `O(n)` calculations, with a worst case `O(n^2)` when all bodies are within 1 cell of each other, which would call for a better grid.

## Install

To install:

    $> go install github.com/Elucidation/go-supreme-pancake

## Run

To run, if you installed Go in the default directory just call `go-supreme-pancake`, otherwise run:
  
    $> $GOPATH/bin/go-supreme-pancake

    Starting Server on http://localhost:2003
      Try going to http://localhost:2003/brute


## Dependencies

    go get github.com/ajstarks/svgo