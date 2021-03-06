package main

// TempData temp data used in the tree loop
type TempData struct {
	Exclusion     Exclusion
	Licenses      Licenses
	Tarballs      Tarballs
	ResponseCache ResponseCache
}

// NewTempData initialize a new tempData structure
func NewTempData() TempData {
	return TempData{
		Exclusion{},
		Licenses{},
		Tarballs{},
		ResponseCache{},
	}
}
