package geotiff

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strings"

	"gospatial/geospatialfiles/raster/geotiff/lzw"
)

var println = fmt.Println
var printf = fmt.Printf

// common errors
var FileWritingError = errors.New("An error occurred while writing the data file.")

type GeoTIFF struct {
	r          io.ReaderAt
	ifdList    map[int]IfdEntry
	geoKeyList map[int]IfdEntry
	ByteOrder  binary.ByteOrder
	NumGeoKeys int

	Rows              uint
	Columns           uint
	Data              []float64
	BitsPerSample     []uint
	samplesPerPixel   uint
	SampleFormat      uint
	PhotometricInterp uint
	mode              imageMode
	buf               []byte
	off               int // Current offset in buf.
	palette           []uint32
	TiepointData      TiepointTransformationParameters
	NodataValue       string
	RasterPixelIsArea bool
	EPSGCode          uint
}

func (g *GeoTIFF) Write(fileName string) (err error) {

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer f.Close()

	// create the buffered writer
	w := bufio.NewWriter(f)

	// Write the header.
	header := leHeader
	if g.ByteOrder == binary.BigEndian {
		header = beHeader
	}
	_, err = w.WriteString(header)
	if err != nil {
		return err
	}

	// output the offset to the IFD
	var totalBytesPerPixel uint32 = 0
	for _, bits := range g.BitsPerSample {
		totalBytesPerPixel += uint32(bits)
	}
	totalBytesPerPixel /= 8
	imageLen := uint32(g.Rows) * uint32(g.Columns) * totalBytesPerPixel
	if err = binary.Write(w, g.ByteOrder, imageLen+8); err != nil {
		return err
	}

	// output the data; compression is not currently supported for output
	g.samplesPerPixel = uint(len(g.BitsPerSample))
	buf := new(bytes.Buffer)
	switch g.PhotometricInterp {
	case PI_BlackIsZero, PI_WhiteIsZero:
		if g.samplesPerPixel != 1 {
			err = errors.New("The number of samples per pixel should be 1 for this photometric interpretation.")
			return err
		}
		switch g.SampleFormat {
		case SF_SignedInteger:
			switch g.BitsPerSample[0] {
			case 8:
				out := make([]int8, len(g.Data))
				for i := 0; i < len(g.Data); i++ {
					out[i] = int8(g.Data[i])
				}
				if err = binary.Write(buf, g.ByteOrder, out); err != nil {
					return FileWritingError
				}
				//for _, v := range g.Data {
				//	if err = binary.Write(buf, g.ByteOrder, int8(v)); err != nil {
				//		return FileWritingError
				//	}
				//}
			case 16:
				out := make([]int16, len(g.Data))
				for i := 0; i < len(g.Data); i++ {
					out[i] = int16(g.Data[i])
				}
				if err = binary.Write(buf, g.ByteOrder, out); err != nil {
					return FileWritingError
				}
				//for _, v := range g.Data {
				//	if err = binary.Write(buf, g.ByteOrder, int16(v)); err != nil {
				//		return FileWritingError
				//	}
				//}
			case 32:
				out := make([]int32, len(g.Data))
				for i := 0; i < len(g.Data); i++ {
					out[i] = int32(g.Data[i])
				}
				if err = binary.Write(buf, g.ByteOrder, out); err != nil {
					return FileWritingError
				}
			case 64:
				out := make([]int64, len(g.Data))
				for i := 0; i < len(g.Data); i++ {
					out[i] = int64(g.Data[i])
				}
				if err = binary.Write(buf, g.ByteOrder, out); err != nil {
					return FileWritingError
				}
			default:
				err = errors.New("Unexpected bit-depth.")
				return err
			}
		case SF_FloatingPoint:
			switch g.BitsPerSample[0] {
			case 32:
				out := make([]float32, len(g.Data))
				for i := 0; i < len(g.Data); i++ {
					out[i] = float32(g.Data[i])
				}
				if err = binary.Write(buf, g.ByteOrder, out); err != nil {
					return FileWritingError
				}
				//for _, v := range g.Data {
				//	if err = binary.Write(buf, g.ByteOrder, float32(v)); err != nil {
				//		return FileWritingError
				//	}
				//}
			case 64:
				if err = binary.Write(buf, g.ByteOrder, g.Data); err != nil {
					return FileWritingError
				}
			default:
				err = errors.New("Unexpected bit-depth.")
				return err
			}
		default: // sfUnsignedInteger
			switch g.BitsPerSample[0] {
			case 8:
				out := make([]uint8, len(g.Data))
				for i := 0; i < len(g.Data); i++ {
					out[i] = uint8(g.Data[i])
				}
				if err = binary.Write(buf, g.ByteOrder, out); err != nil {
					return FileWritingError
				}
				//for _, v := range g.Data {
				//	if err = binary.Write(buf, g.ByteOrder, uint8(v)); err != nil {
				//		return FileWritingError
				//	}
				//}
			case 16:
				out := make([]uint16, len(g.Data))
				for i := 0; i < len(g.Data); i++ {
					out[i] = uint16(g.Data[i])
				}
				if err = binary.Write(buf, g.ByteOrder, out); err != nil {
					return FileWritingError
				}
			case 32:
				out := make([]uint32, len(g.Data))
				for i := 0; i < len(g.Data); i++ {
					out[i] = uint32(g.Data[i])
				}
				if err = binary.Write(buf, g.ByteOrder, out); err != nil {
					return FileWritingError
				}
			case 64:
				out := make([]uint64, len(g.Data))
				for i := 0; i < len(g.Data); i++ {
					out[i] = uint64(g.Data[i])
				}
				if err = binary.Write(buf, g.ByteOrder, out); err != nil {
					return FileWritingError
				}
			default:
				err = errors.New("Unexpected bit-depth.")
				return err
			}
		}
		w.Write(buf.Bytes())
	case PI_RGB:
		i := 0
		bytes := make([]uint8, 3*len(g.Data))
		if g.samplesPerPixel == 3 {
			for _, v := range g.Data {
				val := uint32(v)
				red := uint8((val >> 16) & 0xFF)
				green := uint8((val >> 8) & 0xFF)
				blue := uint8(val & 0xFF)
				bytes[i] = red
				bytes[i+1] = green
				bytes[i+2] = blue
				i += 3
			}
		} else if g.samplesPerPixel == 4 { // RGBa
			for _, v := range g.Data {
				val := uint32(v)
				alpha := uint8((val >> 24) & 0xFF)
				red := uint8((val >> 16) & 0xFF)
				green := uint8((val >> 8) & 0xFF)
				blue := uint8(val & 0xFF)
				bytes[i] = red
				bytes[i+1] = green
				bytes[i+2] = blue
				bytes[i+3] = alpha
				i += 4
			}
		} else {
			err = errors.New("Unexpected number of samples per pixel.")
			return err
		}
		w.Write(bytes)
	case PI_Paletted:
		// TODO write the code for a paletted tiff
	default:
		panic(errors.New("An error has occurred during the writing of the geoTIFF file."))
	}

	// create the ifd's
	ifd := make([]IfdEntry, 0)
	ifd = append(ifd, CreateIfdEntry(tImageWidth, dtShort, 1, uint16(g.Columns), g.ByteOrder))
	ifd = append(ifd, CreateIfdEntry(tImageLength, dtShort, 1, uint16(g.Rows), g.ByteOrder))
	var bps = make([]uint16, g.samplesPerPixel)
	for i := 0; i < int(g.samplesPerPixel); i++ {
		bps[i] = uint16(g.BitsPerSample[i])
	}
	ifd = append(ifd, CreateIfdEntry(tBitsPerSample, dtShort, uint32(g.samplesPerPixel), bps, g.ByteOrder))
	ifd = append(ifd, CreateIfdEntry(tCompression, dtShort, 1, uint16(1), g.ByteOrder))
	ifd = append(ifd, CreateIfdEntry(tPhotometricInterpretation, dtShort, 1, uint16(g.PhotometricInterp), g.ByteOrder))
	stripOffsets := make([]uint32, g.Rows)
	stripByteCount := make([]uint32, g.Rows)
	rowLengthInBytes := uint32(g.Columns) * totalBytesPerPixel
	for i := 0; i < int(g.Rows); i++ {
		stripOffsets[i] = uint32(8 + rowLengthInBytes*uint32(i))
		stripByteCount[i] = rowLengthInBytes
	}
	ifd = append(ifd, CreateIfdEntry(tStripOffsets, dtLong, uint32(g.Rows), stripOffsets, g.ByteOrder))
	ifd = append(ifd, CreateIfdEntry(tSamplesPerPixel, dtShort, 1, uint16(g.samplesPerPixel), g.ByteOrder))
	ifd = append(ifd, CreateIfdEntry(tRowsPerStrip, dtShort, 1, uint16(1), g.ByteOrder))
	ifd = append(ifd, CreateIfdEntry(tStripByteCounts, dtLong, uint32(g.Rows), stripByteCount, g.ByteOrder))
	software := "GoSpatial"
	softwareLength := uint32(len(software))
	ifd = append(ifd, CreateIfdEntry(tSoftware, dtASCII, softwareLength, software, g.ByteOrder))

	sf := make([]uint16, g.samplesPerPixel)
	for i := 0; i < int(g.samplesPerPixel); i++ {
		sf[i] = uint16(g.SampleFormat)
	}
	ifd = append(ifd, CreateIfdEntry(tSampleFormat, dtShort, uint32(g.samplesPerPixel), sf, g.ByteOrder))

	if g.samplesPerPixel > 1 {
		ifd = append(ifd, CreateIfdEntry(tPlanarConfiguration, dtShort, 1, uint16(1), g.ByteOrder))
	}

	if g.PhotometricInterp == PI_RGB && g.samplesPerPixel == 4 {
		ifd = append(ifd, CreateIfdEntry(tExtraSamples, dtShort, 1, uint16(1), g.ByteOrder))
	}

	// There is currently no support for storing the image
	// resolution, so give a bogus value of 72x72 dpi.
	ifd = append(ifd, CreateIfdEntry(tXResolution, dtRational, 1, []uint32{72, 1}, g.ByteOrder))
	ifd = append(ifd, CreateIfdEntry(tYResolution, dtRational, 1, []uint32{72, 1}, g.ByteOrder))
	ifd = append(ifd, CreateIfdEntry(tResolutionUnit, dtShort, 1, []uint32{resPerInch}, g.ByteOrder))

	// Add the ModelPixelScaleTag and ModelTiepointTag tags
	ifd = append(ifd, CreateIfdEntry(tModelTiepointTag, dtDouble, 6, g.TiepointData.getModelTiepointTagData(), g.ByteOrder))
	ifd = append(ifd, CreateIfdEntry(tModelPixelScaleTag, dtDouble, 3, g.TiepointData.getModelPixelScaleTagData(), g.ByteOrder))

	if g.NodataValue != "" {
		ifd = append(ifd, CreateIfdEntry(tGDAL_NODATA, dtASCII, uint32(len(g.NodataValue)), g.NodataValue, g.ByteOrder))
	}

	// Create the geokeys
	geokeys := make([]IfdEntry, 0)
	if g.RasterPixelIsArea {
		geokeys = append(geokeys, CreateIfdEntry(tGTRasterTypeGeoKey, dtShort, 1, uint16(1), g.ByteOrder))
	} else { // RasterPixelIsPoint
		geokeys = append(geokeys, CreateIfdEntry(tGTRasterTypeGeoKey, dtShort, 1, uint16(2), g.ByteOrder))
	}

	asciiParams := CreateIfdEntry(tGeoAsciiParamsTag, dtASCII, 0, nil, g.ByteOrder)
	doubleParams := CreateIfdEntry(tGeoDoubleParamsTag, dtDouble, 0, nil, g.ByteOrder)

	if v, ok := geographicTypeMap[g.EPSGCode]; ok {
		geokeys = append(geokeys, CreateIfdEntry(tGTModelTypeGeoKey, dtShort, 1, uint16(2), g.ByteOrder))
		geokeys = append(geokeys, CreateIfdEntry(tGeographicTypeGeoKey, dtShort, 1, uint16(g.EPSGCode), g.ByteOrder))
		v += "|"
		v = strings.Replace(v, "_", " ", -1)
		geokeys = append(geokeys, CreateIfdEntry(tGTCitationGeoKey, dtASCII, uint32(len(v)), v, g.ByteOrder))
	} else if v, ok := projectedCSMap[g.EPSGCode]; ok {
		geokeys = append(geokeys, CreateIfdEntry(tGTModelTypeGeoKey, dtShort, 1, uint16(1), g.ByteOrder))
		geokeys = append(geokeys, CreateIfdEntry(tProjectedCSTypeGeoKey, dtShort, 1, uint16(g.EPSGCode), g.ByteOrder))
		v += "|"
		v = strings.Replace(v, "_", " ", -1)
		geokeys = append(geokeys, CreateIfdEntry(tGTCitationGeoKey, dtASCII, uint32(len(v)), v, g.ByteOrder))
	} else {
		if g.EPSGCode != 0 {
			panic(errors.New("Unrecognized EPSG code."))
		} else {
			v := "Unknown|"
			geokeys = append(geokeys, CreateIfdEntry(tGTCitationGeoKey, dtASCII, uint32(len(v)), v, g.ByteOrder))
		}
	}

	// sort the geokeys
	sort.Sort(ifdSortedByCode(geokeys))

	// create the GeoKeyDirectoryTag
	gkdtData := make([]uint16, 4+len(geokeys)*4)
	gkdtData[0] = 1
	gkdtData[1] = 1
	gkdtData[2] = 0
	gkdtData[3] = uint16(len(geokeys))
	for i, val := range geokeys {
		gkdtData[i*4+4] = uint16(val.tag.Code)
		if val.count < 5 {
			gkdtData[i*4+5] = 0
			gkdtData[i*4+6] = 1
			v, _ := val.InterpretDataAsInt()
			gkdtData[i*4+7] = uint16(v[0])
		} else {
			gkdtData[i*4+5] = 0
			gkdtData[i*4+6] = 1
			if val.dataType == dtASCII {
				gkdtData[i*4+7] = uint16(asciiParams.count)
				asciiParams.AddData(val.rawData)
				asciiParams.count += val.count
			} else if val.dataType == dtDouble {
				gkdtData[i*4+7] = uint16(doubleParams.count)
				doubleParams.AddData(val.rawData)
				doubleParams.count += val.count
			}

		}
	}

	ifd = append(ifd, CreateIfdEntry(tGeoKeyDirectoryTag, dtShort, uint32(len(gkdtData)), gkdtData, g.ByteOrder))

	if asciiParams.count > 0 {
		ifd = append(ifd, asciiParams)
	}
	if doubleParams.count > 0 {
		ifd = append(ifd, doubleParams)
	}

	// sort the ifd's
	sort.Sort(ifdSortedByCode(ifd))

	// output the ifd's
	writeIFD(w, int(imageLen+8), ifd, g.ByteOrder)

	// The IFD ends with the offset of the next IFD in the file,
	// or zero if it is the last one (page 14).
	if err := binary.Write(w, g.ByteOrder, uint32(0)); err != nil {
		return err
	}

	w.Flush()

	// use ifd to create the ifdList, which is really a map
	g.ifdList = make(map[int]IfdEntry)
	g.geoKeyList = make(map[int]IfdEntry)
	for _, val := range ifd {
		g.ifdList[val.tag.Code] = val
	}

	for _, val := range geokeys {
		g.geoKeyList[val.tag.Code] = val
	}

	return err
}

