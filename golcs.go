// package lcs provides functions to calculate Longest Common Subsequence (LCS)
// values from two arbitrary arrays.
package golcs

import (
	"context"
	"reflect"
)

// LCS is the interface to calculate the LCS of two arrays.
type LCS interface {
	// Values calculates the LCS value of the two arrays.
	Values() (values []interface{})
	// ValuesContext is a context aware version of Values()
	ValuesContext(ctx context.Context) ([]interface{}, error)
	// IndexPairs calculates paris of indices which have the same value in LCS.
	IndexPairs() (pairs []IndexPair)
	// IndexPairsContext is a context aware version of IndexPairs()
	IndexPairsContext(ctx context.Context) ([]IndexPair, error)
	// Length calculates the length of the LCS.
	Length() (length int)
	// LengthContext is a context aware version of Length()
	LengthContext(ctx context.Context) (int, error)
	// Left returns one of the two arrays to be compared.
	Left() []interface{}
	// Right returns the other of the two arrays to be compared.
	Right() []interface{}
}

// IndexPair represents a pair of indices in the Left and Right arrays found in the LCS value.
type IndexPair struct {
	Left  int
	Right int
}

type lcs struct {
	left  []interface{}
	right []interface{}
	/* for caching */
	table      [][]int
	indexPairs []IndexPair
	values     []interface{}
}

// New creates a new LCS calculator from two arrays.
func New(left, right []interface{}) LCS {
	return &lcs{
		left:       left,
		right:      right,
		table:      nil,
		indexPairs: nil,
		values:     nil,
	}
}

// Table implements LCS.Table()
func (lcs *lcs) Table() [][]int {
	table, _ := lcs.TableContext(context.Background())
	return table
}

// TableContext Table implements LCS.TableContext()
func (lcs *lcs) TableContext(ctx context.Context) ([][]int, error) {
	if lcs.table != nil {
		return lcs.table, nil
	}

	sizeX := len(lcs.left) + 1
	sizeY := len(lcs.right) + 1

	table := make([][]int, sizeX)
	for x := 0; x < sizeX; x++ {
		table[x] = make([]int, sizeY)
	}

	for y := 1; y < sizeY; y++ {
		select { // check in each y to save some time
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// nop
		}
		for x := 1; x < sizeX; x++ {
			increment := 0
			if reflect.DeepEqual(lcs.left[x-1], lcs.right[y-1]) {
				increment = 1
			}
			table[x][y] = max(table[x-1][y-1]+increment, table[x-1][y], table[x][y-1])
		}
	}

	lcs.table = table
	return table, nil
}

// Length Table implements LCS.Length()
func (lcs *lcs) Length() int {
	length, _ := lcs.LengthContext(context.Background())
	return length
}

// LengthContext Table implements LCS.LengthContext()
func (lcs *lcs) LengthContext(ctx context.Context) (int, error) {
	if len(lcs.right) > len(lcs.left) {
		lcs.left, lcs.right = lcs.right, lcs.left
	}
	return lcs.lengthContext(ctx)
}

func (lcs *lcs) lengthContext(ctx context.Context) (int, error) {
	m := len(lcs.left)
	n := len(lcs.right)

	// allocate storage for one-dimensional array `curr`
	prev := 0
	curr := make([]int, n+1)

	// fill the lookup table in a bottom-up manner
	for i := 0; i <= m; i++ {
		select { // check in each y to save some time
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			// nop
		}
		prev = curr[0]
		for j := 0; j <= n; j++ {
			backup := curr[j]
			if i == 0 || j == 0 {
				curr[j] = 0
			} else if reflect.DeepEqual(lcs.left[i-1], lcs.right[j-1]) {
				// if the current character of `X` and `Y` matches
				curr[j] = prev + 1
			} else {
				// otherwise, if the current character of `X` and `Y` don't match
				curr[j] = max(curr[j], curr[j-1])
			}
			prev = backup
		}
	}
	// LCS will be the last entry in the lookup table
	return curr[n], nil
}

// IndexPairs Table implements LCS.IndexPairs()
func (lcs *lcs) IndexPairs() []IndexPair {
	pairs, _ := lcs.IndexPairsContext(context.Background())
	return pairs
}

// IndexPairsContext Table implements LCS.IndexPairsContext()
func (lcs *lcs) IndexPairsContext(ctx context.Context) ([]IndexPair, error) {
	if lcs.indexPairs != nil {
		return lcs.indexPairs, nil
	}

	table, err := lcs.TableContext(ctx)
	if err != nil {
		return nil, err
	}

	pairs := make([]IndexPair, table[len(table)-1][len(table[0])-1])
	for x, y := len(lcs.left), len(lcs.right); x > 0 && y > 0; {
		if reflect.DeepEqual(lcs.left[x-1], lcs.right[y-1]) {
			pairs[table[x][y]-1] = IndexPair{Left: x - 1, Right: y - 1}
			x--
			y--
		} else {
			if table[x-1][y] >= table[x][y-1] {
				x--
			} else {
				y--
			}
		}
	}

	lcs.indexPairs = pairs
	return pairs, nil
}

// Values Table implements LCS.Values()
func (lcs *lcs) Values() []interface{} {
	values, _ := lcs.ValuesContext(context.Background())
	return values
}

// ValuesContext Table implements LCS.ValuesContext()
func (lcs *lcs) ValuesContext(ctx context.Context) ([]interface{}, error) {
	if lcs.values != nil {
		return lcs.values, nil
	}

	pairs, err := lcs.IndexPairsContext(ctx)
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(pairs))
	for i, pair := range pairs {
		values[i] = lcs.left[pair.Left]
	}
	lcs.values = values

	return values, nil
}

// Left Table implements LCS.Left()
func (lcs *lcs) Left() []interface{} {
	return lcs.left
}

// Right Table implements LCS.Right()
func (lcs *lcs) Right() []interface{} {
	return lcs.right
}

func max(first int, rest ...int) int {
	maxValue := first
	for _, value := range rest {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}
