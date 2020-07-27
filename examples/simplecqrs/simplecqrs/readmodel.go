package simplecqrs

import (
	"errors"
	"log"

	"github.com/jetbasrawi/go.cqrs"
)

var inMemoryDatabase *InMemoryDatabase

// ReadModelFacade is an interface for the readmodel facade
type ReadModelFacade interface {
	GetInventoryItems() []*InventoryItemListDto
	GetInventoryItemDetails(uuid string) *InventoryItemDetailsDto
}

// InventoryItemDetailsDto holds details for an inventory item.
type InventoryItemDetailsDto struct {
	ID           string
	Name         string
	CurrentCount int
	Version      int
}

// InventoryItemListDto provides a lightweight lookup view of an inventory item
type InventoryItemListDto struct {
	ID   string
	Name string
}

// ReadModel is an implementation of the ReadModelFacade interface.
//
// ReadModel provides an in memory read model.
type ReadModel struct {
}

// NewReadModel constructs a new read model
func NewReadModel() *ReadModel {
	if inMemoryDatabase == nil {
		inMemoryDatabase = NewInMemoryDatabase()
	}

	return &ReadModel{}
}

// GetInventoryItems returns a slice of all inventory items
func (m *ReadModel) GetInventoryItems() []*InventoryItemListDto {
	return inMemoryDatabase.List
}

// GetInventoryItemDetails gets an InventoryItemDetailsDto by ID
func (m *ReadModel) GetInventoryItemDetails(uuid string) *InventoryItemDetailsDto {
	if i, ok := inMemoryDatabase.Details[uuid]; ok {
		return i
	}
	return nil
}

// InventoryListView handles messages related to inventory and builds an
// in memory read model of inventory item summaries in a list.
type InventoryListView struct {
}

// NewInventoryListView constructs a new InventoryListView
func NewInventoryListView() *InventoryListView {
	if inMemoryDatabase == nil {
		inMemoryDatabase = NewInMemoryDatabase()
	}

	return &InventoryListView{}
}

// Handle processes events related to inventory and builds an in memory read model
func (v *InventoryListView) Handle(message ycq.EventMessage) {

	switch event := message.Event().(type) {

	case *InventoryItemCreated:

		inMemoryDatabase.List = append(inMemoryDatabase.List, &InventoryItemListDto{
			ID:   message.AggregateID(),
			Name: event.Name,
		})

	case *InventoryItemRenamed:

		for _, v := range inMemoryDatabase.List {
			if v.ID == message.AggregateID() {
				v.Name = event.NewName
				break
			}
		}

	case *InventoryItemDeactivated:
		i := -1
		for k, v := range inMemoryDatabase.List {
			if v.ID == message.AggregateID() {
				i = k
				break
			}
		}

		if i >= 0 {
			inMemoryDatabase.List = append(
				inMemoryDatabase.List[:i],
				inMemoryDatabase.List[i+1:]...,
			)
		}
	}
}

// InventoryItemDetailView handles messages related to inventory and builds an
// in memory read model of inventory item details.
type InventoryItemDetailView struct {
}

// Handle handles events and build the projection
func (v *InventoryItemDetailView) Handle(message ycq.EventMessage) {

	switch event := message.Event().(type) {

	case *InventoryItemCreated:

		inMemoryDatabase.Details[message.AggregateID()] = &InventoryItemDetailsDto{
			ID:      message.AggregateID(),
			Name:    event.Name,
			Version: 0,
		}

	case *InventoryItemRenamed:

		d, err := v.GetDetailsItem(message.AggregateID())
		if err != nil {
			log.Fatal(err)
		}
		d.Name = event.NewName
		d.Version = *message.Version()

	case *ItemsRemovedFromInventory:

		d, err := v.GetDetailsItem(message.AggregateID())
		if err != nil {
			log.Fatal(err)
		}
		d.CurrentCount -= event.Count

	case *ItemsCheckedIntoInventory:

		d, err := v.GetDetailsItem(message.AggregateID())
		if err != nil {
			log.Fatal(err)
		}
		d.CurrentCount += event.Count

	case *InventoryItemDeactivated:

		delete(inMemoryDatabase.Details, message.AggregateID())

	}
}

// GetDetailsItem gets an InventoryItemDetailsDto by ID
func (v *InventoryItemDetailView) GetDetailsItem(id string) (*InventoryItemDetailsDto, error) {

	d, ok := inMemoryDatabase.Details[id]
	if !ok {
		return nil, errors.New("did not find the original inventory this shouldn't not happen")
	}

	return d, nil
}

// InMemoryDatabase is a simple in memory repository
type InMemoryDatabase struct {
	Details map[string]*InventoryItemDetailsDto
	List    []*InventoryItemListDto
}

// NewInMemoryDatabase constructs a new InMemoryDatabase
func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{
		Details: make(map[string]*InventoryItemDetailsDto),
	}
}
