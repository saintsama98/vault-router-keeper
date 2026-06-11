package brain

import (
	"math"
	"testing"
)

func approx(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestNormCDF(t *testing.T) {
	cases := []struct{ x, want float64 }{
		{0, 0.5},
		{-1, 0.1587},
		{1, 0.8413},
	}
	for _, c := range cases {
		if got := normCDF(c.x); !approx(got, c.want, 0.001) {
			t.Errorf("normCDF(%v) = %v, want %v", c.x, got, c.want)
		}
	}
}

func TestPLiq(t *testing.T) {
	if got := pLiq(0.9, 1.0, 1.0/52); !approx(got, 0.447, 0.01) {
		t.Errorf("pLiq(0.9,1.0,1/52) = %v, want ~0.447", got)
	}
	if got := pLiq(0.5, 0.3, 1.0/52); !approx(got, 0, 0.001) {
		t.Errorf("pLiq(0.5,0.3,1/52) = %v, want ~0", got)
	}
	if got := pLiq(1.0, 0.3, 1.0/52); got != 1 {
		t.Errorf("pLiq(1.0,...) = %v, want 1", got)
	}
}

func TestClosedFormEL(t *testing.T) {
	m := DefaultRiskModel()
	el := m.closedFormEL(RiskInputs{Sigma: 1.0, HorizonYears: 1.0 / 52, LiqPriceRatio: 0.9, LiquidityHaircut: 0.8})
	if el < 0.12 || el > 0.14 {
		t.Errorf("closedFormEL = %v, want 0.12–0.14 (≈0.131)", el)
	}
}

func TestEffectiveCapBps(t *testing.T) {
	m := DefaultRiskModel()
	cases := []struct {
		el   float64
		want uint16
	}{
		{0.01, 5000}, // below soft → full owner cap
		{0.06, 2500}, // midway → half
		{0.10, 0},    // at kill → zero
	}
	for _, c := range cases {
		if got := m.effectiveCapBps(5000, c.el); uint16(got) != c.want {
			t.Errorf("effectiveCapBps(5000,%v) = %v, want %v", c.el, got, c.want)
		}
	}
}

func TestExpectedLossVendorDominates(t *testing.T) {
	m := DefaultRiskModel()
	// Sigma 0 → closed-form EL 0, so the vendor EL must dominate via max().
	if got := m.expectedLoss(RiskInputs{Sigma: 0, HasVendor: true, VendorEL: 0.2}); !approx(got, 0.2, 1e-9) {
		t.Errorf("expectedLoss vendor = %v, want 0.2", got)
	}
}
