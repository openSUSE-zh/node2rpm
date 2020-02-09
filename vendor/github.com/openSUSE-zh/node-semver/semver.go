package semver

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

const (
	MAIN_VER = `(\d+|x|X|\*)`
	PRE_VER  = `([0-9A-Za-z\.-]+)`
	SEMVER   = MAIN_VER + `(\.` + MAIN_VER + `)?(\.` + MAIN_VER + `)?(-` + PRE_VER + `)?(\+` + PRE_VER + `)?`
)

type Collection []Semver

func (c Collection) Len() int {
	return len(c)
}

func (c Collection) Less(i, j int) bool {
	if c[j].GreaterThan(c[i]) {
		return true
	}
	return false
}

func (c Collection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Semver semantic version structure
type Semver struct {
	Major         string
	Minor         string
	Patch         string
	Prerelease    string
	BuildMetadata string
}

// NewSemver initialize a new semantic version
func NewSemver(s string) Semver {
	re := regexp.MustCompile(SEMVER)
	if !re.MatchString(s) {
		log.Fatalf("%s is not a semantic version", s)
	}
	m := re.FindStringSubmatch(s)

	minor := "0"
	patch := "0"
	if len(m[3]) > 0 {
		minor = m[3]
	}
	if len(m[5]) > 0 {
		patch = m[5]
	}
	return Semver{m[1], minor, patch, m[7], m[9]}
}

func (v Semver) mainEqual(v1 Semver) bool {
	if v.Major == v1.Major && v.Minor == v1.Minor && v.Patch == v1.Patch {
		return true
	}
	return false
}

// Equal if two semantic versions equal
func (v Semver) Equal(v1 Semver) bool {
	if v.mainEqual(v1) && v.Prerelease == v1.Prerelease {
		return true
	}
	return false
}

// String format the semantic version as string
func (v Semver) String() string {
	str := v.Major
	for i, j := range []string{v.Minor, v.Patch, v.Prerelease, v.BuildMetadata} {
		if len(j) > 0 {
			if i < 2 {
				str += "."
			}
			if i == 2 {
				str += "-"
			}
			if i == 3 {
				str += "+"
			}
			str += j
		}
	}
	return str
}

// GreaterThan if a semantic version is greater than another
func (v Semver) GreaterThan(v1 Semver) bool {
	return v.gt(v1, false)
}

func (v Semver) GreaterEqualThan(v1 Semver) bool {
	return v.gt(v1, true)
}

func (v Semver) LowerThan(v1 Semver) bool {
	return !v.gt(v1, true)
}

func (v Semver) LowerEqualThan(v1 Semver) bool {
	return !v.gt(v1, false)
}

func (v Semver) NotEqual(v1 Semver) bool {
	return !v.Equal(v1)
}

func (v Semver) gt(v1 Semver, e bool) bool {
	if v.mainEqual(v1) {
		if v.Prerelease == v1.Prerelease {
			return e
		}
		if len(v.Prerelease) == 0 {
			return true
		}
		if len(v1.Prerelease) == 0 {
			return false
		}
		return comparePrerelease(v.Prerelease, v1.Prerelease)
	}
	i := compare(v.Major, v1.Major)
	i1 := compare(v.Minor, v1.Minor)
	i2 := compare(v.Patch, v1.Patch)

	if i != 0 {
		return i > 0
	}

	if i == 0 {
		if i1 == 0 {
			return i2 > 0
		}
		return i1 > 0
	}

	return false
}

func compare(v, v1 string) int {
	if v == v1 {
		return 0
	}
	i, _ := strconv.Atoi(v)
	i1, _ := strconv.Atoi(v1)
	if i > i1 {
		return 1
	}
	return -1
}

func compare1(v, v1 string) int {
	re := regexp.MustCompile("[A-Za-z]")
	b := re.MatchString(v)
	b1 := re.MatchString(v1)

	if b && !b1 {
		return 1
	}

	if !b && b1 {
		return -1
	}

	if b && b1 {
		return strings.Compare(v, v1)
	}

	if !b && !b1 {
		return compare(v, v1)
	}

	return 0
}

func comparePrerelease(v, v1 string) bool {
	s := strings.Split(v, ".")
	s1 := strings.Split(v1, ".")
	l := len(s)
	l1 := len(s1)
	var idx int
	var long string
	if l-l1 >= 0 {
		idx = l1
		long = v
	} else {
		idx = l
		long = v1
	}

	for i := 0; i < idx; i++ {
		j := compare1(s[i], s1[i])
		if j == 0 {
			continue
		}
		return j > 0
	}

	if long == v {
		return true
	}

	return false
}
