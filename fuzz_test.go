package hdrhist_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/uluyol/hdrhist/internal/fuzz"
)

var corpusDir = filepath.Join("testdata", "fuzz", "corpus")

func TestFuzzLogReader(t *testing.T) {
	corpus, err := ioutil.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("unable to open corpus: %v", err)
	}
	for _, fi := range corpus {
		d, err := ioutil.ReadFile(filepath.Join(corpusDir, fi.Name()))
		if err != nil {
			t.Fatalf("unable to read case %s: %v", fi.Name(), err)
		}
		_ = fuzz.Fuzz(d)
	}
}
