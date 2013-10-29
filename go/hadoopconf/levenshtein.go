//package levenshtein
package main

// translated from wikipedia Levenshtein Distance code snippet
func LevenshteinDistance(s, t string) int {
    min := func(x int, xs ...int) int {
	    for _, v := range xs {
		    if x > v {
			    x = v
		    }
	    }
	    return x
    }
    // degenerate cases
    if s == t {
	    return 0
    }
    if len(s) == 0 {
	    return len(t)
    }
    if len(t) == 0 {
	    return len(s)
    }

    // create two work vectors of integer distances
    v0, v1 := make([]int, len(t)+1), make([]int, len(t)+1)

    // initialize v0 (the previous row of distances)
    // this row is A[0][i]: edit distance for an empty s
    // the distance is just the number of characters to delete from t
    for i := range v0 {
	    v0[i] = i
    }

    for i := range s {
        // calculate v1 (current row distances) from the previous row v0

        // first element of v1 is A[i+1][0]
        //   edit distance is delete (i+1) chars from s to match empty t
        v1[0] = i + 1;

        // use formula to fill in the rest of the row
	for j := range t {
		cost := 1
		if s[i] == t[j] {
			cost = 0
		}
		v1[j + 1] = min(v1[j]+1, v0[j+1]+1, v0[j]+cost)
	}

        // copy v1 (current row) to v0 (previous row) for next iteration
	for j := range v0 {
		v0[j] = v1[j]
	}
    }

    return v1[len(t)]
}