func writeIFD(w io.Writer, ifdOffset int, d []IfdEntry, enc binary.ByteOrder) error {
	var buf [ifdLen]byte
	// Make space for "pointer area" containing IFD entry data
	// longer than 4 bytes.
	parea := make([]byte, 1024)
	pstart := ifdOffset + ifdLen*len(d) + 6
	var o int // Current offset in parea.

	// The IFD has to be written with the tags in ascending order.
	sort.Sort(ifdSortedByCode(d))

	// Write the number of entries in this IFD.
	if err := binary.Write(w, enc, uint16(len(d))); err != nil {
		return err
	}
	for _, ent := range d {
		enc.PutUint16(buf[0:2], uint16(ent.tag.Code))
		enc.PutUint16(buf[2:4], uint16(ent.dataType))
		count := uint32(ent.count)
		enc.PutUint32(buf[4:8], count)
		datalen := int(count * lengths[ent.dataType])
		if datalen <= 4 {
			for i, b := range ent.rawData {
				buf[8+i] = b
			}
		} else {
			if (o + datalen) > len(parea) {
				newlen := len(parea) + 1024
				for (o + datalen) > newlen {
					newlen += 1024
				}
				newarea := make([]byte, newlen)
				copy(newarea, parea)
				parea = newarea
			}
			for i, b := range ent.rawData {
				parea[o+i] = b
			}
			enc.PutUint32(buf[8:12], uint32(pstart+o))
			o += datalen
		}
		if _, err := w.Write(buf[:]); err != nil {
			return err
		}
	}
	// The IFD ends with the offset of the next IFD in the file,
	// or zero if it is the last one (page 14).
	if err := binary.Write(w, enc, uint32(0)); err != nil {
		return err
	}
	_, err := w.Write(parea[:o])
	return err
}

