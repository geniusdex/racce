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
		&CarModel{12, "Aston Martin", "V12 Vantage GT3", 2013, GT3},
		&CarModel{20, "Aston Martin", "V8 Vantage GT3", 2019, GT3},
		&CarModel{3, "Audi", "R8 LMS", 2015, GT3},
		&CarModel{19, "Audi", "R8 LMS Evo", 2019, GT3},
		&CarModel{11, "Bentley", "Continental GT3", 2015, GT3},
		&CarModel{8, "Bentley", "Continental GT3", 2018, GT3},
		&CarModel{7, "BMW", "M6 GT3", 2017, GT3},
		&CarModel{14, "Emil Frey Jaguar", "G3", 2012, GT3},
		&CarModel{2, "Ferrari", "488 GT3", 2018, GT3},
		&CarModel{17, "Honda", "NSX GT3", 2017, GT3},
		&CarModel{21, "Honda", "NSX GT3 Evo", 2019, GT3},
		&CarModel{4, "Lamborhini", "Huracan GT3", 2015, GT3},
		&CarModel{16, "Lamborghini", "Huracan GT3 Evo", 2019, GT3},
		&CarModel{18, "Lamborghini", "Huracan Super Trofeo", 2015, SuperTrofeo},
		&CarModel{15, "Lexus", "RC F GT3", 2016, GT3},
		&CarModel{5, "McLaren", "650S GT3", 2015, GT3},
		&CarModel{22, "McLaren", "720S GT3", 2019, GT3},
		&CarModel{1, "Mercedes-AMG", "GT3", 2015, GT3}, // &CarModel{1, "LORD MERCEDES AMG", "AMG GT THE THIRD", 2015, GT3},
		&CarModel{10, "Nissan", "GT-R Nismo GT3", 2015, GT3},
		&CarModel{6, "Nissan", "GT-R Nismo GT3", 2018, GT3},
		&CarModel{0, "Porsche", "991 GT3 R", 2018, GT3},
		&CarModel{9, "Porsche", "991 II GT3 Cup", 2017, PorscheCup},
		&CarModel{23, "Porsche", "991 II GT3 R", 2019, GT3},
		&CarModel{13, "Reiter Engineering", "R-EX GT3", 2017, GT3},

		&CarModel{50, "Alpine", "A110 GT4", 2018, GT4},
		&CarModel{51, "Aston Martin", "Vantage GT4", 2018, GT4},
		&CarModel{52, "Audi", "R8 LMS GT4", 2018, GT4},
		&CarModel{53, "BMW", "M4 GT4", 2018, GT4},
		&CarModel{55, "Chevrolet", "Camaro GT4.R", 2017, GT4},
		&CarModel{56, "Ginetta", "G55 GT4", 2012, GT4},
		&CarModel{57, "KTM", "X-Bow GT4", 2016, GT4},
		&CarModel{58, "Maserati", "GranTurismo MC GT4", 2016, GT4},
		&CarModel{59, "McLaren", "570S GT4", 2016, GT4},
		&CarModel{60, "Mercedes-AMG", "GT4", 2016, GT4},
		&CarModel{61, "Porsche", "718 Cayman GT4 Clubsport", 2019, GT4},
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
