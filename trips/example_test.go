package trips

import "fmt"

func Example() {

	window := 180               // window of days over which trips may only reach compoundStayMaxLength
	compoundStayMaxLength := 90 // compounded trip days maximum in window
	resultsNo := 1              // number of results to show

	// create a new trips struct
	trips, _ := NewTrips(window, compoundStayMaxLength)

	// add trips
	_ = trips.AddTrip("2022-01-01", "2022-01-01")
	_ = trips.AddTrip("2023-01-06", "2023-02-07")
	_ = trips.AddTrip("2023-02-11", "2023-04-04")
	_ = trips.AddTrip("2023-06-10", "2023-06-14")

	// calculate
	_ = trips.Calculate()

	/* show the longest trips */
	breach, windows := trips.LongestTrips(resultsNo)

	fmt.Printf("breach? %t\n", breach)
	fmt.Printf("top window result : %v\n", windows[0])
	fmt.Printf("longest compound trip : %d\n", trips.longestStay)
	// fmt.Printf("trips: %v", trips)

	// Output:
	// breach? true
	// top window result : Saturday 17 December 2022 : Wednesday 14 June 2023 (91) Friday 6 January 2023 to Tuesday 7 February 2023 Saturday 11 February 2023 to Tuesday 4 April 2023 Saturday 10 June 2023 to Wednesday 14 June 2023
	// longest compound trip : 91
}