func (g *GeoTIFF) Read(fileName string) (err error) {
	// initialize some things
	g.ifdList = make(map[int]IfdEntry)
	g.geoKeyList = make(map[int]IfdEntry)
	g.off = 0

	// open the file
	f, err := os.Open(fileName)
	if err != nil {
		return FileOpeningError
	}
	defer f.Close()

	g.r = f

	p := make([]byte, 8)
	if _, err := g.r.ReadAt(p, 0); err != nil && err != io.EOF {
		return FileIsNotProperlyFormated
	}
	switch string(p[0:4]) {
	case leHeader:
		g.ByteOrder = binary.LittleEndian
	case beHeader:
		g.ByteOrder = binary.BigEndian
	default:
		if p[2] == 43 || p[3] == 43 {
			return errors.New("The BigTiff format is not currently supported.")
		} else {
			return FileIsNotProperlyFormated
		}
	}

	offset := int64(g.ByteOrder.Uint32(p[4:8]))

	for offset > 0 {
		offset, err = g.readIFD(offset)
		if err != nil {
			return err
		}
		g.parseGeoKeys()
	}

	//fmt.Println(g.GetTags())

	g.Columns = g.firstVal(tImageWidth)
	g.Rows = g.firstVal(tImageLength)
	g.PhotometricInterp = g.firstVal(tPhotometricInterpretation)
	if ifd, ok := g.ifdList[tBitsPerSample]; ok {
		g.BitsPerSample, _ = ifd.InterpretDataAsInt()
	}

	g.samplesPerPixel = g.firstVal(tSamplesPerPixel)
	g.SampleFormat = g.firstVal(tSampleFormat)

	// See if geokeys has GTRasterTypeGeoKey
	if ifd, ok := g.geoKeyList[tGTRasterTypeGeoKey]; ok {
		val, _ := ifd.InterpretDataAsInt()
		if val[0] == 1 {
			g.RasterPixelIsArea = true
		} else {
			g.RasterPixelIsArea = false
		}
	}

	// Get the EPSG code
	if ifd, ok := g.geoKeyList[tProjectedCSTypeGeoKey]; ok {
		if val, err := ifd.InterpretDataAsInt(); err == nil {
			g.EPSGCode = val[0]
		}
	} else if ifd, ok := g.geoKeyList[tGeographicTypeGeoKey]; ok {
		if val, err := ifd.InterpretDataAsInt(); err == nil {
			g.EPSGCode = val[0]
		}
	}

	// see if the GDAL_NODATA tag has been set
	if ifd, err := g.FindIFDEntryFromCode(tGDAL_NODATA); err == nil {
		strArray, err := ifd.InterpretDataAsASCII()
		//fmt.Println(strArray[0])
		if err == nil {
			g.NodataValue = strArray[0]
		} else {
			return err
		}
	}
	//if entry, err := g.FindIFDEntryFromCode(tGDAL_NODATA); err != TagNotFoundError {
	//	strArray, err := entry.InterpretDataAsASCII()
	//	if err == nil {
	//		g.NodataValue = strArray[0]
	//	} else {
	//		return err
	//	}
	//}

	// Determine the image mode.
	switch g.PhotometricInterp {
	case PI_RGB:
		if g.BitsPerSample[0] == 16 {
			for _, b := range g.BitsPerSample {
				if b != 16 {
					err = errors.New("wrong number of samples for 16bit RGB")
					return
				}
			}
		} else {
			for _, b := range g.BitsPerSample {
				if b != 8 {
					err = errors.New("wrong number of samples for 8bit RGB")
					return
				}
			}
		}
		// RGB images normally have 3 samples per pixel.
		// If there are more, ExtraSamples (p. 31-32 of the spec)
		// gives their meaning (usually an alpha channel).
		//
		// This implementation does not support extra samples
		// of an unspecified type.
		switch len(g.BitsPerSample) {
		case 3:
			g.mode = mRGB
		case 4:
			switch g.firstVal(tExtraSamples) {
			case 1:
				g.mode = mRGBA
			case 2:
				g.mode = mNRGBA
			default:
				err = errors.New("wrong number of samples for RGB")
				return
			}
		default:
			err = errors.New("wrong number of samples for RGB")
			return
		}
	case PI_Paletted:
		g.mode = mPaletted
		// retreive the palette colour data
		if ifd, ok := g.ifdList[tColorMap]; ok {
			val, err := ifd.InterpretDataAsInt()
			if err != nil {
				return err
			}
			numcolors := len(val) / 3
			if len(val)%3 != 0 || numcolors <= 0 || numcolors > 256 {
				return errors.New("bad ColorMap length")
			}
			g.palette = make([]uint32, numcolors)
			for i := 0; i < numcolors; i++ {
				// colours in the colour map are given in 16-bit channels
				// and need to be rescaled to an 8-bit format.
				red := uint32(float64(val[i]) / 65535.0 * 255.0)
				green := uint32(float64(val[i+numcolors]) / 65535.0 * 255.0)
				blue := uint32(float64(val[i+2*numcolors]) / 65535.0 * 255.0)
				a := uint32(255)
				val := uint32((a << 24) | (red << 16) | (green << 8) | blue)
				g.palette[i] = val
			}
		} else {
			err = errors.New("Could not locate the colour map tag.")
			return
		}
	case PI_WhiteIsZero:
		g.mode = mGrayInvert
	case PI_BlackIsZero:
		g.mode = mGray
	default:
		err = errors.New("Unsupported image format")
		return
	}

	g.readData()

	return
}

