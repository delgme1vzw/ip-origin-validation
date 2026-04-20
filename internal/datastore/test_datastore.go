package datastore

import "context"

func NewDummyStorage() Storage {
	return Storage{
		IPCountryMap: &DummyIPCountryMap{},
	}
}

type DummyIPCountryMap struct {
}

func (m *DummyIPCountryMap) GetCountryByIP(context.Context, string, bool) (*MappedCountry, error) {
	return nil, nil
}

func (m *DummyIPCountryMap) Delete(context.Context, string) error {
	return nil
}

func (m *DummyIPCountryMap) Create(context.Context, *MappedIpCountryRecord) error {
	return nil
}

func (m *DummyIPCountryMap) Update(context.Context, *MappedIpCountryRecord) error {
	return nil
}
