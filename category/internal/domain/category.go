package domain

import (
	"github.com/jongyunha/lunchbox/internal/ddd"
	"github.com/stackus/errors"
)

const CategoryAggregate = "categories.Category"

type Category struct {
	ddd.Aggregate
	Name string
}

var (
	ErrCategoryNameCannotBeBlank = errors.Wrap(errors.ErrBadRequest, "the category name cannot be blank")
)

func NewCategory(id, name string) *Category {
	return &Category{
		Aggregate: ddd.NewAggregate(id, CategoryAggregate),
		Name:      name,
	}
}

func RegisterCategory(id, name string) (*Category, error) {
	if name == "" {
		return nil, ErrCategoryNameCannotBeBlank
	}

	category := NewCategory(id, name)

	category.AddEvent(CategoryRegisteredEvent, &CategoryRegistered{
		Category: category,
	})

	return category, nil
}
