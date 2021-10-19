package models

type Dig struct {
	Depth int64 `json:"depth"`
	LicenseID int64 `json:"licenseID"`
	PosX int64 `json:"posX"`
	PosY int64 `json:"posY"`
}

type License struct {
	DigAllowed int64 `json:"digAllowed"`
	DigUsed int64 `json:"digUsed"`
	ID int64 `json:"id"`
}

type Area struct {
	PosX int64 `json:"posX"`
	PosY int64 `json:"posY"`
	SizeX int64 `json:"sizeX,omitempty"`
	SizeY int64 `json:"sizeY,omitempty"`
}

type Report struct {
	Amount int64 `json:"amount"`
	Area Area `json:"area"`
}

type Error struct {
	Code int32 `json:"code"`
	Message string `json:"message"`
}