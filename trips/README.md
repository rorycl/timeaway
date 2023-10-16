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

The example below is taken from `example_test.go`. Note that the last
line of output is wrapped for readability

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
    fmt.Printf("window details: %s\n", trips.Window())
    fmt.Printf("window components: %s", trips.Holidays)

    // Output:
    // breach?: true
    // longest stay: 91
    // window details: Tuesday 10 January 2023 : Saturday 8 July 2023 (91 days, 3 overlaps)
    // window components: [
        01/01/2022 to 01/01/2022 (1 days) 10/01/2023 to 08/02/2023 (30 days) [overlap 30 days]
        11/02/2023 to 04/04/2023 (53 days) [overlap 53 days] 01/07/2023 to 30/07/2023 (30 days) [overlap 8 days]
        10/06/2024 to 14/06/2024 (5 days)
       ]
}
```
