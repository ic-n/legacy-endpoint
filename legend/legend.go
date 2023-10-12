package legend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/ic-n/wait"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type Collector interface {
	Prepare(r *http.Request) (requestID string, err error)
	Collect(r *http.Request) (fragment json.RawMessage, err error)
}

type collectorConfig struct {
	collector Collector
	validity  time.Duration
}

type Controller struct {
	collectors []collectorConfig
}

func New() *Controller {
	return &Controller{
		collectors: make([]collectorConfig, 0),
	}
}

func (ctrl *Controller) Add(validity time.Duration, c Collector) {
	ctrl.collectors = append(ctrl.collectors, collectorConfig{
		collector: c,
		validity:  validity,
	})
}

func (ctrl *Controller) Handler() http.Handler {
	ji := jsoniter.ConfigFastest
	caches := make(map[int]*bigcache.BigCache)
	for i, cc := range ctrl.collectors {
		if cc.validity == time.Duration(0) {
			caches[i] = nil
			continue
		}
		cache, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(cc.validity))
		caches[i] = cache
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g := wait.WithContext[json.RawMessage](r.Context())
		for i := range ctrl.collectors {
			g.Go(pullFragment(ctrl.collectors[i], caches[i], r))
		}

		var document bytes.Buffer

		fs := ji.BorrowStream(&document)
		fs.WriteObjectStart()

		var comma atomic.Bool

		if err := g.Gather(func(f json.RawMessage) {
			fi := ji.BorrowIterator(f).ReadAny()
			keys := fi.Keys()
			for _, k := range keys {
				if comma.Swap(true) {
					fs.WriteMore()
				}

				fs.WriteObjectField(k)
				fi.Get(k).WriteTo(fs)
			}
		}); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fs.WriteObjectEnd()
		fs.Flush()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = document.WriteTo(w)
	})
}

func pullFragment(c collectorConfig, cache *bigcache.BigCache, r *http.Request) func(ctx context.Context) (json.RawMessage, error) {
	return func(ctx context.Context) (json.RawMessage, error) {
		requestID, err := c.collector.Prepare(r)
		if err != nil {
			return nil, errors.Wrap(err, "failed to prepare for request")
		}

		if cache != nil {
			f, err := cache.Get(requestID)
			if !errors.Is(err, bigcache.ErrEntryNotFound) {
				return f, nil
			}
		}

		fragment, err := c.collector.Collect(r)
		if err != nil {
			return nil, errors.Wrap(err, "failed to collect fragment for request")
		}

		if cache != nil {
			_ = cache.Set(requestID, fragment)
		}

		return fragment, nil
	}
}
