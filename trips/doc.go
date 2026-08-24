// Package trips helps calculate if foreign visitors' trips to Schengen countries conform with Regulation (EU) No
// 610/2013 of 26 June 2013 limiting the total length of all trips to Schengen states to no more than 90 days in any 180
// day period. Note that trips to non-Schengen EU states (Cyprus and Ireland) should not be included.
//
// The date of entry should be considered as the first day of stay on the territory of the Member States and the date of
// exit should be considered as the last day of stay on the territory of the Member States.
//
// For more details on the Regulation and its application please see
// https://ec.europa.eu/assets/home/visa-calculator/docs/short_stay_schengen_calculator_user_manual_en.pdf.
//
// The calculation provided by this package uses a moving window configured in days over the trips provided to find the
// maximum length of days, inclusive of trip start and end dates, taken by the trips to learn if these breach the
// permissible length of stay. Trips may not overlap in time.
package trips
