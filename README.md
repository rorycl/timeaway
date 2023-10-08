# timeaway

[![CI/CD](https://github.com/rorycl/timeaway/actions/workflows/pushtag.yml/badge.svg)](https://github.com/rorycl/timeaway/actions/workflows/pushtag.yml)

version v0.5.0 : 08 October 2023

A small web app to calculate if the compound length of trips to Schengen
countries by non-EU visitors conform with Regulation (EU) No 610/2013
limiting the total length of stays to no more than 90 days in any 180
day window.

(This app has also turned into a github actions/workflows experiment
inspired by the book "[Shipping
Go](https://www.manning.com/books/shipping-go)" by Joel Holmes, and
which partly accounts for large number of commits and releases in this
repo!)

The [`trips`](trips/README.md) go module provides the means for
calculation.

The provided web app provides a convenient interface for making
calculations at the `/` or `/home` endpoint. This endpoint can also
receive pairs of Start and End parameters for each trip, and sets the
url parameters each time a calculation is made, allowing calculations to
be conveniently saved or bookmarked.

The calculation method uses a 180 day moving window to calculate the
longest compound trip length (`windowDaysAway`). Where more than one
window has the same `windowDaysAway` the window with the earliest start
date is reported.

The `/trips` POST endpoint can be interacted with over json. This command:

```
curl -s -X POST -d '
[{"Start":"2022-12-01","End":"2022-12-02"},
 {"Start":"2023-01-02","End":"2023-03-30"},
 {"Start":"2023-04-01","End":"2023-04-02"},
 {"Start":"2023-09-03","End":"2023-09-12"}
]' 127.0.0.1:8000/trips | jq .
```

gives the following output, assuming the server is running on `127.0.0.1:8000/`:

```json
{
  "error": null,
  "longestWindow": {
    "Start": "2022-12-01T00:00:00Z","End": "2023-05-29T00:00:00Z",
    "partialHolidays": [
      {"Start": "2022-12-01T00:00:00Z", "End": "2022-12-02T00:00:00Z", "Duration": 2},
      {"Start": "2023-01-02T00:00:00Z", "End": "2023-03-30T00:00:00Z", "Duration": 88},
      {"Start": "2023-04-01T00:00:00Z", "End": "2023-04-02T00:00:00Z", "Duration": 2 }
    ],
    "daysAway": 92
  },
  "daysAway": 92,
  "holidays": [
    {"Start": "2022-12-01T00:00:00Z", "End": "2022-12-02T00:00:00Z", "Duration": 2},
    {"Start": "2023-01-02T00:00:00Z", "End": "2023-03-30T00:00:00Z", "Duration": 88},
    {"Start": "2023-04-01T00:00:00Z", "End": "2023-04-02T00:00:00Z", "Duration": 2},
    {"Start": "2023-09-03T00:00:00Z", "End": "2023-09-12T00:00:00Z", "Duration": 10}
  ],
  "breach": true
}

```

## Example

![](util/example.gif)

## Licence

This project is licensed under the [MIT Licence](LICENCE).
