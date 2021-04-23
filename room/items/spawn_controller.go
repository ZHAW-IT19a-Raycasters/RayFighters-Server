package items

import "github.zhaw.ch/raycasters/server/gamemap"

// itemMap represents the special item map
type itemMap map[gamemap.MapCoordinate]Item

// ItemController hosts all the special items of the
// map and checks for their state.
type ItemController struct {
	items itemMap
}

// NewItemControlller will create a map for the item pickups.
func NewItemControlller(mapData gamemap.Map2d) ItemController {
	im := createItemMap(mapData)
	return ItemController{
		items: im,
	}
}

// GetItem will return teh item at the current position,
// if it exists.
func (ic ItemController) GetItem(coor gamemap.MapCoordinate) (Item, bool) {
	i, b := ic.items[coor]
	return i, b
}

// createItemMap will parse the map data and create
// a item map representation.
func createItemMap(mapData gamemap.Map2d) itemMap {
	im := make(itemMap)

	dim := mapData.GetMapDimension()

	for i := 0; i < dim; i += 1 {
		for j := 0; j < dim; j += 1 {
			// error is not important as we
			// get the dimension before
			item, _ := mapData.MapElementAt(i, j)
			if item == gamemap.Medkit || item == gamemap.Weapon {
				coord := gamemap.NewMapCoordinate(float64(i), float64(j))
				im[coord] = newItem(item)
			}
		}
	}

	return im
}
