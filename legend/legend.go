package legend

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/ic-n/wait"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type CollectorConfig struct {
	Validity time.Duration
	Prepare  func(r *http.Request) (requestID string, err error)
	Collect  func(r *http.Request) (fragment json.RawMessage, err error)
}

type Controller struct {
	Collectors []CollectorConfig
}

func New() *Controller {
	return &Controller{
		Collectors: make([]CollectorConfig, 0),
	}
}

func (ctrl Controller) Handle() http.Handler {
	ji := jsoniter.ConfigFastest
	caches := make(map[int]*bigcache.BigCache)
	for i, cc := range ctrl.Collectors {
		if cc.Validity == time.Duration(0) {
			caches[i] = nil
			continue
		}
		cache, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(cc.Validity))
		caches[i] = cache
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g := wait.WithContext[json.RawMessage](r.Context())
		for i := range ctrl.Collectors {
			g.Go(pullFragment(ctrl.Collectors[i], caches[i], r))
		}

		var document bytes.Buffer

		fs := ji.BorrowStream(&document)
		fs.WriteObjectStart()

		if err := g.Gather(func(f json.RawMessage) {
			fi := ji.BorrowIterator(f).ReadAny()
			keys := fi.Keys()
			for i, k := range keys {
				fs.WriteObjectField(k)
				fi.Get(k).WriteTo(fs)
				if i != len(keys)-1 {
					fs.WriteMore()
				}
			}
		}); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fs.WriteObjectEnd()
		fs.Flush()
		document.WriteRune('\n')

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = document.WriteTo(w)
	})
}

func pullFragment(c CollectorConfig, cache *bigcache.BigCache, r *http.Request) func(ctx context.Context) (json.RawMessage, error) {
	return func(ctx context.Context) (json.RawMessage, error) {
		requestID, err := c.Prepare(r)
		if err != nil {
			return nil, errors.Wrap(err, "failed to prepare for request")
		}

		if cache != nil {
			f, err := cache.Get(requestID)
			if !errors.Is(err, bigcache.ErrEntryNotFound) {
				return f, nil
			}
		}

		fragment, err := c.Collect(r)
		if err != nil {
			return nil, errors.Wrap(err, "failed to collect fragment for request")
		}

		if cache != nil {
			_ = cache.Set(requestID, fragment)
		}

		return fragment, nil
	}
}
