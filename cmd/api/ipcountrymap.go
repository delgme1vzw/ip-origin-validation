package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/delgme1vzw/ip-origin-validation/internal/datastore"
	"github.com/go-chi/chi"
)

// Use a struct for easier validation and json unmarshall/marshalling
// This is what the user will send us
type IPCountriesPayload struct {
	IP        string   `json:"ip" validate:"required" example:"2001:200::"`
	Countries []string `json:"countries" validate:"required" example:"[\"Germany\", \"France\"]"`
}

type CIDRIPPayload struct {
	CIDRIP string `json:"ip" validate:"required" example:"2001:200::/32"`
}

type CIDRIPCountryPayload struct {
	CIDRIP  string `json:"ip" validate:"required" example:"2001:200::/32"`
	Country string `json:"country" validate:"required" example:"Japan"`
}

// @Summary Allows the user to check if a CIDR IP is whitelisted for the given countries
// @Tags Whitelist
// @Accept json
// @Produce json
// @Param ip path string true "Single IP (IPv4 or IPv6)" example(2001:200::) default(2001:200::)
// @Param countries query string true "Comma-separated countries" example(France,Uganda) default(France,Uganda)
// @Success 200 {object} datastore.WhitelistedResults
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /v1/whitelist-status/{ip} [get]
// payload should have the ip and a list of country codes
func (app *application) getWhitelistStatusHandler(w http.ResponseWriter, r *http.Request) {
	rawIp := chi.URLParam(r, "ip")
	ip, err := url.PathUnescape(rawIp)

	countriesParam := r.URL.Query().Get("countries")

	if len(ip) <= 0 || len(countriesParam) <= 0 || err != nil {
		app.badRequestResponse(w, r, datastore.ErrInvalidRequest)
		return
	}

	countries := strings.Split(countriesParam, ",")

	ctx := r.Context()
	cc, err := app.store.IPCountryMap.GetCountryByIP(ctx, ip, false)

	if err != nil {
		switch {
		case errors.Is(err, datastore.ErrNotFound):
			app.notFoundResponse(w, r, err)
		case errors.Is(err, datastore.ErrInvalidRequest):
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	var wl = false
	for _, v := range countries {
		if v == cc.Country {
			wl = true
		}
	}

	whitelist := &datastore.WhitelistedResults{
		Whitelisted: wl,
	}

	if err := writeJSON(w, http.StatusOK, whitelist); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary Allows the user to check what country a CIDR IP is mapped to
// @Tags Maintenance
// @Accept json
// @Produce json
// @Param ip path string true "CIDR IP (IPv4 or IPv6)" example(2001:200::/32) default(2001:200::/32)
// @Success 200 {object} datastore.MappedCountry
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /v1/ip-map/{ip} [get]
// payload should have a single ip only
// we want to return the country code
func (app *application) getCountryByIPHandler(w http.ResponseWriter, r *http.Request) {
	rawIp := chi.URLParam(r, "ip")
	ip, err := url.PathUnescape(rawIp)

	if len(ip) <= 0 || err != nil {
		app.badRequestResponse(w, r, datastore.ErrInvalidRequest)
		return
	}

	ctx := r.Context()
	cc, err := app.store.IPCountryMap.GetCountryByIP(ctx, ip, true)

	if err != nil {
		switch {
		case errors.Is(err, datastore.ErrNotFound):
			app.notFoundResponse(w, r, err)
		case errors.Is(err, datastore.ErrInvalidRequest):
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	mappedCountry := &datastore.MappedCountry{
		Country: cc.Country,
	}

	if err := writeJSON(w, http.StatusOK, mappedCountry); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary Allows the user to create a new entry in CIDR IP - Country table
// @Tags Maintenance
// @Accept json
// @Produce json
// @Param request body CIDRIPCountryPayload true "IP (CIDR IPv4 or IPv6) and Country"
// @Success 201 {object} datastore.MappedIpCountryRecord
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /v1/ip-map/ [post]
func (app *application) getCreateMappedCountryHandler(w http.ResponseWriter, r *http.Request) {
	var payload CIDRIPCountryPayload

	//See if the payload can be decoded
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//If so, see if the payload can be validated based on struct json tags
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	mappedIpCountryRecord := &datastore.MappedIpCountryRecord{
		IP:      payload.CIDRIP,
		Country: payload.Country,
	}

	ctx := r.Context()

	if err := app.store.IPCountryMap.Create(ctx, mappedIpCountryRecord); err != nil {
		switch {
		case errors.Is(err, datastore.ErrNotFound):
			app.notFoundResponse(w, r, err)
		case errors.Is(err, datastore.ErrInvalidRequest):
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, mappedIpCountryRecord); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary Allows the user to update an entry in CIDR IP - Country table by CIDR IP
// @Tags Maintenance
// @Accept json
// @Produce json
// @Param request body CIDRIPCountryPayload true "IP (CIDR IPv4 or IPv6) and Country"
// @Success 200 {object} datastore.MappedIpCountryRecord
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /v1/ip-map/ [patch]
// Technically this accepts by PUT too,
// but is theoretically a PATCH type since it only updates the Country
func (app *application) getUpdateMappedCountryHandler(w http.ResponseWriter, r *http.Request) {
	var payload CIDRIPCountryPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	mappedIpCountryRecord := &datastore.MappedIpCountryRecord{
		IP:      payload.CIDRIP,
		Country: payload.Country,
	}

	ctx := r.Context()

	if err := app.store.IPCountryMap.Update(ctx, mappedIpCountryRecord); err != nil {
		switch {
		case errors.Is(err, datastore.ErrNotFound):
			app.notFoundResponse(w, r, err)
		case errors.Is(err, datastore.ErrInvalidRequest):
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, mappedIpCountryRecord); err != nil {
		app.internalServerError(w, r, err)
	}
}

// @Summary Allows the user to delete an entry in CIDR IP - Country table by the CIDR IP
// @Tags Maintenance
// @Accept json
// @Produce json
// @Param request body CIDRIPPayload true "IP (CIDR IPv4 or IPv6)"
// @Success 200 "Successfully deleted"
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /v1/ip-map/ [delete]
func (app *application) getDeleteMappedCountryHandler(w http.ResponseWriter, r *http.Request) {
	var payload CIDRIPPayload

	//See if the payload can be decoded
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//If so, see if the payload can be validated based on struct json tags
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.IPCountryMap.Delete(ctx, payload.CIDRIP); err != nil {
		switch {
		case errors.Is(err, datastore.ErrNotFound):
			app.notFoundResponse(w, r, err)
		case errors.Is(err, datastore.ErrInvalidRequest):
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
