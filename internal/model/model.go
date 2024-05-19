package model

type Listing struct {
	ListingId     string
	Title         string
	Price         string
	ThumbnailUrl  string
	Url           string
	FuelType      string
	Year          string
	CubicCapacity string
	PowerkW       string
	PowerHP       string
	Milage        string
	IsHighlighted bool
	firstSeen     int64
	lastSeen      int64
}
