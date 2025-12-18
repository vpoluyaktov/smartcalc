package datetime

import (
	"strings"
	"time"
)

// CityTimezones maps city names to IANA timezone identifiers
var CityTimezones = map[string]string{
	// US Cities
	"seattle":       "America/Los_Angeles",
	"los angeles":   "America/Los_Angeles",
	"la":            "America/Los_Angeles",
	"san francisco": "America/Los_Angeles",
	"sf":            "America/Los_Angeles",
	"portland":      "America/Los_Angeles",
	"denver":        "America/Denver",
	"phoenix":       "America/Phoenix",
	"chicago":       "America/Chicago",
	"dallas":        "America/Chicago",
	"houston":       "America/Chicago",
	"austin":        "America/Chicago",
	"new york":      "America/New_York",
	"nyc":           "America/New_York",
	"boston":        "America/New_York",
	"miami":         "America/New_York",
	"atlanta":       "America/New_York",
	"washington":    "America/New_York",
	"dc":            "America/New_York",
	"philadelphia":  "America/New_York",
	"anchorage":     "America/Anchorage",
	"honolulu":      "America/Honolulu",
	"hawaii":        "America/Honolulu",

	// Canada
	"toronto":   "America/Toronto",
	"vancouver": "America/Vancouver",
	"montreal":  "America/Montreal",
	"calgary":   "America/Edmonton",

	// Europe
	"london":     "Europe/London",
	"paris":      "Europe/Paris",
	"berlin":     "Europe/Berlin",
	"amsterdam":  "Europe/Amsterdam",
	"rome":       "Europe/Rome",
	"madrid":     "Europe/Madrid",
	"barcelona":  "Europe/Madrid",
	"vienna":     "Europe/Vienna",
	"zurich":     "Europe/Zurich",
	"stockholm":  "Europe/Stockholm",
	"oslo":       "Europe/Oslo",
	"copenhagen": "Europe/Copenhagen",
	"helsinki":   "Europe/Helsinki",
	"warsaw":     "Europe/Warsaw",
	"prague":     "Europe/Prague",
	"budapest":   "Europe/Budapest",
	"athens":     "Europe/Athens",
	"istanbul":   "Europe/Istanbul",

	// Eastern Europe / Russia
	"moscow":           "Europe/Moscow",
	"st petersburg":    "Europe/Moscow",
	"saint petersburg": "Europe/Moscow",
	"kiev":             "Europe/Kyiv",
	"kyiv":             "Europe/Kyiv",
	"minsk":            "Europe/Minsk",
	"omsk":             "Asia/Omsk",
	"novosibirsk":      "Asia/Novosibirsk",
	"yekaterinburg":    "Asia/Yekaterinburg",
	"vladivostok":      "Asia/Vladivostok",

	// Asia
	"tokyo":     "Asia/Tokyo",
	"osaka":     "Asia/Tokyo",
	"seoul":     "Asia/Seoul",
	"beijing":   "Asia/Shanghai",
	"shanghai":  "Asia/Shanghai",
	"hong kong": "Asia/Hong_Kong",
	"hongkong":  "Asia/Hong_Kong",
	"singapore": "Asia/Singapore",
	"bangkok":   "Asia/Bangkok",
	"jakarta":   "Asia/Jakarta",
	"mumbai":    "Asia/Kolkata",
	"delhi":     "Asia/Kolkata",
	"bangalore": "Asia/Kolkata",
	"kolkata":   "Asia/Kolkata",
	"dubai":     "Asia/Dubai",
	"abu dhabi": "Asia/Dubai",
	"tel aviv":  "Asia/Jerusalem",
	"jerusalem": "Asia/Jerusalem",

	// Australia / Pacific
	"sydney":    "Australia/Sydney",
	"melbourne": "Australia/Melbourne",
	"brisbane":  "Australia/Brisbane",
	"perth":     "Australia/Perth",
	"auckland":  "Pacific/Auckland",

	// South America
	"sao paulo":      "America/Sao_Paulo",
	"rio de janeiro": "America/Sao_Paulo",
	"buenos aires":   "America/Argentina/Buenos_Aires",
	"santiago":       "America/Santiago",
	"lima":           "America/Lima",
	"bogota":         "America/Bogota",

	// Africa
	"cairo":        "Africa/Cairo",
	"johannesburg": "Africa/Johannesburg",
	"lagos":        "Africa/Lagos",
	"nairobi":      "Africa/Nairobi",
}

// TimezoneAbbreviations maps common timezone abbreviations to IANA identifiers
var TimezoneAbbreviations = map[string]string{
	"pst":       "America/Los_Angeles",
	"pdt":       "America/Los_Angeles",
	"mst":       "America/Denver",
	"mdt":       "America/Denver",
	"cst":       "America/Chicago",
	"cdt":       "America/Chicago",
	"est":       "America/New_York",
	"edt":       "America/New_York",
	"utc":       "UTC",
	"gmt":       "UTC",
	"bst":       "Europe/London",
	"cet":       "Europe/Paris",
	"cest":      "Europe/Paris",
	"eet":       "Europe/Kyiv",
	"eest":      "Europe/Kyiv",
	"msk":       "Europe/Moscow",
	"jst":       "Asia/Tokyo",
	"kst":       "Asia/Seoul",
	"cst china": "Asia/Shanghai",
	"ist":       "Asia/Kolkata",
	"aest":      "Australia/Sydney",
	"aedt":      "Australia/Sydney",
	"nzst":      "Pacific/Auckland",
	"nzdt":      "Pacific/Auckland",
}

// LookupTimezone finds a timezone by city name or abbreviation
func LookupTimezone(name string) (*time.Location, error) {
	name = strings.ToLower(strings.TrimSpace(name))

	// Try city lookup first
	if tz, ok := CityTimezones[name]; ok {
		return time.LoadLocation(tz)
	}

	// Try abbreviation
	if tz, ok := TimezoneAbbreviations[name]; ok {
		return time.LoadLocation(tz)
	}

	// Try as IANA timezone directly
	return time.LoadLocation(name)
}

// GetLocalTimezone returns the system's local timezone
func GetLocalTimezone() *time.Location {
	return time.Local
}
