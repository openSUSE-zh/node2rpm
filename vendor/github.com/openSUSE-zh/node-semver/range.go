package semver

import (
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Range []ComparatorSet

func (r Range) Satisfy(v Semver) bool {
	for _, c := range r {
		if c.Satisfy(v) {
			return true
		}
	}
	return false
}

func (r Range) String() string {
	var s string
	for i, v := range r {
		if i == len(r)-1 {
			s += v.String()
		} else {
			s += v.String() + " || "
		}
	}
	return s
}

type ComparatorSet []Comparator

func (c ComparatorSet) Len() int {
	return len(c)
}

func (c ComparatorSet) Less(i, j int) bool {
	if c[i].Version.GreaterThan(c[j].Version) {
		return false
	}
	return true
}

func (c ComparatorSet) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c ComparatorSet) Intersect() ComparatorSet {
	// group by the op. greater/lower/equal
	// n is the final set
	g := ComparatorSet{}
	l := ComparatorSet{}
	e := ComparatorSet{}
	n := ComparatorSet{}

	for _, v := range c {
		switch v.Op {
		case "=":
			e = append(e, v)
		case ">", ">=":
			g = append(g, v)
		default:
			l = append(l, v)
		}
	}

	// two fixed values can not intersect
	if len(e) > 1 {
		log.Fatalf("Not a valid comparatorset %v", c)
	}
	if len(g) > 0 {
		if len(g) < 2 {
			n = append(n, g[0])
		} else {
			sort.Sort(sort.Reverse(g))
			// we do greater than to sort
			// so two values near by may be equal
			// judge by the op if versions are equal
			if g[0].Version.Equal(g[1].Version) {
				if g[0].Op == ">" || g[1].Op == ">" {
					g[0].Op = ">"
				}
			}
			n = append(n, g[0])
		}
	}
	if len(l) > 0 {
		if len(l) < 2 {
			n = append(n, l[0])
		} else {
			sort.Sort(l)
			if l[0].Version.Equal(l[1].Version) {
				if l[0].Op == "<" || l[1].Op == "<" {
					g[0].Op = "<"
				}
			}
			n = append(n, l[0])
		}
	}
	// test fixed values in e against the n set
	if len(n) > 0 {
		if len(e) > 0 {
			for _, v := range n {
				if !v.Satisfy(e[0].Version) {
					log.Fatalf("fixed value %s doesn't fit intersected comparator set %v, can not intersect further", e[0].Version, n)
				}
			}
			return e
		}
	} else {
		return e
	}
	return n
}

func (c ComparatorSet) Satisfy(v Semver) bool {
	for _, v1 := range c {
		if !v1.Satisfy(v) {
			return false
		}
	}
	return true
}

func (c ComparatorSet) String() string {
	var s string
	for _, v := range c {
		s += " " + v.String()
	}
	return strings.TrimSpace(s)
}

type Comparator struct {
	Op      string
	Version Semver
}

func (c Comparator) Equal(c1 Comparator) bool {
	if c.Op == c1.Op && c.Version.Equal(c1.Version) {
		return true
	}
	return false
}

func (c Comparator) IsNil() bool {
	if len(c.Op) == 0 {
		return true
	}
	return false
}

func (c Comparator) Satisfy(v Semver) bool {
	// >1.2.3-beta.2, 2.3.4-beta.3 should not satisfy
	if !c.Version.mainEqual(v) && len(c.Version.Prerelease) > 0 && len(v.Prerelease) > 0 && c.Op != "=" {
		return false
	}
	var b bool
	switch c.Op {
	case ">":
		b = v.GreaterThan(c.Version)
	case ">=":
		b = v.GreaterEqualThan(c.Version)
	case "<":
		b = v.LowerThan(c.Version)
	case "<=":
		b = v.LowerEqualThan(c.Version)
	default:
		b = v.Equal(c.Version)
	}
	return b
}

func (c Comparator) String() string {
	return c.Op + c.Version.String()
}

func NewRange(str string) Range {
	s := strings.Split(str, " || ")
	c := Range{}
	for _, v := range s {
		c = append(c, NewComparatorSet(v))
	}
	return c
}

