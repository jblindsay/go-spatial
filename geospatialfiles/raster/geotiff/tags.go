package geotiff

import "fmt"

type GeoTiffTag struct {
	Name string
	Code int
}

func (g GeoTiffTag) String() string {
	return fmt.Sprintf("Name: %s, Code: %d", g.Name, g.Code)
}

// Tags (see p. 28-41 of the spec).
var tagMap = map[int]GeoTiffTag{
	256: GeoTiffTag{"ImageWidth", 256},
	257: GeoTiffTag{"ImageLength", 257},
	258: GeoTiffTag{"BitsPerSample", 258},
	259: GeoTiffTag{"Compression", 259},
	262: GeoTiffTag{"PhotometricInterpretation", 262},
	266: GeoTiffTag{"FillOrder", 266},
	269: GeoTiffTag{"DocumentName", 269},
	284: GeoTiffTag{"PlanarConfiguration", 284},
	270: GeoTiffTag{"ImageDescription", 270},
	271: GeoTiffTag{"Make", 271},
	272: GeoTiffTag{"Model", 272},
	273: GeoTiffTag{"StripOffsets", 273},
	274: GeoTiffTag{"Orientation", 274},
	277: GeoTiffTag{"SamplesPerPixel", 277},
	278: GeoTiffTag{"RowsPerStrip", 278},
	279: GeoTiffTag{"StripByteCounts", 279},

	282: GeoTiffTag{"XResolution", 282},
	283: GeoTiffTag{"YResolution", 283},
	296: GeoTiffTag{"ResolutionUnit", 296},

	305: GeoTiffTag{"Software", 305},
	306: GeoTiffTag{"DateTime", 306},

	322: GeoTiffTag{"TileWidth", 322},
	323: GeoTiffTag{"TileLength", 323},
	324: GeoTiffTag{"TileOffsets", 324},
	325: GeoTiffTag{"TileByteCounts", 325},

	317: GeoTiffTag{"Predictor", 317},
	320: GeoTiffTag{"ColorMap", 320},
	338: GeoTiffTag{"ExtraSamples", 338},
	339: GeoTiffTag{"SampleFormat", 339},

	34735: GeoTiffTag{"GeoKeyDirectoryTag", 34735},
	34736: GeoTiffTag{"GeoDoubleParamsTag", 34736},
	34737: GeoTiffTag{"GeoAsciiParamsTag", 34737},
	33550: GeoTiffTag{"ModelPixelScaleTag", 33550},
	33922: GeoTiffTag{"ModelTiepointTag", 33922},
	34264: GeoTiffTag{"ModelTransformationTag", 34264},
	42112: GeoTiffTag{"GDAL_METADATA", 42112},
	42113: GeoTiffTag{"GDAL_NODATA", 42113},

	1024:  GeoTiffTag{"GTModelTypeGeoKey", 1024},
	1025:  GeoTiffTag{"GTRasterTypeGeoKey", 1025},
	1026:  GeoTiffTag{"GTCitationGeoKey", 1026},
	2048:  GeoTiffTag{"GeographicTypeGeoKey", 2048},
	2049:  GeoTiffTag{"GeogCitationGeoKey", 2049},
	2050:  GeoTiffTag{"GeogGeodeticDatumGeoKey", 2050},
	2051:  GeoTiffTag{"GeogPrimeMeridianGeoKey", 2051},
	2061:  GeoTiffTag{"GeogPrimeMeridianLongGeoKey", 2061},
	2052:  GeoTiffTag{"GeogLinearUnitsGeoKey", 2052},
	2053:  GeoTiffTag{"GeogLinearUnitSizeGeoKey", 2053},
	2054:  GeoTiffTag{"GeogAngularUnitsGeoKey", 2054},
	2055:  GeoTiffTag{"GeogAngularUnitSizeGeoKey", 2055},
	2056:  GeoTiffTag{"GeogEllipsoidGeoKey", 2056},
	2057:  GeoTiffTag{"GeogSemiMajorAxisGeoKey", 2057},
	2058:  GeoTiffTag{"GeogSemiMinorAxisGeoKey", 2058},
	2059:  GeoTiffTag{"GeogInvFlatteningGeoKey", 2059},
	2060:  GeoTiffTag{"GeogAzimuthUnitsGeoKey", 2060},
	3072:  GeoTiffTag{"ProjectedCSTypeGeoKey", 3072},
	3073:  GeoTiffTag{"PCSCitationGeoKey", 3073},
	3074:  GeoTiffTag{"ProjectionGeoKey", 3074},
	3075:  GeoTiffTag{"ProjCoordTransGeoKey", 3075},
	3076:  GeoTiffTag{"ProjLinearUnitsGeoKey", 3076},
	3077:  GeoTiffTag{"ProjLinearUnitSizeGeoKey", 3077},
	3078:  GeoTiffTag{"ProjStdParallel1GeoKey", 3078},
	3079:  GeoTiffTag{"ProjStdParallel2GeoKey", 3079},
	3080:  GeoTiffTag{"ProjNatOriginLongGeoKey", 3080},
	3081:  GeoTiffTag{"ProjNatOriginLatGeoKey", 3081},
	3082:  GeoTiffTag{"ProjFalseEastingGeoKey", 3082},
	3083:  GeoTiffTag{"ProjFalseNorthingGeoKey", 3083},
	3084:  GeoTiffTag{"ProjFalseOriginLongGeoKey", 3084},
	3085:  GeoTiffTag{"ProjFalseOriginLatGeoKey", 3085},
	3086:  GeoTiffTag{"ProjFalseOriginEastingGeoKey", 3086},
	3087:  GeoTiffTag{"ProjFalseOriginNorthingGeoKey", 3087},
	3088:  GeoTiffTag{"ProjCenterLongGeoKey", 3088},
	3089:  GeoTiffTag{"ProjCenterLatGeoKey", 3089},
	3090:  GeoTiffTag{"ProjCenterEastingGeoKey", 3090},
	3091:  GeoTiffTag{"ProjFalseOriginNorthingGeoKey", 3091},
	3092:  GeoTiffTag{"ProjScaleAtNatOriginGeoKey", 3092},
	3093:  GeoTiffTag{"ProjScaleAtCenterGeoKey", 3093},
	3094:  GeoTiffTag{"ProjAzimuthAngleGeoKey", 3094},
	3095:  GeoTiffTag{"ProjStraightVertPoleLongGeoKey", 3095},
	4096:  GeoTiffTag{"VerticalCSTypeGeoKey", 4096},
	4097:  GeoTiffTag{"VerticalCitationGeoKey", 4097},
	4098:  GeoTiffTag{"VerticalDatumGeoKey", 4098},
	4099:  GeoTiffTag{"VerticalUnitsGeoKey", 4099},
	50844: GeoTiffTag{"RPCCoefficientTag", 50844},
}
