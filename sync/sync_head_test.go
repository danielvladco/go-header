package sync

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ipfs/go-datastore"
	sync2 "github.com/ipfs/go-datastore/sync"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/go-header"
	"github.com/celestiaorg/go-header/headertest"
	"github.com/celestiaorg/go-header/local"
	"github.com/celestiaorg/go-header/store"
)

func TestSyncer_incomingNetworkHeadRaces(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	t.Cleanup(cancel)

	suite := headertest.NewTestSuite(t)

	store := headertest.NewStore[*headertest.DummyHeader](t, suite, 1)
	syncer, err := NewSyncer[*headertest.DummyHeader](
		store,
		store,
		headertest.NewDummySubscriber(),
	)
	require.NoError(t, err)

	incoming := suite.NextHeader()

	var hits atomic.Uint32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if syncer.incomingNetworkHead(ctx, incoming) == pubsub.ValidationAccept {
				hits.Add(1)
			}
		}()
	}

	wg.Wait()
	assert.EqualValues(t, 1, hits.Load())

}

// TestSyncer_HeadWithDisabledSubjectiveInit tests whether the syncer
// requests Head (new sync target) with subjective initialisation disabled when
// it already has a subjective head within the unbonding period.
func TestSyncer_HeadWithDisabledSubjectiveInit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	t.Cleanup(cancel)

	suite := headertest.NewTestSuite(t)
	head := suite.Head()

	localStore := store.NewTestStore(ctx, t, head)

	remoteStore, err := store.NewStoreWithHead(ctx, sync2.MutexWrap(datastore.NewMapDatastore()), head)
	require.NoError(t, err)
	err = remoteStore.Append(ctx, suite.GenDummyHeaders(100)...)
	require.NoError(t, err)

	// create a wrappedGetter to track exchange interactions
	wrappedGetter := newWrappedGetter(local.NewExchange[*headertest.DummyHeader](remoteStore))

	syncer, err := NewSyncer[*headertest.DummyHeader](
		wrappedGetter,
		localStore,
		headertest.NewDummySubscriber(),
		// forces a request for a new sync target
		WithBlockTime(time.Nanosecond),
		// ensures that syncer's store contains a subjective head that is within
		// the unbonding period so that the syncer can use a header from the network
		// as a sync target
		WithTrustingPeriod(time.Hour),
	)
	require.NoError(t, err)

	// start the syncer which triggers a Head request that will
	// load the syncer's subjective head from the store, and request
	// a new sync target from the network rather than from trusted peers
	err = syncer.Start(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = syncer.Stop(ctx)
		require.NoError(t, err)
	})

	// ensure the syncer really requested Head from the network
	// rather than from trusted peers
	require.True(t, wrappedGetter.disabledSubjectiveInit)
}

type wrappedGetter struct {
	ex header.Exchange[*headertest.DummyHeader]

	disabledSubjectiveInit bool
}

func newWrappedGetter(ex header.Exchange[*headertest.DummyHeader]) *wrappedGetter {
	return &wrappedGetter{
		ex:                     ex,
		disabledSubjectiveInit: false,
	}
}

func (t *wrappedGetter) Head(ctx context.Context, options ...header.HeadOption) (*headertest.DummyHeader, error) {
	params := header.DefaultHeadRequestParams()
	for _, opt := range options {
		opt(&params)
	}
	if params.DisableSubjectiveInit != nil {
		t.disabledSubjectiveInit = true
	}
	return t.ex.Head(ctx, options...)
}

func (t *wrappedGetter) Get(ctx context.Context, hash header.Hash) (*headertest.DummyHeader, error) {
	//TODO implement me
	panic("implement me")
}

func (t *wrappedGetter) GetByHeight(ctx context.Context, u uint64) (*headertest.DummyHeader, error) {
	//TODO implement me
	panic("implement me")
}

func (t *wrappedGetter) GetRangeByHeight(ctx context.Context, from, amount uint64) ([]*headertest.DummyHeader, error) {
	//TODO implement me
	panic("implement me")
}

func (t *wrappedGetter) GetVerifiedRange(ctx context.Context, from *headertest.DummyHeader, amount uint64) ([]*headertest.DummyHeader, error) {
	//TODO implement me
	panic("implement me")
}
