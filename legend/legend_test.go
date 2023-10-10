package legend_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ic-n/legacy-endpoint/legend"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	le := legend.New()

	le.Collectors = append(le.Collectors, legend.CollectorConfig{
		Prepare: func(r *http.Request) (requestID string, err error) {
			return r.Header.Get("Authorization"), nil
		},
		Collect: func(r *http.Request) (fragment map[string]any, err error) {
			return map[string]any{
				"test": 1,
			}, nil
		},
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	r.Header.Add("Authorization", "Bearer test123")
	le.Handle().ServeHTTP(w, r)

	require.Equal(t, "{\"test\":1}\n", w.Body.String())
}
