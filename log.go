package hdrhist

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func splitLog(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	for i := range data {
		switch data[i] {
		case '\r', '\n', ' ', ',':
			return i + 1, data[0:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

type LogReader struct {
	s   *bufio.Scanner
	err error

	startTime time.Time
	baseTime  struct {
		sec  int64
		nano int64
	}

	foundStartTime bool
	foundBaseTime  bool

	cur *Hist
}

func NewLogReader(r io.Reader) *LogReader {
	s := bufio.NewScanner(r)
	return &LogReader{
		s: s,
	}
}

func getFloat64Prefix(s string) (float64, error) {
	s = strings.TrimSpace(s)
	hasPeriod := false
	end := len(s)
	for i := 0; i < len(s); i++ {
		if '0' <= s[i] && s[i] <= '9' {
			continue
		}
		if s[i] == '.' && !hasPeriod {
			hasPeriod = true
			continue
		}
		end = i
		break
	}
	return strconv.ParseFloat(s[:end], 64)
}

func (l *LogReader) Scan() bool {
	if l.err != nil {
		return false
	}
	for l.s.Scan() {
		s := l.s.Text()
		switch {
		case strings.HasPrefix(s, "#[StartTime:"):
			t, err := getFloat64Prefix(strings.TrimPrefix(s, "#[StartTime:"))
			if err != nil {
				l.err = errors.Wrap(err, "unable to parse start time")
				return false
			}
			sec, nano := math.Modf(t)
			l.startTime = time.Unix(int64(sec), int64(nano*1e9))
			l.foundStartTime = true
			continue
		case strings.HasPrefix(s, "#[BaseTime:"):
			t, err := getFloat64Prefix(strings.TrimPrefix(s, "#[BaseTime:"))
			if err != nil {
				l.err = errors.Wrap(err, "unable to parse base time")
				return false
			}
			sec, nano := math.Modf(t)
			l.baseTime.sec = int64(sec)
			l.baseTime.nano = int64(nano * 1e9)
			l.foundBaseTime = true
			continue
		case strings.HasPrefix(s, "\"StartTimestamp\""), strings.HasPrefix(s, "#"):
			// skip legend and comments
			continue
		}

		scanner := bufio.NewScanner(bytes.NewReader(l.s.Bytes()))
		scanner.Split(splitLog)

		if !scanner.Scan() {
			if scanner.Err() != nil {
				l.err = errors.Wrap(scanner.Err(), "no histogram data")
				return false
			} else {
				continue
			}
		}

		s = scanner.Text()

		// decode Tag=[junk],
		if strings.HasPrefix(s, "Tag=") {
			// skip histogram tags, we don't support them
			if !scanner.Scan() {
				l.err = errors.New("malformed input, expected start timestamp")
				return false
			}
			s = scanner.Text()
		}

		// decode startTimestamp,intervalLength,maxval,histPayload
		t, err := strconv.ParseFloat(s, 64)
		if err != nil {
			l.err = errors.Wrap(err, "invalid timestamp")
			return false
		}

		sec, nano := math.Modf(t)
		tstamp := time.Unix(int64(sec), int64(nano*1e9))
		if !l.foundStartTime {
			l.startTime = tstamp
			l.foundStartTime = true
		}

		if !l.foundBaseTime {
			if l.startTime.Sub(tstamp) > 365*24*time.Hour {
				// NOTE: assume that timestamps in the log are not absolute
				// if the log timestamp is > 1 year ago
				l.baseTime.sec = l.startTime.Unix()
				l.baseTime.nano = int64(l.startTime.Nanosecond())
			} else {
				l.baseTime.sec = 0
				l.baseTime.nano = 0
			}
			l.foundBaseTime = true
		}
		// need to create tstamp twice because duration might overflow
		tstamp = time.Unix(tstamp.Unix()+l.baseTime.sec, int64(tstamp.Nanosecond())+l.baseTime.nano)

		if !scanner.Scan() {
			l.err = errors.New("malformed input, expected interval length")
			return false
		}
		s = scanner.Text()
		t, err = strconv.ParseFloat(s, 64)
		if err != nil {
			l.err = errors.Wrap(err, "invalid interval length")
			return false
		}
		sec, nano = math.Modf(t)
		tstampEnd := tstamp.Add(time.Duration(sec)*time.Second + time.Duration(nano*1e9)*time.Nanosecond)

		if !scanner.Scan() {
			l.err = errors.New("malformed input, expected max hist value")
			return false
		}
		// skip max hist value, already is in the histogram

		if !scanner.Scan() {
			l.err = errors.New("malformed input, expect encoded histogram")
			return false
		}

		n := base64.StdEncoding.DecodedLen(len(scanner.Bytes()))
		buf := make([]byte, n)
		_, err = base64.StdEncoding.Decode(buf, scanner.Bytes())
		if err != nil {
			l.err = errors.Wrap(err, "malformed base64 histogram")
			return false
		}
		var hist Hist
		err = decodeCompressed(&hist, buf)
		if err != nil {
			l.err = errors.Wrap(err, "unable to decode histogram")
			return false
		}

		hist.SetStartTime(tstamp)
		hist.SetEndTime(tstampEnd)

		l.cur = &hist
		return true
	}
	return false
}

func (l *LogReader) Hist() *Hist {
	return l.cur
}

func (l *LogReader) Err() error {
	return l.err
}
