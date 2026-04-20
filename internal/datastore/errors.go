package datastore

import (
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

var (
	ErrNotFound          = errors.New("record not found")
	ErrInvalidRequest    = errors.New("invalid request")
	QueryTimeoutDuration = time.Second * 5
)

func checkDbErrors(err error) error {

	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return fmt.Errorf(
			"db error: [code=%s] %s",
			pgErr.Code,
			pgErr.Message,
		)
	}
	return err

}
