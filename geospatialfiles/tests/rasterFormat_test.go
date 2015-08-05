package tests

import (
	. "fmt"
	"os"
	"testing"

	"github.com/jblindsay/go-spatial/geospatialfiles/raster"
)

//var println = fmt.Println
//var printf = fmt.Printf
//var sprintf = fmt.Sprintf

var testIdrisiRead = true
var testIdrisiWrite = true
var testWhiteboxRead = true
var testGeoTiffRead = true

func TestIdrisiRead(t *testing.T) {
	if testIdrisiRead {
		inFile := "./testdata/DEM.rst"
		rin, err := raster.CreateRasterFromFile(inFile)
		if err != nil {
			t.Error("Failed to read file")
		}

		Println(rin.GetRasterConfig().String())

		if rin.Value(100, 100) != 429.42730712890625 {
			t.Fail()
		} else {
			Println("cell (100, 100) =", rin.Value(100, 100))
		}

	} else {
		t.SkipNow()
	}
}

func TestIdrisiWrite(t *testing.T) {
	if testIdrisiWrite {
		// read in an existing Whitebox file and output an Idrisi file
		inFile := "./testdata/DEM.dep"
		rin, err := raster.CreateRasterFromFile(inFile)
		if err != nil {
			t.Error("Failed to read file")
		}
		inConfig := rin.GetRasterConfig()
		rows := rin.Rows
		columns := rin.Columns

		config := raster.NewDefaultRasterConfig()
		config.DataType = raster.DT_FLOAT32
		config.InitialValue = 0.0
		config.CoordinateRefSystemWKT = inConfig.CoordinateRefSystemWKT
		config.EPSGCode = inConfig.EPSGCode
		outFile := "./testdata/DeleteMe.rst"
		rout, err := raster.CreateNewRaster(outFile, rows, columns,
			rin.North, rin.South, rin.East, rin.West, config)
		if err != nil {
			println("Failed to write raster")
			return
		}

		var row, column int
		var z float64
		for row = 0; row < rows; row++ {
			for column = 0; column < columns; column++ {
				z = rin.Value(row, column)
				rout.SetValue(row, column, z)
			}
		}

		rout.Save()

		rin, err = raster.CreateRasterFromFile(outFile)
		if err != nil {
			t.Error("Failed to read file")
		}

		Println(rin.GetRasterConfig().String())

		if rin.Value(100, 100) != 429.42730712890625 {
			t.Fail()
		} else {
			Println("cell (100, 100) =", rin.Value(100, 100))
		}

		// now clean up
		if _, err = os.Stat("./testdata/DeleteMe.rst"); err == nil {
			if err = os.Remove("./testdata/DeleteMe.rst"); err != nil {
				panic(err)
				t.Fail()
			}
		}
		if _, err = os.Stat("./testdata/DeleteMe.rdc"); err == nil {
			if err = os.Remove("./testdata/DeleteMe.rdc"); err != nil {
				panic(err)
				t.Fail()
			}
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
		Println(rin.GetRasterConfig().String())

		if rin.Value(100, 100) != 429.42730712890625 {
			t.Fail()
		} else {
			Println("cell (100, 100) =", rin.Value(100, 100))
		}

	} else {
		t.SkipNow()
	}
}

func TestGeoTiffRead(t *testing.T) {
	if testGeoTiffRead {
		//inFile := "./testdata/Sample64Bit.tif"
		inFile := "./testdata/DEM.tif"
		rin, err := raster.CreateRasterFromFile(inFile)
		if err != nil {
			t.Error("Failed to read file")
		}
		//println(rin.GetRasterConfig().String())

		tagInfo := rin.GetMetadataEntries()
		if len(tagInfo) > 0 {
			Println(tagInfo[0])
		} else {
			Println("Error reading metadata entries.")
			t.Fail()
		}

		//if rin.Value(100, 100) != 0.16925102519989013 {
		if rin.Value(100, 100) != 429.42730712890625 {
			t.Fail()
		} else {
			Println("cell (100, 100) =", rin.Value(100, 100))
		}

	} else {
		t.SkipNow()
	}
}