func (g *GeoTIFF) readData() (err error) {
	compressionType := g.firstVal(tCompression)
	g.SampleFormat = g.firstVal(tSampleFormat)

	width := int(g.Columns)
	height := int(g.Rows)
	//if g.mode == mGray || g.mode == mGrayInvert {
	g.Data = make([]float64, width*height)
	//} else {
	//	g.ColorData = make([]color.Color, width*height)
	//}

	blockPadding := false
	blockWidth := int(g.Columns)
	blockHeight := int(g.Rows)
	blocksAcross := 1
	blocksDown := 1

	var blockOffsets, blockCounts []uint

	if int(g.firstVal(tTileWidth)) != 0 {
		blockPadding = true

		blockWidth = int(g.firstVal(tTileWidth))
		blockHeight = int(g.firstVal(tTileLength))

		blocksAcross = (width + blockWidth - 1) / blockWidth
		blocksDown = (height + blockHeight - 1) / blockHeight

		if ifd, ok := g.ifdList[tTileOffsets]; ok {
			blockOffsets, _ = ifd.InterpretDataAsInt()
		}
		if ifd, ok := g.ifdList[tTileByteCounts]; ok {
			blockCounts, _ = ifd.InterpretDataAsInt()
		}

	} else {
		if int(g.firstVal(tRowsPerStrip)) != 0 {
			blockHeight = int(g.firstVal(tRowsPerStrip))
		}

		blocksDown = (height + blockHeight - 1) / blockHeight

		if ifd, ok := g.ifdList[tStripOffsets]; ok {
			blockOffsets, _ = ifd.InterpretDataAsInt()
		}
		if ifd, ok := g.ifdList[tStripByteCounts]; ok {
			blockCounts, _ = ifd.InterpretDataAsInt()
		}
	}

	for i := 0; i < blocksAcross; i++ {
		blkW := blockWidth
		if !blockPadding && i == blocksAcross-1 && width%blockWidth != 0 {
			blkW = width % blockWidth
		}
		for j := 0; j < blocksDown; j++ {
			blkH := blockHeight
			if !blockPadding && j == blocksDown-1 && height%blockHeight != 0 {
				blkH = height % blockHeight
			}
			offset := int64(blockOffsets[j*blocksAcross+i])
			n := int64(blockCounts[j*blocksAcross+i])
			switch compressionType {
			case cNone:
				if b, ok := g.r.(*buffer); ok {
					g.buf, err = b.Slice(int(offset), int(n))
				} else {
					g.buf = make([]byte, n)
					_, err = g.r.ReadAt(g.buf, offset)
				}
			case cLZW:
				r := lzw.NewReader(io.NewSectionReader(g.r, offset, n), lzw.MSB, 8)
				defer r.Close()
				g.buf, err = ioutil.ReadAll(r)
				if err != nil {
					println(err)
					//println("Block X: ", i, "Block Y: ", j, "Offset: ", offset, "n: ", n, "buf len: ", len(g.buf))
					//	panic(err)
				}
			case cDeflate, cDeflateOld:
				r, err := zlib.NewReader(io.NewSectionReader(g.r, offset, n))
				if err != nil {
					return err
				}
				g.buf, err = ioutil.ReadAll(r)
				r.Close()
			case cPackBits:

			default:
				err = errors.New(fmt.Sprintf("Unsupported compression value %d", compressionType))

			}
			xmin := i * blockWidth
			ymin := j * blockHeight
			xmax := xmin + blkW
			ymax := ymin + blkH

			xmax = minInt(xmax, width)
			ymax = minInt(ymax, height)

			g.off = 0

			// Apply horizontal predictor if necessary.
			// In this case, p contains the color difference to the preceding pixel.
			// See page 64-65 of the spec.
			if g.firstVal(tPredictor) == prHorizontal {
				// does it make sense to extend this to 32 and 64 bits?
				if g.BitsPerSample[0] == 16 {
					var off int
					spp := len(g.BitsPerSample) // samples per pixel
					bpp := spp * 2              // bytes per pixel
					for y := ymin; y < ymax; y++ {
						off += spp * 2
						for x := 0; x < (xmax-xmin-1)*bpp; x += 2 {
							v0 := g.ByteOrder.Uint16(g.buf[off-bpp : off-bpp+2])
							v1 := g.ByteOrder.Uint16(g.buf[off : off+2])
							g.ByteOrder.PutUint16(g.buf[off:off+2], v1+v0)
							off += 2
						}
					}
				} else if g.BitsPerSample[0] == 8 {
					var off int
					spp := len(g.BitsPerSample) // samples per pixel
					for y := ymin; y < ymax; y++ {
						off += spp
						for x := 0; x < (xmax-xmin-1)*spp; x++ {
							g.buf[off] += g.buf[off-spp]
							off++
						}
					}
				}
			}

			switch g.mode {
			case mGray, mGrayInvert:
				switch g.SampleFormat {
				case 1: // Unsigned integer data
					switch g.BitsPerSample[0] {
					case 8:
						for y := ymin; y < ymax; y++ {
							for x := xmin; x < xmax; x++ {
								i := y*width + x
								g.Data[i] = float64(g.buf[g.off])
								g.off++
							}
						}
					case 16:
						for y := ymin; y < ymax; y++ {
							for x := xmin; x < xmax; x++ {
								value := g.ByteOrder.Uint16(g.buf[g.off : g.off+2])
								i := y*width + x
								g.Data[i] = float64(value)
								g.off += 2
							}
						}
					case 32:
						for y := ymin; y < ymax; y++ {
							for x := xmin; x < xmax; x++ {
								value := g.ByteOrder.Uint32(g.buf[g.off : g.off+4])
								i := y*width + x
								g.Data[i] = float64(value)
								g.off += 4
							}
						}
					case 64:
						for y := ymin; y < ymax; y++ {
							for x := xmin; x < xmax; x++ {
								value := g.ByteOrder.Uint64(g.buf[g.off : g.off+8])
								i := y*width + x
								g.Data[i] = float64(value)
								g.off += 8
							}
						}
					default:
						err = errors.New("Unsupported data format")
						return
					}
				case 2: // Signed integer data
					switch g.BitsPerSample[0] {
					case 8:
						for y := ymin; y < ymax; y++ {
							for x := xmin; x < xmax; x++ {
								i := y*width + x
								g.Data[i] = float64(int8(g.buf[g.off]))
								g.off++
							}
						}
					case 16:
						for y := ymin; y < ymax; y++ {
							for x := xmin; x < xmax; x++ {
								value := int16(g.ByteOrder.Uint16(g.buf[g.off : g.off+2]))
								i := y*width + x
								g.Data[i] = float64(value)
								g.off += 2
							}
						}
					case 32:
						for y := ymin; y < ymax; y++ {
							for x := xmin; x < xmax; x++ {
								value := int32(g.ByteOrder.Uint32(g.buf[g.off : g.off+4]))
								i := y*width + x
								g.Data[i] = float64(value)
								g.off += 4
							}
						}
					case 64:
						for y := ymin; y < ymax; y++ {
							for x := xmin; x < xmax; x++ {
								value := int64(g.ByteOrder.Uint64(g.buf[g.off : g.off+8]))
								i := y*width + x
								g.Data[i] = float64(value)
								g.off += 8
							}
						}
					default:
						err = errors.New("Unsupported data format")
						return
					}
				case 3: // Floating point data
					switch g.BitsPerSample[0] {
					case 32:
						for y := ymin; y < ymax; y++ {
							for x := xmin; x < xmax; x++ {
								if g.off <= len(g.buf) {
									bits := g.ByteOrder.Uint32(g.buf[g.off : g.off+4])
									float := math.Float32frombits(bits)
									i := y*width + x
									g.Data[i] = float64(float)
									g.off += 4
								}
							}
						}
					case 64:
						for y := ymin; y < ymax; y++ {
							for x := xmin; x < xmax; x++ {
								if g.off <= len(g.buf) {
									bits := g.ByteOrder.Uint64(g.buf[g.off : g.off+8])
									float := math.Float64frombits(bits)
									i := y*width + x
									g.Data[i] = float
									g.off += 8
								}
							}
						}
					default:
						err = errors.New("Unsupported data format")
						return
					}
				default:
					err = errors.New("Unsupported sample format")
					return
				}
			case mPaletted:
				for y := ymin; y < ymax; y++ {
					for x := xmin; x < xmax; x++ {
						i := y*width + x
						val := int(g.buf[g.off])
						g.Data[i] = float64(g.palette[val])
						g.off++
					}
				}

			case mRGB:
				if g.BitsPerSample[0] == 8 {
					for y := ymin; y < ymax; y++ {
						for x := xmin; x < xmax; x++ {
							red := uint32(g.buf[g.off])
							green := uint32(g.buf[g.off+1])
							blue := uint32(g.buf[g.off+2])
							a := uint32(255)
							g.off += 3
							i := y*width + x
							val := uint32((a << 24) | (red << 16) | (green << 8) | blue)
							g.Data[i] = float64(val)
						}
					}
				} else if g.BitsPerSample[0] == 16 {
					for y := ymin; y < ymax; y++ {
						for x := xmin; x < xmax; x++ {
							// the spec doesn't talk about 16-bit RGB images so
							// I'm not sure why I bother with this. They specifically
							// say that RGB images are 8-bits per channel. Anyhow,
							// I rescale the 16-bits to an 8-bit channel for simplicity.
							red := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+0:g.off+2])) / 65535.0 * 255.0)
							green := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+2:g.off+4])) / 65535.0 * 255.0)
							blue := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+4:g.off+6])) / 65535.0 * 255.0)
							a := uint32(255)
							g.off += 6
							i := y*width + x
							val := uint32((a << 24) | (red << 16) | (green << 8) | blue)
							g.Data[i] = float64(val)
						}
					}
				} else {
					err = errors.New("Unsupported data format")
					return
				}
			case mNRGBA:
				if g.BitsPerSample[0] == 8 {
					for y := ymin; y < ymax; y++ {
						for x := xmin; x < xmax; x++ {
							red := uint32(g.buf[g.off])
							green := uint32(g.buf[g.off+1])
							blue := uint32(g.buf[g.off+2])
							a := uint32(g.buf[g.off+3])
							g.off += 4
							i := y*width + x
							val := uint32((a << 24) | (red << 16) | (green << 8) | blue)
							g.Data[i] = float64(val)
						}
					}
				} else if g.BitsPerSample[0] == 16 {
					for y := ymin; y < ymax; y++ {
						for x := xmin; x < xmax; x++ {
							red := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+0:g.off+2])) / 65535.0 * 255.0)
							green := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+2:g.off+4])) / 65535.0 * 255.0)
							blue := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+4:g.off+6])) / 65535.0 * 255.0)
							a := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+6:g.off+8])) / 65535.0 * 255.0)
							g.off += 8
							i := y*width + x
							val := uint32((a << 24) | (red << 16) | (green << 8) | blue)
							g.Data[i] = float64(val)
						}
					}
				} else {
					err = errors.New("Unsupported data format")
					return
				}
			case mRGBA:
				if g.BitsPerSample[0] == 16 {
					for y := ymin; y < ymax; y++ {
						for x := xmin; x < xmax; x++ {
							red := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+0:g.off+2])) / 65535.0 * 255.0)
							green := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+2:g.off+4])) / 65535.0 * 255.0)
							blue := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+4:g.off+6])) / 65535.0 * 255.0)
							a := uint32(float64(g.ByteOrder.Uint16(g.buf[g.off+6:g.off+8])) / 65535.0 * 255.0)
							g.off += 8
							i := y*width + x
							val := uint32((a << 24) | (red << 16) | (green << 8) | blue)
							g.Data[i] = float64(val)
						}
					}
				} else {
					for y := ymin; y < ymax; y++ {
						for x := xmin; x < xmax; x++ {
							red := uint32(g.buf[g.off])
							green := uint32(g.buf[g.off+1])
							blue := uint32(g.buf[g.off+2])
							a := uint32(g.buf[g.off+3])
							g.off += 4
							i := y*width + x
							val := uint32((a << 24) | (red << 16) | (green << 8) | blue)
							g.Data[i] = float64(val)
						}
					}
				}
			}
		}
	}
	return nil
}

