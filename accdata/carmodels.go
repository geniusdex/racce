package accdata

// CarGroup defines the group to which a car model belongs
type CarGroup string

const (
	GT3         CarGroup = "GT3"
	GT4                  = "GT4"
	PorscheCup           = "Cup"
	SuperTrofeo          = "ST"
)

// CarModel describes a single car model
type CarModel struct {
	// ID is the numerical ID of this car
	ID int
	// ManufacturerLabel is the label of the car manufacturer
	ManufacturerLabel string
	// Manufacturer is the manufacturer of the car
	Manufacturer string
	// Model is the name of the car model
	Model string
	// Year is the year in which the car model was made available
	Year int
	// Group is the group to which this car model belongs
	Group CarGroup
}

var (
	// CarModels contains information on all available car models
	CarModels = []*CarModel{
		{12, "aston_martin", "Aston Martin", "V12 Vantage GT3", 2013, GT3},
		{20, "aston_martin", "Aston Martin", "V8 Vantage GT3", 2019, GT3},
		{3, "audi", "Audi", "R8 LMS", 2015, GT3},
		{19, "audi", "Audi", "R8 LMS Evo", 2019, GT3},
		{11, "bentley", "Bentley", "Continental GT3", 2015, GT3},
		{8, "bentley", "Bentley", "Continental GT3", 2018, GT3},
		{30, "bmw", "BMW", "M4 GT3", 2022, GT3},
		{7, "bmw", "BMW", "M6 GT3", 2017, GT3},
		{14, "jaguar", "Emil Frey Jaguar", "G3", 2012, GT3},
		{2, "ferrari", "Ferrari", "488 GT3", 2018, GT3},
		{24, "ferrari", "Ferrari", "488 GT3 Evo", 2020, GT3},
		{17, "honda", "Honda", "NSX GT3", 2017, GT3},
		{21, "honda", "Honda", "NSX GT3 Evo", 2019, GT3},
		{4, "lamborghini", "Lamborghini", "Huracan GT3", 2015, GT3},
		{16, "lamborghini", "Lamborghini", "Huracan GT3 Evo", 2019, GT3},
		{18, "lamborghini", "Lamborghini", "Huracan Super Trofeo", 2015, SuperTrofeo},
		{15, "lexus", "Lexus", "RC F GT3", 2016, GT3},
		{5, "mclaren", "McLaren", "650S GT3", 2015, GT3},
		{22, "mclaren", "McLaren", "720S GT3", 2019, GT3},
		{1, "mercedes-amg", "Mercedes-AMG", "GT3", 2015, GT3}, // {1, "lord_mercedes_amg", "LORD MERCEDES AMG", "AMG GT THE THIRD", 2015, GT3},
		{25, "mercedes-amg", "Mercedes-AMG", "GT3 2020", 2020, GT3},
		{10, "nissan", "Nissan", "GT-R Nismo GT3", 2015, GT3},
		{6, "nissan", "Nissan", "GT-R Nismo GT3", 2018, GT3},
		{0, "porsche", "Porsche", "991 GT3 R", 2018, GT3},
		{9, "porsche", "Porsche", "991 II GT3 Cup", 2017, PorscheCup},
		{23, "porsche", "Porsche", "991 II GT3 R", 2019, GT3},
		{13, "reiter_engineering", "Reiter Engineering", "R-EX GT3", 2017, GT3},

		{50, "alpine", "Alpine", "A110 GT4", 2018, GT4},
		{51, "aston_martin", "Aston Martin", "Vantage GT4", 2018, GT4},
		{52, "audi", "Audi", "R8 LMS GT4", 2018, GT4},
		{53, "bmw", "BMW", "M4 GT4", 2018, GT4},
		{55, "chevrolet", "Chevrolet", "Camaro GT4.R", 2017, GT4},
		{56, "ginetta", "Ginetta", "G55 GT4", 2012, GT4},
		{57, "ktm", "KTM", "X-Bow GT4", 2016, GT4},
		{58, "maserati", "Maserati", "GranTurismo MC GT4", 2016, GT4},
		{59, "mclaren", "McLaren", "570S GT4", 2016, GT4},
		{60, "mercedes-amg", "Mercedes-AMG", "GT4", 2016, GT4},
		{61, "porsche", "Porsche", "718 Cayman GT4 Clubsport", 2019, GT4},
	}

	// carModelsByID is a cache for CarModelByID
	carModelsByID map[int]*CarModel
)

// CarModelByID returns the car model with the given ID
func CarModelByID(ID int) *CarModel {
	if carModelsByID == nil {
		carModelsByID = make(map[int]*CarModel)
		for _, model := range CarModels {
			carModelsByID[model.ID] = model
		}
	}
	return carModelsByID[ID]
}
