package simplecqrs

import (
	"fmt"
	"reflect"

	"github.com/jetbasrawi/go.cqrs"
)

// InventoryItemRepo is a repository specialized for persistence of
// InventoryItems.
//
// While it is not required to construct a repository specialized for a
// specific aggregate type, it is better to do so. There can be quite a lot of
// repository configuration that is specific to a type and it is cleaner if that
// code is contained in a specialized repository as shown here.
// Also because the CommonDomainRepository Load method returns an interface{}, a
// type assertion is required. Here the type assertion is contained in this specialized
// repo and a *InventoryItem is returned from the repo.
type InventoryItemRepo struct {
	repo *ycq.GetEventStoreCommonDomainRepo
}

// Load loads events for an aggregate.
//
// Returns an *InventoryAggregate.
func (r *InventoryItemRepo) Load(aggregateType, id string) (*InventoryItem, error) {
	ar, err := r.repo.Load(reflect.TypeOf(&InventoryItem{}).Elem().Name(), id)
	if _, ok := err.(*ycq.ErrAggregateNotFound); ok {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if ret, ok := ar.(*InventoryItem); ok {
		return ret, nil
	}

	return nil, fmt.Errorf("Could not cast aggregate returned to type of %s", reflect.TypeOf(&InventoryItem{}).Elem().Name())
}

// Save persists an aggregate.
func (r *InventoryItemRepo) Save(aggregate ycq.AggregateRoot, expectedVersion *int) error {
	return r.repo.Save(aggregate, expectedVersion)
}
