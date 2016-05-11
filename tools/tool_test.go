package tools

import "testing"

var testFD8FA = false
var testDevFromMean = false
var testDevFromMeanTraditional = false
var testBreachStreams = false
var testWhiteboxRaster2GeoTiff = true

func TestFD8FA(t *testing.T) {
	if testFD8FA {
		fd8 := FD8FlowAccum{}
		args := make([]string, 4)
		args[0] = "/Users/johnlindsay/Documents/Research/FastBreaching/data/SRTM1GL/tmp4.dep"
		args[1] = "/Users/johnlindsay/Documents/Research/FastBreaching/data/SRTM1GL/tmp5.dep"
		//args[0] = "/Users/johnlindsay/Documents/Research/FastBreaching/data/breached.dep"
		//args[1] = "/Users/johnlindsay/Documents/Research/FastBreaching/data/tmp1.dep"

		args[2] = "true"
		args[3] = "true"
		fd8.ParseArguments(args)
	} else {
		t.SkipNow()
	}
}

func TestDevFromMean(t *testing.T) {
	if testDevFromMean {
		dfm := DeviationFromMean{}
		args := make([]string, 3)
		args[0] = "/Users/johnlindsay/Documents/Research/Multi-scale Topographic Position paper/data/Appalachians/UTM/DEM large lakes removed.dep"
		args[1] = "/Users/johnlindsay/Documents/Research/Multi-scale Topographic Position paper/data/Appalachians/UTM/tmp20.dep"
		args[2] = "18"
		dfm.ParseArguments(args)
	} else {
		t.SkipNow()
	}
}

func TestDevFromMeanTraditional(t *testing.T) {
	if testDevFromMeanTraditional {
		dfmt := DeviationFromMeanTraditional{}
		args := make([]string, 3)
		args[0] = "/Users/johnlindsay/Documents/Research/Multi-scale Topographic Position paper/data/Appalachians/UTM/DEM large lakes removed.dep"
		args[1] = "/Users/johnlindsay/Documents/Research/Multi-scale Topographic Position paper/data/Appalachians/UTM/tmp21.dep"
		args[2] = "18"
		dfmt.ParseArguments(args)
	} else {
		t.SkipNow()
	}
}

func TestBreachStreams(t *testing.T) {
	if testBreachStreams {
		bs := BreachStreams{}
		args := make([]string, 3)
		args[0] = "/Users/johnlindsay/Documents/Data/SouthernOnt/streams.dep"
		args[1] = "/Users/johnlindsay/Documents/Data/SouthernOnt/DEM_erased.dep"
		args[2] = "/Users/johnlindsay/Documents/Data/SouthernOnt/tmp8.dep"

		bs.ParseArguments(args)
	} else {
		t.SkipNow()
	}
}

func TestWhiteboxRaster2GeoTiff(t *testing.T) {
	if testWhiteboxRaster2GeoTiff {
		w2g := Whitebox2GeoTiff{}
		args := make([]string, 2)
		args[0] = "/Users/johnlindsay/Documents/Data/SouthernOnt/colour comp.dep"
		args[1] = "/Users/johnlindsay/Documents/Data/SouthernOnt/deleteMe2.tif"

		w2g.ParseArguments(args)
		println("Hello goofball")
	} else {
		t.SkipNow()
	}
}
