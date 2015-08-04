package geospatialfilestesting

import (
	"fmt"
	"gospatial/geospatialfiles/raster"
	"testing"
)

var println = fmt.Println
var printf = fmt.Printf
var sprintf = fmt.Sprintf

var testIdrisiRead = true

func TestIdrisiRead(t *testing.T) {
	if testIdrisiRead {
		inFile := "./testdata/DEM.rst"
		rin, err := raster.CreateRasterFromFile(inFile)
		if err != nil {
			t.Error("Failed to read file")
		}

		rows := rin.Rows
		columns := rin.Columns

		println(rows, columns)

	} else {
		t.SkipNow()
	}
}