func NewComparatorSet(str string) ComparatorSet {
	re := regexp.MustCompile(SEMVER + " - " + SEMVER)
	c := ComparatorSet{}
	m := re.FindAllStringSubmatch(str, -1)
	if len(m) > 0 {
		for _, v := range m {
			str = strings.Replace(str, v[0], "", -1)
			high, low := parseHyphen(v[0])
			c = append(c, low)
			c = append(c, high)
		}
	}
	if len(strings.TrimSpace(str)) == 0 {
		return c
	}
	re1 := regexp.MustCompile(`([>=<^~]+)?(\s+)?(v)?` + SEMVER)
	m1 := re1.FindAllStringSubmatch(str, -1)
	if len(m1) > 0 {
		for _, v := range m1 {
			var op string
			ver := v[0]
			if len(v[1]) > 0 {
				op = v[1]
				ver = strings.Replace(ver, v[1], "", 1)
			}
			ver = strings.TrimPrefix(strings.TrimSpace(ver), "v")
			switch op {
			case "~":
				low, high := parseTilde(ver)
				c = append(c, low)
				c = append(c, high)
			case "^":
				low, high := parseCaret(ver)
				c = append(c, low)
				c = append(c, high)
			case "":
				ver = strings.Replace(ver, "X", "x", -1)
				if complete(ver) < 3 || strings.Contains(ver, ".x") {
					low, high := parseX(ver)
					c = append(c, low)
					if !high.IsNil() {
						c = append(c, high)
					}
				} else {
					c = append(c, NewComparator("=", ver))
				}
			default:
				c = append(c, NewComparator(op, ver))
			}
		}
	}
	return c.Intersect()
}

func NewComparator(op, ver string) Comparator {
	return Comparator{op, NewSemver(ver)}
}

func parseHyphen(str string) (Comparator, Comparator) {
	s := strings.Split(str, " - ")
	low := s[0]
	high := s[1]
	l1 := complete(low)
	l2 := complete(high)
	if l1 < 3 {
		low += strings.Repeat(".0", 3-l1)
	}
	op := "<="
	if l2 < 3 {
		high = incLastNum(high)
		high += strings.Repeat(".0", 3-l2)
		op = "<"
	}
	return NewComparator(">=", low), NewComparator(op, high)
}

func parseX(str string) (Comparator, Comparator) {
	if str == "*" {
		return NewComparator(">=", "0"), Comparator{}
	}
	str = strings.Replace(str, ".x", "", -1)
	str1 := incLastNum(str)
	idx := complete(str)
	str += strings.Repeat(".0", 3-idx)
	str1 += strings.Repeat(".0", 3-idx)
	return NewComparator(">=", str), NewComparator("<", str1)
}

func parseTilde(str string) (Comparator, Comparator) {
	idx := complete(str)
	i := 1
	if idx == 1 {
		i = 0
	}
	str1 := stripPrerelease(str)
	s1 := strings.Split(str1, ".")
	s1[i] = incLastNum(s1[i])
	if i != len(s1)-1 {
		for j := i + 1; j < len(s1); j++ {
			s1[j] = "0"
		}
	}
	str2 := strings.Join(s1, ".")
	if idx < 3 {
		str += strings.Repeat(".0", 3-idx)
		str2 += strings.Repeat(".0", 3-idx)
	}
	return NewComparator(">=", str), NewComparator("<", str2)
}

func parseCaret(str string) (Comparator, Comparator) {
	str1 := strings.Replace(stripPrerelease(str), ".x", "", -1)
	pre := strings.Replace(str, stripPrerelease(str), "", 1)
	s := strings.Split(str1, ".")
	idx := len(s) - 1
	for i, v := range s {
		if v != "0" {
			idx = i
			break
		}
	}
	s[idx] = incLastNum(s[idx])
	if idx != len(s)-1 {
		for j := idx + 1; j < len(s); j++ {
			s[j] = "0"
		}
	}
	str2 := strings.Join(s, ".")
	if len(s) < 3 {
		str1 += strings.Repeat(".0", 3-len(s))
		str2 += strings.Repeat(".0", 3-len(s))
	}
	return NewComparator(">=", str1+pre), NewComparator("<", str2)
}

func incLastNum(str string) string {
	s := strings.Split(str, ".")
	j, _ := strconv.Atoi(s[len(s)-1])
	j += 1
	s[len(s)-1] = strconv.Itoa(j)
	return strings.Join(s, ".")
}

func stripPrerelease(str string) string {
	idx := strings.Index(str, "-")
	if idx < 0 {
		idx = len(str)
	}
	return str[:idx]
}

func complete(str string) int {
	str = stripPrerelease(str)
	return len(strings.Split(str, "."))
}
