package trips

import "fmt"

func Example() {

	window := 180               // window in days
	compoundStayMaxLength := 90 // longest allowed compound trip length in days

	// create a new trips struct
	trs, _ := NewTrips(window, compoundStayMaxLength)

	// add trips
	_ = trs.AddTrip("2022-01-01", "2022-01-01")
	_ = trs.AddTrip("2023-01-06", "2023-02-07")
	_ = trs.AddTrip("2023-02-11", "2023-04-04")
	_ = trs.AddTrip("2023-06-10", "2023-06-14")

	// calculate
	_ = trs.Calculate()

	// show the longest trips
	breach, longestWindow, _ := trs.LongestTrip()

	fmt.Printf("breach? %t\n", breach)
	fmt.Printf("top longestWindow result : %v\n", longestWindow)
	fmt.Printf("trip details:%v", trs)

	// Output:
	// breach? true
	// top longestWindow result : Saturday 17 December 2022 : Wednesday 14 June 2023 (91) Friday 6 January 2023 to Tuesday 7 February 2023 Saturday 11 February 2023 to Tuesday 4 April 2023 Saturday 10 June 2023 to Wednesday 14 June 2023
	// trip details:
	// window      180
	// maxStay     90
	// startFrame  Saturday 1 January 2022
	// endFrame    Saturday 17 December 2022
	// longestStay 91
	// trips       4
	// windows     351
	// breach      true
}
