package hdrhist

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
)

// A LogWriter writes Hist's to a log file.
//
//     · A log file may optionally have a start time.
//     · A log file may optionally have a base time,
//       which will be used to reduce the space consumption of timestamps.
//     · A log file may have comments interleaved with histograms.
//
// Typical usage of a LogWriter would be as follows:
//     now := time.Now()
//     lw := NewLogWriter(w)
//     lw.WriteStartTime(now)
//     lw.SetBaseTime(now)
//     lw.WriteLegend()
//
// LogWriter is unfortunately difficult to use and was taken
// almost directly from the Java pacakge.
// This API should be revisited.
type LogWriter struct {
	w        io.Writer
	buf      bytes.Buffer
	baseTime *time.Time
}

func NewLogWriter(w io.Writer) *LogWriter {
	return &LogWriter{w: w}
}

func (l *LogWriter) WriteStartTime(start time.Time) error {
	const JavaDate = "Mon Jan 02 15:04:05 MST 2006"

	sec := start.Unix()
	millis := float64(start.Nanosecond()) / 1e6 // Java version only stores ms

	_, err := fmt.Fprintf(l.w, "#[StartTime: %.3f (seconds since epoch), %s]\n",
		float64(sec)+millis, start.Format(JavaDate))
	return errors.Wrap(err, "unable to write start time")
}

func (l *LogWriter) WriteBaseTime(base time.Time) error {
	sec := base.Unix()
	millis := float64(base.Nanosecond()) / 1e6 // Java version only stores ms
	_, err := fmt.Fprintf(l.w, "#[BaseTime: %.3f (seconds since epoch)]\n", float64(sec)+millis)
	return errors.Wrap(err, "unable to write base time")
}

func (l *LogWriter) WriteComment(text string) error {
	_, err := l.w.Write([]byte("#" + text + "\n"))
	return errors.Wrapf(err, "unable to write comment")
}

var logWriterLegend = []byte("\"StartTimestamp\",\"Interval_Length\",\"Interval_Max\",\"Interval_Compressed_Histogram\"\n")

func (l *LogWriter) WriteLegend() error {
	_, err := l.w.Write(logWriterLegend)
	return err
}

func (l *LogWriter) SetBaseTime(base time.Time) {
	l.baseTime = &base
}

func (l *LogWriter) GetBaseTime() (time.Time, bool) {
	if l.baseTime == nil {
		return time.Time{}, false
	}
	return *l.baseTime, true
}

func (l *LogWriter) WriteIntervalHist(h *Hist) error {
	t, ok := h.StartTime()
	e, okEnd := h.EndTime()
	if ok && okEnd {
		if b, ok := l.GetBaseTime(); ok {
			d := t.Sub(b)
			t = time.Unix(int64(d/time.Second), int64(d%time.Second))
			d = e.Sub(b)
			e = time.Unix(int64(d/time.Second), int64(d%time.Second))
		}
	}
	return l.writeHist(h, t, e)
}

func (l *LogWriter) writeHist(h *Hist, start time.Time, end time.Time) error {
	const MaxValueUnitRatio = 1000000.0
	l.buf.Reset()
	max := h.Max()
	fmt.Fprintf(&l.buf, "%.3f,%.3f,%.3f,",
		float64(start.Unix())+(float64(start.Nanosecond()/1e6)/1e3),
		float64(end.Sub(start)/time.Millisecond)/1e3,
		float64(max)/MaxValueUnitRatio)
	b64w := base64.NewEncoder(base64.StdEncoding, &l.buf)
	encodeCompressed(h, b64w, max) // not writing to disk yet, won't fail
	b64w.Close()
	l.buf.WriteString("\n")
	_, err := l.buf.WriteTo(l.w)
	return errors.Wrap(err, "unable to write hist")
}
