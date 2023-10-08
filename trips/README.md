# Trips package

## Background

This package helps calculate if foreign visitors' trips to Schengen
countries conform with Regulation (EU) No 610/2013 of 26 June 2013
limiting the total length of all trips to Schengen states to no more
than 90 days in any 180 day period.

The date of entry should be considered as the first day of stay on the
territory of the Member States and the date of exit should be considered
as the last day of stay on the territory of the Member States.

For more details on the Regulation and its application please see
https://ec.europa.eu/assets/home/visa-calculator/docs/short_stay_schengen_calculator_user_manual_en.pdf.

The calculation provided by this package uses a moving window configured
in days over the trips provided to find the maximum length of days,
inclusive of trip start and end dates, taken by the trips to learn if
these breach the permissible length of stay. Trips cannot overlap in
time.

## Example

```go
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

	fe := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	// add trips by url
	url, _ := url.ParseRequestURI(
		"http://test.com/?" +
			"Start=2022-01-01&End=2022-01-01&" +
			"Start=2023-01-06&End=2023-02-07&" +
			"Start=2023-02-11&End=2023-04-04&" +
			"Start=2023-06-10&End=2023-06-14",
	)
	_, err := HolidaysURLDecoder(url.Query()) // replace _ with holidays
	fe(err)

	// or add trips by json
	json := []byte(
		`[{"Start":"2022-01-01", "End":"2022-01-01"},
          {"Start":"2023-01-06", "End":"2023-02-07"},
		  {"Start":"2023-02-11", "End":"2023-04-04"},
		  {"Start":"2023-06-10", "End":"2023-06-14"}]`,
	)
	holidays, err := HolidaysJSONDecoder(json)
	fe(err)

	// calculate
	trips, err := Calculate(holidays)
	fe(err)

	// show whether or not trips breach, the maximum compound days away,
	// and other trip details
	fmt.Printf("breach?        : %t\n", trips.Breach)
	fmt.Printf("longest stay   : %v\n", trips.DaysAway)
	fmt.Printf("window details : %s : %s\n", trips.Window.Start, trips.Window.End)
	fmt.Printf("                 %s\n", trips.Window.HolidayParts)

	// Output:
	// breach?        : true
	// longest stay   : 91
	// window details : 2022-12-17 00:00:00 +0000 UTC : 2023-06-14 00:00:00 +0000 UTC
	//                  [Friday 6 January 2023 to Tuesday 7 February 2023 (33) Saturday 11 February 2023 to Tuesday 4 April 2023 (53) Saturday 10 June 2023 to Wednesday 14 June 2023 (5)]

}
```
