package legend_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ic-n/legacy-endpoint/legend"
	"github.com/stretchr/testify/require"
)

type MockCollector struct {
	counter atomic.Int64
	label   string
}

func (mc *MockCollector) Prepare(r *http.Request) (requestID string, err error) {
	return r.Header.Get("Authorization"), nil
}

func (mc *MockCollector) Collect(_ *http.Request) (fragment json.RawMessage, err error) {
	time.Sleep(100 * time.Millisecond)

	return json.RawMessage(fmt.Sprintf(`{"%s":%d}`, mc.label, mc.counter.Add(1))), nil
}

func TestHandler(t *testing.T) {
	le := legend.New()

	mc1 := &MockCollector{label: "test1"}
	le.Add(time.Second, mc1)
	mc2 := &MockCollector{label: "test2"}
	le.Add(0, mc2)

	h := le.Handler()

	for i := 1; i < 6; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		r.Header.Add("Authorization", "Bearer test123")
		h.ServeHTTP(w, r)

		require.JSONEq(t, fmt.Sprintf(`{"test1":1,"test2":%d}`, i), w.Body.String())
	}

	require.Equal(t, int64(1), mc1.counter.Load())
	require.Equal(t, int64(5), mc2.counter.Load())
}
