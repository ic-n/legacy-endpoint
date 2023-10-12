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

func TestHandler(t *testing.T) {
	le := legend.New()

	var counter atomic.Int64

	le.Collectors = append(le.Collectors, legend.CollectorConfig{
		Validity: time.Second,
		Prepare: func(r *http.Request) (requestID string, err error) {
			return r.Header.Get("Authorization"), nil
		},
		Collect: func(r *http.Request) (json.RawMessage, error) {
			time.Sleep(time.Second)

			return json.RawMessage(fmt.Sprintf(`{"test":%d}`, counter.Add(1))), nil
		},
	})
	h := le.Handle()

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		r.Header.Add("Authorization", "Bearer test123")
		h.ServeHTTP(w, r)

		require.Equal(t, "{\"test\":1}\n", w.Body.String())
	}

	require.Equal(t, int64(1), counter.Load())
}
