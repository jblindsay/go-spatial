package tests

import (
	"fmt"
	"testing"

	"github.com/jblindsay/go-spatial/geospatialfiles/raster"
)

var println = fmt.Println
var printf = fmt.Printf
var sprintf = fmt.Sprintf

var testIdrisiRead = true
var testWhiteboxRead = true
var testGeoTiffRead = true

func TestIdrisiRead(t *testing.T) {
	if testIdrisiRead {
		inFile := "./testdata/DEM.rst"
		rin, err := raster.CreateRasterFromFile(inFile)
		if err != nil {
			t.Error("Failed to read file")
		}

		println(rin.GetRasterConfig().String())

		if rin.Value(100, 100) != 429.42730712890625 {
			t.Fail()
		} else {
			println("cell (100, 100) =", rin.Value(100, 100))
		}

	} else {
		t.SkipNow()
	}
}

func TestWhiteboxRead(t *testing.T) {
	if testWhiteboxRead {
		inFile := "./testdata/DEM.dep"
		rin, err := raster.CreateRasterFromFile(inFile)
		if err != nil {
			t.Error("Failed to read file")
		}
		println(rin.GetRasterConfig().String())

		if rin.Value(100, 100) != 429.42730712890625 {
			t.Fail()
		} else {
			println("cell (100, 100) =", rin.Value(100, 100))
		}

	} else {
		t.SkipNow()
	}
}

func TestGeoTiffRead(t *testing.T) {
	if testGeoTiffRead {
		inFile := "./testdata/land_shallow_topo_2048.tiff"
		rin, err := raster.CreateRasterFromFile(inFile)
		if err != nil {
			t.Error("Failed to read file")
		}
		println(rin.GetRasterConfig().String())

		//		if rin.Value(100, 100) != 429.42730712890625 {
		//			t.Fail()
		//		} else {
		//			println("cell (100, 100) =", rin.Value(100, 100))
		//		}

	} else {
		t.SkipNow()
	}
}