func (g *GeoTIFF) GetTags() (ret string) {
	ret = "IMAGE TAG ENTRIES:\n"
	ifd := make([]IfdEntry, 0)
	for _, entry := range g.ifdList {
		ifd = append(ifd, entry)
	}
	sort.Sort(ifdSortedByCode(ifd))
	for _, entry := range ifd {
		ret += entry.String() + "\n"
	}

	if len(g.geoKeyList) > 0 {
		ret += "\nIMAGE GEOKEY ENTRIES:\n"
		ifd = make([]IfdEntry, 0)
		for _, entry := range g.geoKeyList {
			ifd = append(ifd, entry)
		}
		sort.Sort(ifdSortedByCode(ifd))
		for _, entry := range ifd {
			ret += entry.String() + "\n"
		}
	}
	return ret
}

func (g *GeoTIFF) readIFD(offset int64) (nextIFDOffset int64, err error) {
	p := make([]byte, 8)
	// The first two bytes contain the number of entries (12 bytes each).
	if _, err := g.r.ReadAt(p[0:2], offset); err != nil && err != io.EOF {
		return -1, FileIsNotProperlyFormated
	}
	numItems := int(g.ByteOrder.Uint16(p[0:2]))

	// All IFD entries are read in one chunk.
	p = make([]byte, ifdLen*numItems)
	if _, err := g.r.ReadAt(p, offset+2); err != nil && err != io.EOF {
		return -1, err
	}

	for i := 0; i < len(p); i += ifdLen {
		if err := g.parseEntry(p[i : i+ifdLen]); err != nil {
			//return -1, err
			panic(err)
		}
	}

	// get the offset to the next IFD
	p = make([]byte, 5)
	offset += int64(2 + ifdLen*numItems)
	if _, err = g.r.ReadAt(p[0:5], offset); err != nil {
		return -1, FileIsNotProperlyFormated
	}
	nextIFDOffset = int64(g.ByteOrder.Uint32(p[0:5]))
	return nextIFDOffset, nil
}

