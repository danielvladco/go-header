package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/celestiaorg/go-header"
)

// Init ensures a Store is initialized. If it is not already initialized,
// it initializes the Store by requesting the header with the given hash.
func Init[H header.Header[H]](ctx context.Context, store header.Store[H], ex header.Exchange[H], hash header.Hash) error {
	fmt.Printf("DUMMY INIT HEADER TRUSTED HASH: %v \n", hash)
	_, err := store.Head(ctx)
	switch {
	default:
		fmt.Printf("DUMMY DEFAULT ERROR CASE: %s \n", err)
		return err
	case errors.Is(err, header.ErrNoHead):
		initial, err := ex.Get(ctx, hash)
		fmt.Printf("DUMMY HEADER EXCHANGE INIT HASH: %v - err: %s \n", initial, err)
		if err != nil {
			return err
		}

		a := store.Init(ctx, initial)
		fmt.Printf("DUMMY STORE INIT INFO: %+v \n", a)
		return a
	}
}
