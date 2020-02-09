package main

import (
	//"reflect"
	"testing"
)

/*func Test_dedupeParents(t *testing.T) {
	var r Parents
	brothers := map[string]struct{}{}
	r = append(r, Parent{"root", brothers})
	r = append(r, Parent{"rimraf@1.0.0", brothers})
	r = append(r, Parent{"wrappy@1.0.0", brothers})
	o := make(Parents, len(r))
	n := make(Parents, len(r))
	copy(o, r)
	copy(n, r)
	r = append(r, Parent{})
	o = append(o, Parent{"A", brothers})
	n = append(n, Parent{"B", brothers})
	n = append(n, Parent{"C", brothers})
	testResult := dedupeParents(o, n)
	if reflect.DeepEqual(testResult, r) {
		t.Log("Test passed")
	} else {
		t.Errorf("dedupeParents() failed with result %v, should be %v", testResult, r)
	}
}*/
