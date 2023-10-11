package legend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ic-n/wait"
	"github.com/pkg/errors"
)

type CollectorConfig struct {
	Validity time.Duration
	Prepare  func(r *http.Request) (requestID string, err error)
	Collect  func(r *http.Request) (fragment map[string]any, err error)
}

type Controller struct {
	Collectors []CollectorConfig
}

func New() *Controller {
	return &Controller{
		Collectors: make([]CollectorConfig, 0),
	}
}

type collector struct {
	lastRefresh time.Time
	fragment    map[string]any
}

func (ctrl Controller) Handle() http.Handler {
	mem := make(map[string]collector)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		document := make(map[string]any)

		g := wait.WithContext[map[string]any](r.Context())
		for i := range ctrl.Collectors {
			i := i
			cc := ctrl.Collectors[i]

			g.Go(func(ctx context.Context) (map[string]any, error) {
				requestID, err := cc.Prepare(r)
				if err != nil {
					return nil, errors.Wrap(err, "failed to prepare for request")
				}

				memKey := fmt.Sprintf("%d:%s", i, requestID)
				c, ok := mem[memKey]
				if ok && time.Until(c.lastRefresh) < cc.Validity {
					return c.fragment, nil
				}

				fragment, err := cc.Collect(r)
				if err != nil {
					return nil, errors.Wrap(err, "failed to collect fragment for request")
				}

				mem[memKey] = collector{
					lastRefresh: time.Now(),
					fragment:    fragment,
				}

				return fragment, nil
			})
		}

		if err := g.Gather(func(f map[string]any) {
			for k, v := range f {
				document[k] = v
			}
		}); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(document)
	})
}
