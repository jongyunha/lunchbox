package domain

type RestaurantV1 struct {
	Name string
}

func (RestaurantV1) SnapshotName() string { return "restaurants.RestaurantV1" }
