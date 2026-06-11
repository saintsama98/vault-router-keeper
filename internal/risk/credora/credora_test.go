package credora

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/vault-router-keeper/pkg/types"
)

// fakeDoer returns a canned HTTP response, standing in for the live endpoint.
type fakeDoer struct {
	body   string
	status int
	err    error
}

func (f fakeDoer) Do(*http.Request) (*http.Response, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &http.Response{
		StatusCode: f.status,
		Body:       io.NopCloser(strings.NewReader(f.body)),
	}, nil
}

func mid(b byte) types.StrategyID {
	var id types.StrategyID
	id[0] = b
	return id
}

func TestReaderVendorEL(t *testing.T) {
	id := mid(1)
	markets := map[types.StrategyID]string{id: "morpho-usdc"}

	cases := []struct {
		name   string
		rating string
		wantEL float64
	}{
		{"A best", "A", 0.00},
		{"B", "B", 0.03},
		{"C", "C", 0.08},
		{"D worst", "D", 0.15},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body := `{"data":{"market":{"credora":{"rating":"` + c.rating + `"}}}}`
			r := NewReader("http://credora.test/graphql", "", fakeDoer{body: body, status: 200}, markets)
			ri, ok := r.Risk(id)
			if !ok {
				t.Fatalf("rating %q: ok=false, want true", c.rating)
			}
			if !ri.HasVendor {
				t.Errorf("rating %q: HasVendor=false, want true", c.rating)
			}
			if ri.VendorEL != c.wantEL {
				t.Errorf("rating %q: VendorEL=%v, want %v", c.rating, ri.VendorEL, c.wantEL)
			}
		})
	}
}

func TestReaderDegradesSafely(t *testing.T) {
	id := mid(1)
	markets := map[types.StrategyID]string{id: "morpho-usdc"}

	cases := []struct {
		name string
		doer fakeDoer
	}{
		{"http error", fakeDoer{err: io.ErrUnexpectedEOF}},
		{"non-200", fakeDoer{body: "{}", status: 500}},
		{"empty rating", fakeDoer{body: `{"data":{"market":{"credora":{"rating":""}}}}`, status: 200}},
		{"unknown rating", fakeDoer{body: `{"data":{"market":{"credora":{"rating":"Z"}}}}`, status: 200}},
		{"garbage json", fakeDoer{body: "not json", status: 200}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := NewReader("http://credora.test/graphql", "", c.doer, markets)
			if _, ok := r.Risk(id); ok {
				t.Errorf("%s: ok=true, want false (safe degradation)", c.name)
			}
		})
	}
}

func TestReaderUnmappedStrategy(t *testing.T) {
	// A strategy with no market mapping must report no data, regardless of feed.
	r := NewReader("http://credora.test/graphql", "", fakeDoer{body: `{}`, status: 200}, nil)
	if _, ok := r.Risk(mid(9)); ok {
		t.Errorf("unmapped strategy: ok=true, want false")
	}
}
