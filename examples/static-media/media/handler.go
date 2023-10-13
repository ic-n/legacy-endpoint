package media

import (
	"context"
	"net/http"
	"time"

	"github.com/ic-n/legacy-endpoint/examples/static-media/media/collectors"
	"github.com/ic-n/legacy-endpoint/legend"
	"gorm.io/gorm"
)

type Config struct {
	Database          *gorm.DB
	Client            *http.Client
	SupportedVersions []string
	FlowersURL        string
}

func Handler(config Config) http.Handler {
	media := legend.New()

	versions := make(map[string]struct{})
	for _, v := range config.SupportedVersions {
		versions[v] = struct{}{}
	}
	media.Add(0, &collectors.Version{
		Database:      config.Database,
		ValidVersions: versions,
	})

	media.Add(time.Second, &collectors.FlowerCDN{
		Client: config.Client,
		URL:    config.FlowersURL,
	})

	h := media.Handler()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		claims, err := getUserData(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx = context.WithValue(ctx, userDataKey, claims)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
