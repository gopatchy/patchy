package patchy

import (
	"net/http"

	"github.com/gopatchy/jsrest"
)

func (api *API) put(cfg *config, id string, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	api.SetEventData(ctx,
		"operation", "replace",
		"typeName", cfg.apiName,
		"id", id,
	)

	replace := cfg.factory()
	opts := parseUpdateOpts(r)

	err := jsrest.Read(r, replace)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "read request failed (%w)", err)
	}

	replace, err = api.replaceInt(ctx, cfg, id, replace, opts)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "replace failed (%w)", err)
	}

	err = jsrest.Write(w, replace)
	if err != nil {
		return jsrest.Errorf(jsrest.ErrInternalServerError, "write response failed (%w)", err)
	}

	return nil
}
