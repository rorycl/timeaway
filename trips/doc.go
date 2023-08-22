// This package helps calculate if foreign visitors' trips to Schengen countries conform with Regulation (EU) No
// 610/2013 of 26 June 2013 limiting the total length of all trips to Schengen states to no more than 90 days in any 180
// day period.
//
// The date of entry should be considered as the first day of stay on the territory of the Member States and the date of
// exit should be considered as the last day of stay on the territory of the Member States.
//
// For more details on the Regulation and its application please see
// https://ec.europa.eu/assets/home/visa-calculator/docs/short_stay_schengen_calculator_user_manual_en.pdf.
//
// The calculation provided by this package uses a moving window configured in days over the trips provided to find the
// maximum length of days, inclusive of trip start and end dates, taken by the trips to learn if these breach the
// permissible length of stay. Trips cannot overlap in time.
//
// example:
//
// 	window := 180               // window in days
// 	compoundStayMaxLength := 90 // longest allowed compound trip length in days
//
// 	// create a new trips struct
// 	trs, _ := NewTrips(window, compoundStayMaxLength)
//
// 	// add trips
// 	_ = trs.AddTrip("2022-01-01", "2022-01-01")
// 	_ = trs.AddTrip("2023-01-06", "2023-02-07")
// 	_ = trs.AddTrip("2023-02-11", "2023-04-04")
// 	_ = trs.AddTrip("2023-06-10", "2023-06-14")
//
// 	// calculate
// 	_ = trs.Calculate()
//
// 	/* show the longest trips */
// 	breach, longestWindow, _ := trs.LongestTrip()
//
// 	fmt.Printf("breach? %t\n", breach)
// 	fmt.Printf("top longestWindow result : %v\n", longestWindow)
// 	fmt.Printf("trip details:%v", trs)
package trips

// vim: noai:ts=4:tw=120
