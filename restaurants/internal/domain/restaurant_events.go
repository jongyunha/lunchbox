package domain

const (
	RestaurantRegisteredEvent = "restaurant.RestaurantRegistered"
)

type RestaurantRegistered struct {
	Name string
}

func (RestaurantRegistered) Key() string { return RestaurantRegisteredEvent }
