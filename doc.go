/*
Package hdrhist provides high dynamic range (HDR) histograms.

HDR histograms can be used to accurately record and analyze
distributions with very large ranges of data.
HDR histograms are compact: with a precision of 3 significant digits
and a range of 1000-100 billion (e.g. 1 μs to 100 s),
an HDR histogram consumes about 156 KB.

A typical usecase for HDR Histograms would be recording latency values
in client or server software.

This package is a Go port of Gil Tene's HdrHistogram Java package
(http://hdrhistogram.github.io/HdrHistogram/).
Package hdrhist aims to interoperate with HdrHistogram
to the fullest extent possible.

Unless otherwise noted, none of the types in the package are safe for concurrent use.

*/
package hdrhist
