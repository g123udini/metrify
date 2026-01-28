package agent

import (
	"math"
	"strings"
	"testing"
)

func TestCollectGauge_HasExpectedKeysAndFiniteValues(t *testing.T) {
	g := CollectGauge()

	required := []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue",
	}

	for _, k := range required {
		v, ok := g[k]
		if !ok {
			t.Fatalf("missing key: %s", k)
		}
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("key %s has non-finite value: %v", k, v)
		}
	}

	if len(g) != len(required) {
		t.Fatalf("len=%d want=%d", len(g), len(required))
	}

	if g["RandomValue"] < 0 || g["RandomValue"] >= 1 {
		t.Fatalf("RandomValue=%v want [0,1)", g["RandomValue"])
	}
}

func TestCollectGopsutilGauges_KeysAndRanges(t *testing.T) {
	g := CollectGopsutilGauges()

	for k, v := range g {
		if strings.HasPrefix(k, "CPUutilization") {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Fatalf("%s is non-finite: %v", k, v)
			}
			if v < 0 || v > 100 {
				t.Fatalf("%s=%v out of range [0..100]", k, v)
			}
		}

		if k == "TotalMemory" || k == "FreeMemory" {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Fatalf("%s is non-finite: %v", k, v)
			}
			if v < 0 {
				t.Fatalf("%s=%v must be >= 0", k, v)
			}
		}
	}

	if tm, ok := g["TotalMemory"]; ok && tm == 0 {
		t.Fatalf("TotalMemory is present but 0")
	}
	if fm, ok := g["FreeMemory"]; ok && fm < 0 {
		t.Fatalf("FreeMemory=%v must be >=0", fm)
	}
	if tm, ok := g["TotalMemory"]; ok {
		if fm, ok2 := g["FreeMemory"]; ok2 && fm > tm {
			t.Fatalf("FreeMemory=%v > TotalMemory=%v", fm, tm)
		}
	}
}
