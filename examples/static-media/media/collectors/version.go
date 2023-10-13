package collectors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/ic-n/legacy-endpoint/examples/static-media/store"
	"gorm.io/gorm"
)

var ErrUnsuportedVersion = errors.New("unsuported api version")

const apiVersionHeader = "Api-Version"

var apiVersionRegexp = regexp.MustCompile(`api/v\d+[\.\d]*/`)

type Version struct {
	ValidVersions map[string]struct{}
	Database      *gorm.DB
}

func (c *Version) Prepare(r *http.Request) (string, error) {
	apiVersion := apiVersionRegexp.FindString(r.RequestURI)

	if _, ok := c.ValidVersions[apiVersion]; ok {
		r.Header.Add(apiVersionHeader, apiVersion)

		return apiVersion, nil
	}

	return "", fmt.Errorf("%w: %s", ErrUnsuportedVersion, apiVersion)
}

type VersionFragment struct {
	EULA        string `json:"user_agreement"`
	PrivacyHTML string `json:"privacy_document"`
}

func (c *Version) Collect(r *http.Request) (json.RawMessage, error) {
	var a store.Agreement

	q := c.Database.Where("version = ?", r.Header.Get(apiVersionHeader)).First(&a)
	if q.Error != nil {
		return nil, q.Error
	}

	return json.Marshal(VersionFragment{EULA: a.EULA, PrivacyHTML: a.PrivacyHTML})
}
