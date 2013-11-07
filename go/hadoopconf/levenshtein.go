//package levenshtein
package main

/*import (
	"github.com/elazarl/hadoophelpers/go/lib/table"
	"strconv"
)*/

var (
	DeleteCost = 5 * 10
	ReplaceCost = 10 * 10
	AddCost    = 1
)

// translated from wikipedia Levenshtein Distance code snippet

// LevenshteinDistance calculate how many transformations do we
// need to apply to t, in order to make it an s.
// For example LevenshteinDistance("a", "b") = 1*ReplaceCost, since
// since we need one replace action to get from "b" to "a",
// LevenshteinDistance("a", "ab") is 1*DeleteCost since we need to
// delete one character to get from "ab" to "b".
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
    /*tbl := table.New(len(t)+2)
    headers := []string{"", ""}
    for _, r := range t {
	    headers = append(headers, string(r))
    }
    tbl.Add(headers...)*/

    // create two work vectors of integer distances
    v0, v1 := make([]int, len(t)+1), make([]int, len(t)+1)

    // initialize v0 (the previous row of distances)
    // this row is A[0][i]: edit distance for an empty s
    // the distance is just the number of characters to delete from t
    for i := range v0 {
	    v0[i] = i*DeleteCost
    }

    //printRow(tbl, ' ', v0)
    for i := range s {
        // calculate v1 (current row distances) from the previous row v0

        // first element of v1 is A[i+1][0]
        //   edit distance is delete (i+1) chars from s to match empty t
        v1[0] = (i + 1) * AddCost

        // use formula to fill in the rest of the row
	for j := range t {
		cost := ReplaceCost
		if s[i] == t[j] {
			cost = 0
		}
		v1[j + 1] = min(v1[j]+DeleteCost, v0[j+1]+AddCost, v0[j]+cost)
	}

        // copy v1 (current row) to v0 (previous row) for next iteration
	for j := range v0 {
		v0[j] = v1[j]
	}
	//printRow(tbl, rune(s[i]), v0)
    }
    //println(tbl.String())

    return v1[len(t)]
}

/*func printRow(tbl *table.Table, r rune, v0 []int) {
	row := []string{string(r)}
	for _, e := range v0 {
		row = append(row, strconv.Itoa(e))
	}
	tbl.Add(row...)
}*/