func (g *GeoTIFF) parseEntry(p []byte) error {
	var newEntry IfdEntry
	tagNum := int(g.ByteOrder.Uint16(p[0:2]))
	if myTag, ok := tagMap[tagNum]; !ok {
		// unrecognized tag
		printf("Unrecognized tag: %d\n", tagNum)
		//return errors.New("Unrecognized tag.")
	} else {
		newEntry.tag = myTag
	}

	var raw []byte
	dt := g.ByteOrder.Uint16(p[2:4])
	newEntry.dataType = GeotiffDataType(dt)
	newEntry.count = g.ByteOrder.Uint32(p[4:8])
	if datalen := newEntry.dataType.GetBitLength() * newEntry.count; datalen > 4 {
		// The IFD contains a pointer to the real value.
		raw = make([]byte, datalen)
		_, err := g.r.ReadAt(raw, int64(g.ByteOrder.Uint32(p[8:12])))
		if err != nil && err != io.EOF {
			println(int64(g.ByteOrder.Uint32(p[8:12])))

			printf("Data Length: %d, Bit Length: %d, Count: %d\n", datalen, newEntry.dataType.GetBitLength(), newEntry.count)
			s := fmt.Sprintf("Error: %v; Encountered on tag: %v\n", err, newEntry.tag)
			panic(errors.New(s))

		}
	} else {
		raw = p[8 : 8+datalen]
	}

	newEntry.rawData = raw
	newEntry.byteOrder = g.ByteOrder

	g.ifdList[newEntry.tag.Code] = newEntry

	return nil
}

