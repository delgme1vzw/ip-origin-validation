package datastore

import (
	"context"
	"database/sql"
)

// Create the storage interface so testing or changing the storage
// does not require new signatures
type Storage struct {
	IPCountryMap interface {
		GetCountryByIP(context.Context, string, bool) (*MappedCountry, error)
		Delete(context.Context, string) error
		Create(context.Context, *MappedIpCountryRecord) error
		Update(context.Context, *MappedIpCountryRecord) error
	}
}

// Keep flexible repository so we can add stores in the futre
// if needed, like authentication storage
// Probably overkill for this POC
func NewStorage(db *sql.DB) Storage {
	return Storage{
		IPCountryMap: &IPCountryMapStore{db},
	}
}
