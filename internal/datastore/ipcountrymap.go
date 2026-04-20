package datastore

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"
)

// This is what we will be retrieving from our query
type MappedCountry struct {
	Country string
}

type MappedIpCountryRecord struct {
	IP        string
	Country   string
	ID        int
	CreatedAt string
}

type WhitelistedResults struct {
	Whitelisted bool
}

// Using a struct to pass as our db when needing the whitelist map
type IPCountryMapStore struct {
	db *sql.DB
}

// This method is called either by the user passing in a single ip address and wanting back a country code (cidrLookup = false)
// Or a user passing in a cidr format address and wanting back a country code (cidrLookup true)
func (s *IPCountryMapStore) GetCountryByIP(ctx context.Context, ip string, cidrLookup bool) (*MappedCountry, error) {

	ipTable, err := getIpTable(ip)
	if err != nil {
		return nil, err
	}

	//If it is not a cidrLookup, it is a single ip address and we need to find where it lives amoungst the cidr mask
	//We handle the case where it may match multiple rows, but assume the user wants only one row back
	//If it is a cidrLookup, the ip is the cidr format and we can just do a search for that record directly
	var whereClause string
	if cidrLookup {
		whereClause = `$1 = r.cidr_ip`
	} else {
		whereClause = `$1::inet << r.cidr_ip`
	}

	query := `
	SELECT g.country_name
	FROM ` + ipTable + ` r
	JOIN geolocation_country_map g
	ON r.geoname_id = g.geoname_id
	WHERE ` + whereClause + `
	ORDER BY masklen(r.cidr_ip) DESC
	LIMIT 1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var mappedCountry MappedCountry
	err = s.db.QueryRowContext(
		ctx,
		query,
		ip).Scan(
		&mappedCountry.Country,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, checkDbErrors(err)
	}

	return &mappedCountry, nil
}

func (s *IPCountryMapStore) Delete(ctx context.Context, ip string) error {

	ipTable, err := getIpTable(ip)
	if err != nil {
		return err
	}

	query := `DELETE FROM ` + ipTable + ` WHERE cidr_ip = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, ip)
	if err != nil {
		return checkDbErrors(err)
	}

	if rows, err := res.RowsAffected(); err != nil {
		return err
	} else if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *IPCountryMapStore) Create(ctx context.Context, mappedIpCountryRecord *MappedIpCountryRecord) error {

	ipTable, ipFmtErr := getIpTable(mappedIpCountryRecord.IP)
	if ipFmtErr != nil {
		return ipFmtErr
	}

	query := `
		INSERT INTO ` + ipTable + ` (cidr_ip, geoname_id)
		SELECT $1, g.geoname_id
		FROM geolocation_country_map g
		WHERE g.country_name = $2
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	row := s.db.QueryRowContext(
		ctx,
		query,
		mappedIpCountryRecord.IP,
		mappedIpCountryRecord.Country,
	)
	err := row.Scan(
		&mappedIpCountryRecord.ID,
		&mappedIpCountryRecord.CreatedAt,
	)

	if err != nil {
		return checkDbErrors(err)
	}

	return nil
}

func (s *IPCountryMapStore) Update(ctx context.Context, mappedIpCountryRecord *MappedIpCountryRecord) error {

	ipTable, ipFmtErr := getIpTable(mappedIpCountryRecord.IP)
	if ipFmtErr != nil {
		return ipFmtErr
	}

	query := `
		UPDATE ` + ipTable + ` 
		SET geoname_id = (SELECT geoname_id from geolocation_country_map where country_name = $1)
		WHERE cidr_ip = $2
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	row := s.db.QueryRowContext(
		ctx,
		query,
		mappedIpCountryRecord.Country,
		mappedIpCountryRecord.IP,
	)
	err := row.Scan(
		&mappedIpCountryRecord.ID,
		&mappedIpCountryRecord.CreatedAt,
	)

	if err != nil {
		return checkDbErrors(err)
	}

	return nil
}

// Multiple datastore interfaces need this method so putting it here
// not enough other validation warrants a validate util file
/* Humorous side note:
I tried hard to make this work with regexp but handling the compressed ipv6 was getting out of hand
When I asked AI for a better ipv6 regex, it suggested the net.ParseIP method which was much more concise
Frying pans...who knew?
//regex to match ipv4
reIpv4 := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?).){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
//regex to match ipv6
reIpv6 := regexp.MustCompile(`^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$|^::1$|^::$`)
*/
func getIpVersion(ip string) int {
	//Check if this is CIDR notation
	//	If it is, see if it is good CIDR notation
	//		If it is, remove the mask
	if strings.Contains(ip, "/") {
		if ipa, _, err := net.ParseCIDR(ip); err == nil {
			ip = ipa.String()
		}
	}

	//Now we have a non-cidr format ip
	v := net.ParseIP(ip)

	if v == nil {
		return 0
	}
	if v.To4() != nil {
		return 4
	}
	return 6
}

// Return the table based on the IP
// If I had more constants to worry about I would move the tablename as a contant in the datastore dir
func getIpTable(ip string) (string, error) {
	var ipTable string

	switch getIpVersion(ip) {
	case 4:
		ipTable = "ipv4_geolocation_map"
	case 6:
		ipTable = "ipv6_geolocation_map"
	default:
		return "", ErrInvalidRequest
	}
	return ipTable, nil
}
