package items

import (
	"github.zhaw.ch/raycasters/server/gamemap"
	"github.zhaw.ch/raycasters/server/utils"
)

type Item struct {
	isActive utils.AtomicBool
	itemType gamemap.MapElement
}

func newItem(itemType gamemap.MapElement) Item {
	return Item{
		isActive: utils.NewAtomicBool(true),
		itemType: itemType,
	}
}

// IsSpawned will check if the item is currenly
// spawned or not.
func (i Item) IsSpawned() bool {
	return i.isActive.Get()
}

// IsPickedUp will return true if the item was
// picked up, given it's current state.
func (i Item) IsPickedUp() bool {
	return i.isActive.SetFalse()
}

// IsType will return the type of map element
// this is.
func (i Item) IsType() gamemap.MapElement {
	return i.itemType
}
