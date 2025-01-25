package domain

const (
	CategoryRegisteredEvent = "category.CategoryRegistered"
)

type CategoryRegistered struct {
	Category *Category
}

func (CategoryRegistered) Key() string { return CategoryRegisteredEvent }