func (g *GeoTIFF) parseGeoKeys() error {
	// get the the GeoKeyDirectoryTag
	if gkDir, err := g.FindIFDEntryFromCode(tGeoKeyDirectoryTag); err == nil { //
		// parse the geokeys
		d, err := gkDir.InterpretDataAsInt()
		if err != nil {
			return err
		}
		g.NumGeoKeys = int(d[3])
		for i := 4; i < len(d); i += 4 {
			var newGeoKey IfdEntry
			newGeoKey.byteOrder = g.ByteOrder
			tagNum := int(d[i])
			if myTag, ok := tagMap[tagNum]; !ok {
				// unrecognized tag
				fmt.Printf("Could not find tag %d\n", tagNum)
				//panic(errors.New("Unrecognized tag."))
			} else {
				newGeoKey.tag = myTag
			}
			tagLoc := d[i+1]
			newGeoKey.count = uint32(d[i+2])
			valOffset := d[i+3]
			if tagLoc == 0 {
				// it's a short and valOffset IS the data
				b := make([]byte, 2)
				g.ByteOrder.PutUint16(b, uint16(valOffset))
				newGeoKey.rawData = b
				newGeoKey.dataType = DT_Short

			} else {
				// it's either going to be located in GeoDoubleParamsTag
				// or GeoAsciiParamsTag at valOffset
				if tagLoc == tGeoDoubleParamsTag { // 34736 it's a double
					// first get the GeoDoubleParamsTag
					if gkDoubleParams, err := g.FindIFDEntryFromCode(tGeoDoubleParamsTag); err == nil {
						// I think that the offset is "based on the natural data type", which in this case is the number of
						// 8-byte doubles. Unfortunately the GeoTiff specs don't clarify this.
						raw := gkDoubleParams.rawData[valOffset*8 : valOffset*8+uint(newGeoKey.count)]
						newGeoKey.rawData = raw
						newGeoKey.dataType = DT_Double
					} else {
						panic(errors.New("Could not locate the GeoAsciiParamsTag. The file may not be a GeoTIFF file."))
					}
				} else if tagLoc == tGeoAsciiParamsTag { // 34737 it's an ASCII field
					// first get the GeoAsciiParamsTag
					if gkAsciiParams, err := g.FindIFDEntryFromCode(tGeoAsciiParamsTag); err == nil {
						raw := gkAsciiParams.rawData[valOffset : valOffset+uint(newGeoKey.count)]
						newGeoKey.rawData = raw
						newGeoKey.dataType = DT_ASCII
					} else {
						panic(errors.New("Could not locate the GeoAsciiParamsTag. The file may not be a GeoTIFF file."))
					}

				}
			}
			//println(newGeoKey)
			g.geoKeyList[newGeoKey.tag.Code] = newGeoKey

		}
	} else {
		panic(errors.New("Could not locate the GeoKeyDirectory. The file may not be a GeoTIFF file."))
	}
	return nil
}

