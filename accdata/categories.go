package accdata

import "image/color"

// DriverCategory describes a single driver category
type DriverCategory struct {
	// ID is the numerical ID of this category
	ID int
	// Name is the name of this category
	Name string
}

// CupCategory describes a single cup category
type CupCategory struct {
	// ID is the numerical ID of this category
	ID int
	// Name is the name of this category
	Name string
	// Color is the color used to indicate this category in driver numbers
	Color color.RGBA
	// TextColor is a text color that can be used for the driver numbers if the background is Color
	TextColor color.RGBA
}

var (
	// DriverCategories contains information on all available driver categories
	DriverCategories = []*DriverCategory{
		&DriverCategory{3, "Platinum"},
		&DriverCategory{2, "Gold"},
		&DriverCategory{1, "Silver"},
		&DriverCategory{0, "Bronze"},
	}

	// CupCategories contains information on all available cup categories
	CupCategories = []*CupCategory{
		&CupCategory{0, "Overall", color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 255}},
		&CupCategory{1, "ProAm", color.RGBA{128, 128, 128, 255}, color.RGBA{0, 0, 0, 255}},
		&CupCategory{2, "Am", color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 0, 255}},
		&CupCategory{3, "Silver", color.RGBA{128, 128, 128, 255}, color.RGBA{0, 0, 0, 255}},
		&CupCategory{4, "National", color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}},
	}

	// driverCategoriesByID is a cache for DriverCategoryByID
	driverCategoriesByID map[int]*DriverCategory

	// cupCategoriesByID is a cache for CupCategoryByID
	cupCategoriesByID map[int]*CupCategory
)

// DriverCategoryByID returns the driver category with the given ID
func DriverCategoryByID(ID int) *DriverCategory {
	if driverCategoriesByID == nil {
		driverCategoriesByID = make(map[int]*DriverCategory)
		for _, category := range DriverCategories {
			driverCategoriesByID[category.ID] = category
		}
	}
	return driverCategoriesByID[ID]
}

// CupCategoryByID returns the driver category with the given ID
func CupCategoryByID(ID int) *CupCategory {
	if cupCategoriesByID == nil {
		cupCategoriesByID = make(map[int]*CupCategory)
		for _, category := range CupCategories {
			cupCategoriesByID[category.ID] = category
		}
	}
	return cupCategoriesByID[ID]
}
