package orm

import (
	"context"
	"fmt"
	"sync"
)

type txCtxKey struct{}

var txKey txCtxKey

type atomicLayer struct {
	db    *DB
	mu    sync.Mutex
	tx    Tx
	level int
}

type atomicManager struct {
	mu    sync.RWMutex
	layers map[string]*atomicLayer
}

var atomicMgr = &atomicManager{
	layers: make(map[string]*atomicLayer),
}

func Atomic(ctx context.Context, db *DB, fn func(ctx context.Context) error) error {
	return AtomicUsing(ctx, db, "default", fn)
}

func AtomicUsing(ctx context.Context, db *DB, alias string, fn func(ctx context.Context) error) error {
	atomicMgr.mu.Lock()
	layer, exists := atomicMgr.layers[alias]
	if !exists {
		layer = &atomicLayer{db: db}
		atomicMgr.layers[alias] = layer
	}
	atomicMgr.mu.Unlock()

	layer.mu.Lock()
	defer layer.mu.Unlock()

	existingTx := txFromContext(ctx, alias)
	if existingTx != nil {
		layer.tx = existingTx
		layer.level++
		savepointName := fmt.Sprintf("s%d", layer.level)
		if err := existingTx.Savepoint(ctx, savepointName); err != nil {
			return fmt.Errorf("orm: savepoint error: %w", err)
		}

		err := fn(ctx)
		if err != nil {
			_ = existingTx.RollbackToSavepoint(ctx, savepointName)
			return err
		}
		return nil
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("orm: begin transaction: %w", err)
	}

	layer.tx = tx
	layer.level = 1

	newCtx := context.WithValue(ctx, txKey, txMapValue{alias: alias, tx: tx})

	err = fn(newCtx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("orm: transaction error: %w (rollback error: %w)", err, rollbackErr)
		}
		return err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("orm: commit error: %w", commitErr)
	}

	layer.tx = nil
	layer.level = 0
	return nil
}

type txMapValue struct {
	alias string
	tx    Tx
}

func txFromContext(ctx context.Context, alias string) Tx {
	val := ctx.Value(txKey)
	if val == nil {
		return nil
	}
	txMap, ok := val.(txMapValue)
	if !ok {
		return nil
	}
	if txMap.alias == alias {
		return txMap.tx
	}
	return nil
}

func GetTx(ctx context.Context) Tx {
	val := ctx.Value(txKey)
	if val == nil {
		return nil
	}
	txMap, ok := val.(txMapValue)
	if !ok {
		return nil
	}
	return txMap.tx
}

func OnCommit(ctx context.Context, fn func()) {
}

func OnRollback(ctx context.Context, fn func()) {
}