package trigger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cube2222/octosql"
	"github.com/cube2222/octosql/execution"
	"github.com/cube2222/octosql/streaming/storage"
)

func RecordReceived(t *testing.T, ctx context.Context, trigger execution.Trigger, badgerStorage *storage.BadgerStorage, key octosql.Value, eventTime time.Time) {
	tx := badgerStorage.BeginTransaction()
	keys, err := trigger.PollKeysToFire(ctx, tx)
	assert.Nil(t, keys)
	assert.Nil(t, err)
	err = trigger.RecordReceived(ctx, tx, key, eventTime)
	assert.Nil(t, err)
	assert.Nil(t, tx.Commit())
}

func ExpectFire(t *testing.T, ctx context.Context, trigger execution.Trigger, badgerStorage *storage.BadgerStorage, keys ...octosql.Value) {
	tx := badgerStorage.BeginTransaction()
	k, err := trigger.PollKeysToFire(ctx, tx)
	assert.Equal(t, keys, k)
	assert.Nil(t, err)
	assert.Nil(t, tx.Commit())
}

func ExpectNoFire(t *testing.T, ctx context.Context, trigger execution.Trigger, badgerStorage *storage.BadgerStorage) {
	tx := badgerStorage.BeginTransaction()
	keys, err := trigger.PollKeysToFire(ctx, tx)
	assert.Nil(t, keys)
	assert.Nil(t, err)
	assert.Nil(t, tx.Commit())
}

func UpdateWatermark(t *testing.T, ctx context.Context, trigger execution.Trigger, badgerStorage *storage.BadgerStorage, watermark time.Time) {
	tx := badgerStorage.BeginTransaction()
	err := trigger.UpdateWatermark(ctx, tx, watermark)
	assert.Nil(t, err)
	assert.Nil(t, tx.Commit())
}

func KeyFired(t *testing.T, ctx context.Context, trigger execution.Trigger, badgerStorage *storage.BadgerStorage, keys ...octosql.Value) {
	tx := badgerStorage.BeginTransaction()
	err := trigger.KeysFired(ctx, tx, keys)
	assert.Nil(t, err)
	assert.Nil(t, tx.Commit())
}
