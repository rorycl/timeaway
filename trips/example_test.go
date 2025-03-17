package trips

import (
	"fmt"
	"log"
	"net/url"
)

func Example() {

	// package variables
	WindowMaxDays = 180      // maximum window of days to calculate over
	CompoundStayMaxDays = 90 // longest allowed compound trip length in days

	// fail immediately on error
	fe := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	// add trips by url
	url, _ := url.ParseRequestURI(
		"http://test.com/?" +
			"Start=2022-01-01&End=2022-01-01&" +
			"Start=2023-01-10&End=2023-02-08&" +
			"Start=2023-02-11&End=2023-04-04&" +
			"Start=2023-07-01&End=2023-07-30&" +
			"Start=2024-06-10&End=2024-06-14",
	)
	_, err := HolidaysURLDecoder(url.Query()) // replace _ with holidays
	fe(err)

	// or add trips by json
	json := []byte(
		`[{"Start":"2022-01-01", "End":"2022-01-01"},
		  {"Start":"2023-01-10", "End":"2023-02-08"},
		  {"Start":"2023-02-11", "End":"2023-04-04"},
		  {"Start":"2023-07-01", "End":"2023-07-30"},
		  {"Start":"2024-06-10", "End":"2024-06-14"}]`,
	)
	holidays, err := HolidaysJSONDecoder(json)
	fe(err)

	// calculate
	trips, err := Calculate(holidays)
	fe(err)

	// show whether or not trips breach, the maximum compound days away,
	// and other trip details
	fmt.Printf("breach?: %t\n", trips.Breach)
	fmt.Printf("longest stay: %d\n", trips.DaysAway)
	fmt.Printf("window details: %s\n", trips.Window)
	fmt.Printf("window components: %s", trips.Holidays)

	// Output:
	// breach?: true
	// longest stay: 91
	// window details: window Tuesday 10 January 2023:Saturday 8 July 2023 overlap Tuesday 10 January 2023:Saturday 8 July 2023 (91 days, 3 overlaps)
	// window components: [01/01/2022 to 01/01/2022 (1 days) 10/01/2023 to 08/02/2023 (30 days) [overlap 30 days] 11/02/2023 to 04/04/2023 (53 days) [overlap 53 days] 01/07/2023 to 30/07/2023 (30 days) [overlap 8 days] 10/06/2024 to 14/06/2024 (5 days)]
}
