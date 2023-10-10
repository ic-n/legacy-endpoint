package legend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type CollectorConfig struct {
	Validity time.Duration
	Prepare  func(r *http.Request) (requestID string, err error)
	Collect  func(r *http.Request) (fragment map[string]any, err error)
}

type Controller struct {
	Collectors []CollectorConfig
	ErrHandler func(error)
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

		for i, cc := range ctrl.Collectors {
			requestID, err := cc.Prepare(r)
			if err != nil {
				if ctrl.ErrHandler != nil {
					ctrl.ErrHandler(errors.Wrap(err, "failed to prepare for request"))
				}
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			memKey := fmt.Sprintf("%d:%s", i, requestID)
			c, ok := mem[memKey]
			if ok && time.Until(c.lastRefresh) < cc.Validity {
				for k, v := range c.fragment {
					document[k] = v
				}
				continue
			}

			fragment, err := cc.Collect(r)
			if err != nil {
				if ctrl.ErrHandler != nil {
					ctrl.ErrHandler(errors.Wrap(err, "failed to collect fragment for request"))
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			mem[memKey] = collector{
				lastRefresh: time.Now(),
				fragment:    fragment,
			}

			for k, v := range fragment {
				document[k] = v
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		err := json.NewEncoder(w).Encode(document)
		if err != nil {
			if ctrl.ErrHandler != nil {
				ctrl.ErrHandler(errors.Wrap(err, "failed to write response"))
			}
			return
		}
	})
}
