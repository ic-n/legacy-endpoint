package media_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/ic-n/legacy-endpoint/examples/static-media/media"
	"github.com/ic-n/legacy-endpoint/examples/static-media/store"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var tokenString = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJpc3MiOiJ0ZXN0IiwiYXVkIjoic2luZ2xlIn0.QAWg1vGvnqRuCFTMcPkjZljXHh8U3L_qUjszOtQbeaA"

func TestMedia(t *testing.T) {
	db := testDatabase(t)

	flowerServer := httptest.NewServer(mockFlowers())
	defer func() { flowerServer.Close() }()

	c := media.Config{
		Database:          db,
		Client:            http.DefaultClient,
		SupportedVersions: []string{"api/v1/", "api/v2/"},
		FlowersURL:        flowerServer.URL,
	}
	h := media.Handler(c)

	t.Run("no_auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/media", http.NoBody)
		h.ServeHTTP(w, r)

		require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("invalid_version", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/media", http.NoBody)
		r.Header.Add("Authorization", "Bearer "+tokenString)
		h.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	})

	t.Run("unsupported_version", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v4/media", http.NoBody)
		r.Header.Add("Authorization", "Bearer "+tokenString)
		h.ServeHTTP(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	})

	t.Run("requst_v1", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/media", http.NoBody)
		r.Header.Add("Authorization", "Bearer "+tokenString)
		h.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.JSONEq(t, `{
			"privacy_document": "<h1>Privacy</h1>",
			"user_agreement": "...",
			"flowers": [
				"https://cdn.ic-n.uk/flowers/Daffodil.png",
				"https://cdn.ic-n.uk/flowers/Jasmine.png"
			]
		}`, w.Body.String())
	})

	t.Run("requst_v2", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v2/media", http.NoBody)
		r.Header.Add("Authorization", "Bearer "+tokenString)
		h.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.JSONEq(t, `{
			"privacy_document": "<h1>No privacy</h1>",
			"user_agreement": "...",
			"flowers": [
				"https://cdn.ic-n.uk/flowers/Daffodil.png",
				"https://cdn.ic-n.uk/flowers/Jasmine.png"
			]
		}`, w.Body.String())
	})
}

func mockFlowers() http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(`[
			"https://cdn.ic-n.uk/flowers/Daffodil.png",
			"https://cdn.ic-n.uk/flowers/Jasmine.png"
		]`))
	})
}

func testDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&store.Agreement{})
	require.NoError(t, err)
	db.Create([]store.Agreement{
		{
			Version:     "api/v1/",
			EULA:        "...",
			PrivacyHTML: "<h1>Privacy</h1>",
		},
		{
			Version:     "api/v2/",
			EULA:        "...",
			PrivacyHTML: "<h1>No privacy</h1>",
		},
	})
	return db
}
