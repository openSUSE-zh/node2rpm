package main

// ParentTree a place holding all nodes in the tree now with its parents
// used to compute unified parents for the unified package, or one package
// with a specified version may occur everywhere in the tree. (Dedupe)
type ParentTree map[string]Parents

// Parents a group of Parent, its indexing is important. smaller indexing means higher level of parent
type Parents []Parent

// Contains if a package is provided by one of its direct parents or direct parents' brothers
func (p Parents) Contains(s string) bool {
	for i := 0; i < len(p)-1; i++ {
		if p[i].Name == s {
			return true
		}
		if _, ok := p[i].Brothers[s]; ok {
			return true
		}
	}
	return false
}

// DirectParents the direct parents of a package
func (p Parents) DirectParents() []string {
	a := []string{}
	for _, v := range p {
		a = append(a, v.Name)
	}
	return a
}

// Parent Parent contains the name of the direct parent and the direct parent's counterparts as brothers
type Parent struct {
	Name     string
	Brothers map[string]struct{}
}

// dedupeParents find intersection of two Parents
func dedupeParents(o, n Parents, t Tree) Parents {
	var low, high Parents
	if len(o)-len(n) >= 0 {
		low = n
		high = o
	} else {
		low = o
		high = n
	}
	idx := 0
	// "-2" because the last element is always different
	for i := len(low) - 2; i >= 0; i-- {
		if low[i].Name == high[i].Name {
			idx = i
			break
		}
	}
	// the last element' brothers should be the new direct parent's existing dependencies.
	brothers := map[string]struct{}{}
	for k := range t.FindChild(idx, low) {
		brothers[k] = struct{}{}
	}
	p := make([]Parent, idx+1, idx+1)
	copy(p, low)
	return append(p, Parent{low[len(low)-1].Name, brothers})
}