func (g *GeoTIFF) FindIFDEntryFromCode(tagCode int) (*IfdEntry, error) {
	for _, ifd := range g.ifdList {
		if ifd.tag.Code == tagCode {
			return &ifd, nil
		}
	}
	return nil, TagNotFoundError
}

func (g *GeoTIFF) FindIFDEntryFromName(tagName string) (*IfdEntry, error) {
	for _, ifd := range g.ifdList {
		if ifd.tag.Name == tagName {
			return &ifd, nil
		}
	}

	for _, ifd := range g.geoKeyList {
		if ifd.tag.Name == tagName {
			return &ifd, nil
		}
	}
	return nil, TagNotFoundError
}

// firstVal returns the first uint of the features entry with the given tag,
// or 0 if the tag does not exist.
func (g *GeoTIFF) firstVal(tag int) uint {
	// which map is the tag in? The ifdList or geoKeyList?
	if v, ok := g.ifdList[tag]; ok {
		if v.dataType == DT_Short || v.dataType == DT_Byte || v.dataType == DT_Long {
			if v2, err := v.InterpretDataAsInt(); err == nil {
				return v2[0]
			} else {
				return 0
			}
		} else {
			return 0
		}
	} else if v, ok = g.geoKeyList[tag]; ok {
		if v.dataType == DT_Short || v.dataType == DT_Byte || v.dataType == DT_Long {
			if v2, err := v.InterpretDataAsInt(); err == nil {
				return v2[0]
			} else {
				return 0
			}
		} else {
			return 0
		}
	} else {
		return 0
	}
	return 0
}

func minInt(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

type TiepointTransformationParameters struct {
	I, J, K, X, Y, Z       float64
	ScaleX, ScaleY, ScaleZ float64
}

func (t *TiepointTransformationParameters) getModelTiepointTagData() []float64 {
	ret := []float64{t.I, t.J, t.K, t.X, t.Y, t.Z}
	return ret
}

func (t *TiepointTransformationParameters) getModelPixelScaleTagData() []float64 {
	ret := []float64{t.ScaleX, t.ScaleY, t.ScaleZ}
	return ret
}

// errors
var FileIsNotProperlyFormated = errors.New("The file does not appear to be properly formatted")
var FileOpeningError = errors.New("An error occurred while opening the data file.")
var UnsupportedDataTypeError = errors.New("Unsupported data type")
var TagNotFoundError = errors.New("Tag not found")
