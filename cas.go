//go:generate go tool yacc -p Calc -o interp.go interp.y
//go:generate golex -o tokenizer.go tokenizer.l

package cas

import "fmt"

type EvalState struct {
	defined map[string]Ex
}

func NewEvalState() *EvalState {
	var es EvalState
	es.defined = make(map[string]Ex)
	return &es
}

// Ex stands for Expression. Most structs should implement this
type Ex interface {
	Eval(es *EvalState) Ex
	Replace(r *Rule, es *EvalState) Ex
	ToString() string
	IsEqual(b Ex, es *EvalState) string
	IsSameQ(b Ex, es *EvalState) bool
	IsMatchQ(b Ex, es *EvalState) bool
	DeepCopy() Ex
}

// Some utility functions that span multiple files

func CommutativeIsEqual(components []Ex, other_components []Ex, es *EvalState) string {
	if len(components) != len(other_components) {
		return "EQUAL_FALSE"
	}
	matched := make(map[int]struct{})
	for _, e1 := range components {
		foundmatch := false
		for j, e2 := range other_components {
			_, taken := matched[j]
			if taken {
				continue
			}
			res := e1.IsEqual(e2, es)
			switch res {
			case "EQUAL_FALSE":
			case "EQUAL_TRUE":
				matched[j] = struct{}{}
				foundmatch = true
			case "EQUAL_UNK":
			}
			if foundmatch {
				break
			}
		}
		if !foundmatch {
			return "EQUAL_UNK"
		}
	}
	return "EQUAL_TRUE"
}

func FunctionIsEqual(components []Ex, other_components []Ex, es *EvalState) string {
	if len(components) != len(other_components) {
		return "EQUAL_FALSE"
	}
	for i := range components {
		res := components[i].IsEqual(other_components[i], es)
		switch res {
		case "EQUAL_FALSE":
			return "EQUAL_UNK"
		case "EQUAL_TRUE":
		case "EQUAL_UNK":
			return "EQUAL_UNK"
		}
	}
	return "EQUAL_TRUE"
}

func FunctionIsSameQ(components []Ex, other_components []Ex, es *EvalState) bool {
	if len(components) != len(other_components) {
		return false
	}
	for i := range components {
		res := components[i].IsSameQ(other_components[i], es)
		if !res {
			return false
		}
	}
	return true
}

func IterableReplace(components *[]Ex, r *Rule, es *EvalState) {
	for i := range *components {
		if (*components)[i].IsMatchQ(r.Lhs, es) {
			(*components)[i] = r.Rhs.DeepCopy()
		}
	}
}

func permutations(iterable []int, r int) [][]int {
	res := make([][]int, 0)
	pool := iterable
	n := len(pool)

	if r > n {
		return res
	}

	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	cycles := make([]int, r)
	for i := range cycles {
		cycles[i] = n - i
	}

	result := make([]int, r)
	for i, el := range indices[:r] {
		result[i] = pool[el]
	}

	c := make([]int, len(result))
	copy(c, result)
	res = append(res, c)

	for n > 0 {
		i := r - 1
		for ; i >= 0; i -= 1 {
			cycles[i] -= 1
			if cycles[i] == 0 {
				index := indices[i]
				for j := i; j < n-1; j += 1 {
					indices[j] = indices[j+1]
				}
				indices[n-1] = index
				cycles[i] = n - i
			} else {
				j := cycles[i]
				indices[i], indices[n-j] = indices[n-j], indices[i]

				for k := i; k < r; k += 1 {
					result[k] = pool[indices[k]]
				}

				c := make([]int, len(result))
				copy(c, result)
				res = append(res, c)

				break
			}
		}

		if i < 0 {
			return res
		}

	}
	return res

}

func CommutativeReplace(components *[]Ex, lhs_components []Ex, rhs Ex, es *EvalState) {
	toPermute := make([]int, len(lhs_components))
	for i := range toPermute { toPermute[i] = i }
	perms := permutations(toPermute, len(lhs_components))
	fmt.Println(perms)

	if len(lhs_components) == 2 {
		// Choose to match first comp on LHS first
		for i := range *components {
			if (*components)[i].IsMatchQ(lhs_components[0], es) {
				// Should I be deep copying?
				//popped Ex
				popped := (*components)[i]
				*components = append((*components)[:i], (*components)[i+1:]...)
				//var p *Plus = &Plus{*components}
				//fmt.Println(p.ToString())
				for j := range *components {
					if (*components)[j].IsMatchQ(lhs_components[1], es) {
						// Reached end of line
						*components = append((*components)[:j], (*components)[j+1:]...)
						*components = append(*components, []Ex{rhs}...)
						return
					}
				}
				*components = append((*components)[:i], append([]Ex{popped}, (*components)[i:]...)...)
				//var pn *Plus = &Plus{*components}
				//fmt.Println(pn.ToString())
			}
		}
		// Choose to match first comp on LHS first
		for i := range *components {
			if (*components)[i].IsMatchQ(lhs_components[1], es) {
				// Should I be deep copying?
				//popped Ex
				popped := (*components)[i]
				*components = append((*components)[:i], (*components)[i+1:]...)
				//var p *Plus = &Plus{*components}
				//fmt.Println(p.ToString())
				for j := range *components {
					if (*components)[j].IsMatchQ(lhs_components[0], es) {
						// Reached end of line
						*components = append((*components)[:j], (*components)[j+1:]...)
						*components = append(*components, []Ex{rhs}...)
						return
					}
				}
				*components = append((*components)[:i], append([]Ex{popped}, (*components)[i:]...)...)
				//var pn *Plus = &Plus{*components}
				//fmt.Println(pn.ToString())
			}
		}
	}
}
