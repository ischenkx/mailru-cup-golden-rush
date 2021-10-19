package api

type TreasureNotFoundErr struct{}
type WrongCoordinatesErr struct{}
type NoMoreLicensesAllowedErr struct{}
type TreasureIsNotDugErr struct{}
type NoSuchLicenseErr struct{}
type WrongDepthErr struct{}
type NotStatedErr struct{}

func (TreasureNotFoundErr) Error() string {
	return "no treasure found"
}

func (WrongDepthErr) Error() string {
	return "wrong depth"
}

func (WrongCoordinatesErr) Error() string {
	return "wrong coordinates (x,y)"
}

func (NotStatedErr) Error() string {
	return "unknown error"
}

func (NoMoreLicensesAllowedErr) Error() string {
	return "no more licenses allowed"
}

func (TreasureIsNotDugErr) Error() string {
	return "treasure has not been digged"
}

func (NoSuchLicenseErr) Error() string {
	return "no such license"
}
