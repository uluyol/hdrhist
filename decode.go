package hdrhist

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io/ioutil"
	"math"

	"github.com/pkg/errors"
)

const (
	encodingV0CookieBase           = 0x1c849308
	compressedEncodingV0CookieBase = 0x1c849309

	encodingV1CookieBase           = 0x1c849301
	compressedEncodingV1CookieBase = 0x1c849302

	encodingV2CookieBase           = 0x1c849303
	compressedEncodingV2CookieBase = 0x1c849304

	encodingV2maxWordSize = 9

	encodingHeaderSize   = 40
	encodingV0HeaderSize = 32
)

func decodeCompressed(h *Hist, buf []byte) error {
	const doubleHistCookie = 0x0c72124e
	const doubleHistCompressedCookie = 0x0c72124f

	r := bytes.NewReader(buf)
	var cookie int32
	if err := binary.Read(r, binary.BigEndian, &cookie); err != nil {
		return errors.Wrap(err, "unable to decode cookie")
	}

	if cookie == doubleHistCookie || cookie == doubleHistCompressedCookie {
		return errors.New("double histograms are unsupported")
	}

	var headerSize int
	switch cookie & ^0xf0 {
	case compressedEncodingV1CookieBase, compressedEncodingV2CookieBase:
		headerSize = encodingHeaderSize
	case compressedEncodingV0CookieBase:
		headerSize = encodingV0HeaderSize
	default:
		return errors.New("no histogram in buffer")
	}
	var compressedLen int32
	if err := binary.Read(r, binary.BigEndian, &compressedLen); err != nil {
		return errors.Wrap(err, "unable decode length")
	}
	zr, err := zlib.NewReader(bytes.NewReader(buf[8 : 8+compressedLen]))
	if err != nil {
		return errors.Wrap(err, "can't create decompressor")
	}
	b, err := ioutil.ReadAll(zr)
	zr.Close()
	if err != nil {
		return errors.Wrap(err, "unable to decompress encoded hist")
	}

	return decode(h, b[:headerSize], b[headerSize:])
}

func decode(h *Hist, headerBuf, buf []byte) error {
	hr := bytes.NewReader(headerBuf)
	var cookie int32
	if err := binary.Read(hr, binary.BigEndian, &cookie); err != nil {
		return errors.Wrap(err, "unable to read cookie")
	}
	var (
		payloadLen              int32
		normalizingIndexOff     int32 // ignored
		sigfigs                 int32
		lowestDiscernible       int64
		highestTrackable        int64
		intToF64ConversionRatio float64 // ignored
	)
	switch cookie & ^0xf0 {
	case encodingV1CookieBase, encodingV2CookieBase:
		vals := []struct {
			dest interface{}
			name string
		}{
			{&payloadLen, "payload size"},
			{&normalizingIndexOff, "normalizing index offset"},
			{&sigfigs, "sigfig count"},
			{&lowestDiscernible, "lowest discernible value"},
			{&highestTrackable, "highest trackable value"},
			{&intToF64ConversionRatio, "int to double conversion ratio"},
		}
		for _, v := range vals {
			if err := binary.Read(hr, binary.BigEndian, v.dest); err != nil {
				return errors.Wrapf(err, "unable to read %s", v.name)
			}
		}
	case encodingV0CookieBase:
		return errors.New("v0 encoding not supported")
	default:
		return errors.New("no valid cookie found")
	}
	h.Init(Config{
		LowestDiscernible: lowestDiscernible,
		HighestTrackable:  highestTrackable,
		SigFigs:           sigfigs,
	})
	h.Clear()

	// TODO: consider handling uncompressed histograms where
	//       headerBuf contains the full data and buf is nil

	if int(payloadLen) > len(buf) {
		return errors.New("buffer does not contain full payload")
	}

	if cookie & ^0xf0 == encodingV1CookieBase {
		return errors.New("v1 encoding not supported")
	}

	_, err := fillCountsFromBuf(h, buf, int(payloadLen))
	if err != nil {
		return err
	}

	return nil
}

func fillCountsFromBuf(h *Hist, buf []byte, payloadLen int) (int, error) {
	desti := 0
	pos := 0
	for pos < payloadLen {
		count, clen, err := decodeZigZag(buf[pos:])
		pos += clen
		if err != nil {
			return 0, errors.Wrap(err, "invalid count")
		}
		var zerosCount int32
		if count < 0 {
			zc := -count
			if zc > math.MaxInt32 {
				return 0, errors.New("got zero count > math.MaxInt32")
			}
			zerosCount = int32(zc)
		}
		if zerosCount > 0 {
			desti += int(zerosCount)
		} else {
			h.b.counts[desti] = count
			h.totalCount += count
			desti++
		}
	}
	return desti, nil
}