package day10

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrNoMachines      = errors.New("no machines provided")
	ErrInvalidMachine  = errors.New("invalid machine description")
	ErrNoRequirements  = errors.New("no joltage requirements")
	ErrUnsolvable      = errors.New("machine cannot be configured with given buttons")
	ErrProblemTooLarge = errors.New("problem too large for available algorithms")
	ErrOverflow        = errors.New("overflow")
)

type Result struct {
	TotalPresses int64
}

type incEntry struct {
	idx   int
	delta uint16
}

type button struct {
	entries []incEntry
}

type bfsButton struct {
	entries []incEntry
	deltaID uint64
}

// Compute reads one machine per line, finds the minimum number of button presses
// required to configure its joltage counters to the target requirements, and
// returns the sum across all machines.
//
// Indicator light diagrams in `[...]` are ignored.
func Compute(r io.Reader) (Result, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	type job struct {
		lineNo int
		text   string
	}
	var jobs []job
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		jobs = append(jobs, job{lineNo: lineNo, text: raw})
	}
	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	if len(jobs) == 0 {
		return Result{}, ErrNoMachines
	}

	workerCount := runtime.NumCPU() / 2
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > 6 {
		workerCount = 6
	}

	type result struct {
		lineNo  int
		presses int
		err     error
	}

	jobCh := make(chan job)
	resCh := make(chan result, workerCount)

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for j := range jobCh {
				presses, err := minPressesForLine(j.text)
				resCh <- result{lineNo: j.lineNo, presses: presses, err: err}
			}
		}()
	}

	go func() {
		for _, j := range jobs {
			jobCh <- j
		}
		close(jobCh)
		wg.Wait()
		close(resCh)
	}()

	var total int64
	var firstErr *result
	for r := range resCh {
		if r.err != nil {
			if firstErr == nil || r.lineNo < firstErr.lineNo {
				tmp := r
				firstErr = &tmp
			}
			continue
		}
		total += int64(r.presses)
	}

	if firstErr != nil {
		return Result{}, fmt.Errorf("line %d: %w", firstErr.lineNo, firstErr.err)
	}
	return Result{TotalPresses: total}, nil
}

func minPressesForLine(line string) (int, error) {
	targets, buttons, err := parseMachine(line)
	if err != nil {
		return 0, err
	}
	buttons = normalizeButtons(buttons)

	allZero := true
	for _, t := range targets {
		if t != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return 0, nil
	}
	if len(buttons) == 0 {
		return 0, ErrUnsolvable
	}

	return minPresses(targets, buttons)
}

func normalizeButtons(buttons []button) []button {
	out := make([]button, 0, len(buttons))
	for _, b := range buttons {
		if len(b.entries) == 0 {
			continue
		}
		out = append(out, b)
	}
	return out
}

func parseMachine(line string) (targets []uint16, buttons []button, err error) {
	// Ignore anything in `[...]` (indicator lights) and focus on `{...}` and `(...)`.
	jOpen := strings.IndexByte(line, '{')
	jClose := strings.IndexByte(line, '}')
	if jOpen == -1 || jClose == -1 || jClose < jOpen {
		return nil, nil, ErrInvalidMachine
	}

	targets, err = parseTargets(line[jOpen+1 : jClose])
	if err != nil {
		return nil, nil, err
	}
	if len(targets) == 0 {
		return nil, nil, ErrNoRequirements
	}

	rest := line[:jOpen]
	for {
		l := strings.IndexByte(rest, '(')
		if l == -1 {
			break
		}
		r := strings.IndexByte(rest[l+1:], ')')
		if r == -1 {
			return nil, nil, ErrInvalidMachine
		}
		r = l + 1 + r
		content := strings.TrimSpace(rest[l+1 : r])
		btn, err := parseButton(content, len(targets))
		if err != nil {
			return nil, nil, err
		}
		buttons = append(buttons, btn)
		rest = rest[r+1:]
	}

	return targets, buttons, nil
}

func parseTargets(content string) ([]uint16, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrInvalidMachine
	}

	parts := strings.Split(content, ",")
	out := make([]uint16, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, ErrInvalidMachine
		}
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 || v > math.MaxUint16 {
			return nil, ErrInvalidMachine
		}
		out = append(out, uint16(v))
	}
	return out, nil
}

func parseButton(content string, counters int) (button, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return button{}, ErrInvalidMachine
	}

	parts := strings.Split(content, ",")
	counts := make(map[int]uint16, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return button{}, ErrInvalidMachine
		}
		idx, err := strconv.Atoi(p)
		if err != nil || idx < 0 || idx >= counters {
			return button{}, ErrInvalidMachine
		}
		counts[idx]++
	}

	entries := make([]incEntry, 0, len(counts))
	for idx, delta := range counts {
		entries = append(entries, incEntry{idx: idx, delta: delta})
	}
	sortIncEntries(entries)
	return button{entries: entries}, nil
}

func sortIncEntries(entries []incEntry) {
	for i := 1; i < len(entries); i++ {
		j := i
		for j > 0 && entries[j-1].idx > entries[j].idx {
			entries[j-1], entries[j] = entries[j], entries[j-1]
			j--
		}
	}
}

func minPressesBFS(targets []uint16, buttons []button) (int, error) {
	mults, bases, stateCount, goalID, err := buildRadix(targets)
	if err != nil {
		return 0, err
	}

	// Avoid allocating gigantic arrays.
	const maxArrayStates = 30_000_000
	if stateCount > maxArrayStates {
		return 0, ErrProblemTooLarge
	}

	bfsButtons, err := compileButtons(buttons, mults, bases)
	if err != nil {
		return 0, err
	}
	if len(bfsButtons) == 0 {
		return 0, ErrUnsolvable
	}

	dist := make([]int32, stateCount)
	for i := range dist {
		dist[i] = -1
	}
	dist[0] = 0

	queue := make([]uint32, 0, 1024)
	queue = append(queue, 0)
	head := 0

	for head < len(queue) {
		state := uint64(queue[head])
		head++

		if state == goalID {
			return int(dist[state]), nil
		}

		nextDist := dist[state] + 1
		for _, b := range bfsButtons {
			next, ok := applyButton(state, b, mults, bases)
			if !ok {
				continue
			}
			if dist[next] != -1 {
				continue
			}
			dist[next] = nextDist
			queue = append(queue, uint32(next))
		}
	}

	return 0, ErrUnsolvable
}

func minPresses(targets []uint16, buttons []button) (int, error) {
	if presses, ok, err := minPressesByLinearSolve(targets, buttons); err != nil {
		return 0, err
	} else if ok {
		return presses, nil
	}

	if presses, ok, err := minPressesByRREFEnumeration(targets, buttons); err != nil {
		return 0, err
	} else if ok {
		return presses, nil
	}

	// Try the dense mixed-radix BFS when it fits (fast, array-backed).
	if canUseDenseBFS(targets, 30_000_000) {
		return minPressesBFS(targets, buttons)
	}

	// Fall back to a sparse shortest-path search that does not require a full
	// (prod(target_i+1)) state array.
	return minPressesBidirectionalSparse(targets, buttons)
}

func minPressesByLinearSolve(targets []uint16, buttons []button) (presses int, ok bool, err error) {
	m := len(targets)
	k := len(buttons)
	if m == 0 || k == 0 {
		return 0, false, nil
	}
	if k < m {
		return 0, false, nil
	}

	// This approach enumerates free variables, so only attempt it when the number
	// of free variables is small enough to be practical.
	freeDim := k - m
	if freeDim > 6 {
		return 0, false, nil
	}

	A := buildMatrix(m, k, buttons)
	ubs := buttonUpperBounds(A, targets)

	pivots, free, found := choosePivotColumns(A, ubs, m, k)
	if !found {
		return 0, false, nil
	}

	estimate := estimateEnumerationSize(free, ubs)
	// If enumeration is too large, fall back to graph search.
	if estimate > 5_000_000 {
		return 0, false, nil
	}

	B := buildSubmatrix(A, pivots)
	invB, err := invertMatrixInt64(B)
	if err != nil {
		return 0, false, nil
	}

	// Order free variables by increasing upper bound for smaller branching.
	free = append([]int(nil), free...)
	for i := 1; i < len(free); i++ {
		j := i
		for j > 0 && ubs[free[j-1]] > ubs[free[j]] {
			free[j-1], free[j] = free[j], free[j-1]
			j--
		}
	}

	rhs := make([]int64, m)
	for i, t := range targets {
		rhs[i] = int64(t)
	}

	best := int64(-1)
	var dfs func(pos int, currentPresses int64)

	dfs = func(pos int, currentPresses int64) {
		if best != -1 && currentPresses >= best {
			return
		}
		if pos == len(free) {
			pivotSum, ok := pivotPressesSum(invB, rhs)
			if !ok {
				return
			}
			total := currentPresses + pivotSum
			if total < 0 {
				return
			}
			if best == -1 || total < best {
				best = total
			}
			return
		}

		col := free[pos]
		ub := ubs[col]

		// Try smaller values first to find good upper bounds early.
		for v := int64(0); v <= ub; v++ {
			if v > 0 {
				for r := 0; r < m; r++ {
					rhs[r] -= A[r][col] * v
					if rhs[r] < 0 {
						// Undo and stop increasing v: larger v will only decrease rhs further.
						for rr := 0; rr <= r; rr++ {
							rhs[rr] += A[rr][col] * v
						}
						goto nextColumn
					}
				}
			}

			dfs(pos+1, currentPresses+v)

			if v > 0 {
				for r := 0; r < m; r++ {
					rhs[r] += A[r][col] * v
				}
			}
		}

	nextColumn:
		return
	}

	dfs(0, 0)
	if best == -1 {
		return 0, false, nil
	}
	if best > int64(math.MaxInt) {
		return 0, false, ErrOverflow
	}
	return int(best), true, nil
}

func minPressesByRREFEnumeration(targets []uint16, buttons []button) (presses int, ok bool, err error) {
	m := len(targets)
	k := len(buttons)
	if m == 0 || k == 0 {
		return 0, false, nil
	}

	A := buildMatrix(m, k, buttons)
	ubs := buttonUpperBounds(A, targets)

	mat, pivots, err := rrefAugmented(A, targets)
	if err != nil {
		return 0, false, nil
	}

	pivotSet := make(map[int]struct{}, len(pivots))
	for _, pc := range pivots {
		pivotSet[pc] = struct{}{}
	}
	free := make([]int, 0, k-len(pivots))
	for c := 0; c < k; c++ {
		if _, isPivot := pivotSet[c]; !isPivot {
			free = append(free, c)
		}
	}

	// Only enumerate when free dimension is small.
	if len(free) > 6 {
		return 0, false, nil
	}
	if estimate := estimateEnumerationSize(free, ubs); estimate > 10_000_000 {
		return 0, false, nil
	}

	// Order free variables by increasing upper bound for smaller branching.
	for i := 1; i < len(free); i++ {
		j := i
		for j > 0 && ubs[free[j-1]] > ubs[free[j]] {
			free[j-1], free[j] = free[j], free[j-1]
			j--
		}
	}

	freeVals := make([]int64, len(free))
	values := make([]int64, k)

	type pivot struct {
		row int
		col int
	}
	pivotRows := make([]pivot, 0, len(pivots))
	// In rrefAugmented, pivots are recorded in row order (one per pivot row).
	for row, col := range pivots {
		pivotRows = append(pivotRows, pivot{row: row, col: col})
	}

	eval := func() (int64, bool) {
		for i := range values {
			values[i] = 0
		}
		for i, col := range free {
			values[col] = freeVals[i]
		}

		for _, pr := range pivotRows {
			r := pr.row
			pivotCol := pr.col

			var acc big.Rat
			acc.Set(&mat[r][k]) // rhs
			for i, col := range free {
				if mat[r][col].Sign() == 0 {
					continue
				}
				var term big.Rat
				var fv big.Rat
				fv.SetInt64(freeVals[i])
				term.Mul(&mat[r][col], &fv)
				acc.Sub(&acc, &term)
			}

			if !acc.IsInt() || acc.Sign() < 0 {
				return 0, false
			}
			num := acc.Num()
			if num.BitLen() > 62 {
				return 0, false
			}
			v := num.Int64()
			if v > ubs[pivotCol] {
				return 0, false
			}
			values[pivotCol] = v
		}

		var sum int64
		for _, v := range values {
			if sum > math.MaxInt64-v {
				return 0, false
			}
			sum += v
		}
		return sum, true
	}

	best := int64(-1)
	var dfs func(pos int, currentSum int64)
	dfs = func(pos int, currentSum int64) {
		if best != -1 && currentSum >= best {
			return
		}
		if pos == len(free) {
			sum, ok := eval()
			if !ok {
				return
			}
			if best == -1 || sum < best {
				best = sum
			}
			return
		}

		col := free[pos]
		ub := ubs[col]
		for v := int64(0); v <= ub; v++ {
			if best != -1 && currentSum+v >= best {
				break
			}
			freeVals[pos] = v
			dfs(pos+1, currentSum+v)
		}
	}

	dfs(0, 0)
	if best == -1 {
		return 0, false, nil
	}
	if best > int64(math.MaxInt) {
		return 0, false, ErrOverflow
	}
	return int(best), true, nil
}

func buildMatrix(m, k int, buttons []button) [][]int64 {
	A := make([][]int64, m)
	for i := 0; i < m; i++ {
		A[i] = make([]int64, k)
	}
	for j, b := range buttons {
		for _, e := range b.entries {
			A[e.idx][j] = int64(e.delta)
		}
	}
	return A
}

func buttonUpperBounds(A [][]int64, targets []uint16) []int64 {
	m := len(A)
	k := len(A[0])
	ubs := make([]int64, k)
	for j := 0; j < k; j++ {
		ub := int64(math.MaxInt64)
		touches := false
		for i := 0; i < m; i++ {
			if A[i][j] == 0 {
				continue
			}
			touches = true
			maxHere := int64(targets[i]) / A[i][j]
			if maxHere < ub {
				ub = maxHere
			}
		}
		if !touches {
			ub = 0
		}
		ubs[j] = ub
	}
	return ubs
}

func estimateEnumerationSize(free []int, ubs []int64) uint64 {
	product := uint64(1)
	for _, col := range free {
		choices := uint64(ubs[col]) + 1
		if choices == 0 {
			return math.MaxUint64
		}
		if product > math.MaxUint64/choices {
			return math.MaxUint64
		}
		product *= choices
	}
	return product
}

func choosePivotColumns(A [][]int64, ubs []int64, m, k int) (pivots []int, free []int, ok bool) {
	// For moderate k, brute force pivot selection to minimize the enumeration size
	// of the remaining (free) variables.
	if k <= 20 && m <= 16 {
		cols := make([]int, k)
		for i := 0; i < k; i++ {
			cols[i] = i
		}

		bestCost := uint64(math.MaxUint64)
		var bestPivots []int

		var comb []int
		var rec func(start int)
		rec = func(start int) {
			if len(comb) == m {
				if !isInvertibleModSubmatrix(A, comb, 1_000_000_007) && !isInvertibleModSubmatrix(A, comb, 1_000_000_009) {
					return
				}
				isPivot := make([]bool, k)
				for _, c := range comb {
					isPivot[c] = true
				}
				var freeCols []int
				for c := 0; c < k; c++ {
					if !isPivot[c] {
						freeCols = append(freeCols, c)
					}
				}
				cost := estimateEnumerationSize(freeCols, ubs)
				if cost < bestCost {
					bestCost = cost
					bestPivots = append([]int(nil), comb...)
				}
				return
			}
			if start >= k {
				return
			}
			need := m - len(comb)
			if k-start < need {
				return
			}
			for c := start; c < k; c++ {
				comb = append(comb, cols[c])
				rec(c + 1)
				comb = comb[:len(comb)-1]
			}
		}
		rec(0)

		if bestPivots == nil {
			return nil, nil, false
		}

		isPivot := make([]bool, k)
		for _, c := range bestPivots {
			isPivot[c] = true
		}
		var freeCols []int
		for c := 0; c < k; c++ {
			if !isPivot[c] {
				freeCols = append(freeCols, c)
			}
		}
		return bestPivots, freeCols, true
	}

	// Fallback heuristic: pick pivot columns from larger-UB columns first so that
	// the remaining free variables tend to have smaller ranges.
	order := make([]int, k)
	for i := 0; i < k; i++ {
		order[i] = i
	}
	for i := 1; i < k; i++ {
		j := i
		for j > 0 && ubs[order[j-1]] < ubs[order[j]] {
			order[j-1], order[j] = order[j], order[j-1]
			j--
		}
	}

	pivots = selectPivotColumnsMod(A, order, 1_000_000_007)
	if len(pivots) != m {
		pivots = selectPivotColumnsMod(A, order, 1_000_000_009)
	}
	if len(pivots) != m {
		return nil, nil, false
	}

	isPivot := make([]bool, k)
	for _, c := range pivots {
		isPivot[c] = true
	}
	for c := 0; c < k; c++ {
		if !isPivot[c] {
			free = append(free, c)
		}
	}
	return pivots, free, true
}

func selectPivotColumnsMod(A [][]int64, colOrder []int, mod int64) []int {
	m := len(A)
	k := len(colOrder)
	mat := make([][]int64, m)
	for i := 0; i < m; i++ {
		mat[i] = make([]int64, k)
		for j := 0; j < k; j++ {
			v := A[i][colOrder[j]] % mod
			if v < 0 {
				v += mod
			}
			mat[i][j] = v
		}
	}

	pivots := make([]int, 0, m)
	row := 0
	for col := 0; col < k && row < m; col++ {
		pivot := -1
		for r := row; r < m; r++ {
			if mat[r][col] != 0 {
				pivot = r
				break
			}
		}
		if pivot == -1 {
			continue
		}
		mat[row], mat[pivot] = mat[pivot], mat[row]

		inv := modInv(mat[row][col], mod)
		for r := row + 1; r < m; r++ {
			if mat[r][col] == 0 {
				continue
			}
			factor := (mat[r][col] * inv) % mod
			for c := col; c < k; c++ {
				mat[r][c] = (mat[r][c] - factor*mat[row][c]) % mod
				if mat[r][c] < 0 {
					mat[r][c] += mod
				}
			}
		}

		pivots = append(pivots, colOrder[col])
		row++
	}
	return pivots
}

func isInvertibleModSubmatrix(A [][]int64, cols []int, mod int64) bool {
	m := len(A)
	if len(cols) != m {
		return false
	}
	mat := make([][]int64, m)
	for i := 0; i < m; i++ {
		mat[i] = make([]int64, m)
		for j := 0; j < m; j++ {
			v := A[i][cols[j]] % mod
			if v < 0 {
				v += mod
			}
			mat[i][j] = v
		}
	}

	row := 0
	for col := 0; col < m && row < m; col++ {
		pivot := -1
		for r := row; r < m; r++ {
			if mat[r][col] != 0 {
				pivot = r
				break
			}
		}
		if pivot == -1 {
			return false
		}
		mat[row], mat[pivot] = mat[pivot], mat[row]
		inv := modInv(mat[row][col], mod)
		for r := row + 1; r < m; r++ {
			if mat[r][col] == 0 {
				continue
			}
			factor := (mat[r][col] * inv) % mod
			for c := col; c < m; c++ {
				mat[r][c] = (mat[r][c] - factor*mat[row][c]) % mod
				if mat[r][c] < 0 {
					mat[r][c] += mod
				}
			}
		}
		row++
	}
	return row == m
}

func modInv(a, mod int64) int64 {
	return modPow(a, mod-2, mod)
}

func modPow(a, e, mod int64) int64 {
	a %= mod
	if a < 0 {
		a += mod
	}
	res := int64(1)
	for e > 0 {
		if e&1 == 1 {
			res = (res * a) % mod
		}
		a = (a * a) % mod
		e >>= 1
	}
	return res
}

func buildSubmatrix(A [][]int64, cols []int) [][]int64 {
	m := len(A)
	B := make([][]int64, m)
	for i := 0; i < m; i++ {
		B[i] = make([]int64, len(cols))
		for j, c := range cols {
			B[i][j] = A[i][c]
		}
	}
	return B
}

func invertMatrixInt64(B [][]int64) ([][]big.Rat, error) {
	m := len(B)
	if m == 0 {
		return nil, ErrInvalidMachine
	}
	if len(B[0]) != m {
		return nil, ErrInvalidMachine
	}

	aug := make([][]big.Rat, m)
	for i := 0; i < m; i++ {
		aug[i] = make([]big.Rat, 2*m)
		for j := 0; j < m; j++ {
			aug[i][j].SetInt64(B[i][j])
		}
		for j := 0; j < m; j++ {
			if i == j {
				aug[i][m+j].SetInt64(1)
			} else {
				aug[i][m+j].SetInt64(0)
			}
		}
	}

	for col := 0; col < m; col++ {
		pivot := -1
		for r := col; r < m; r++ {
			if aug[r][col].Sign() != 0 {
				pivot = r
				break
			}
		}
		if pivot == -1 {
			return nil, ErrUnsolvable
		}
		if pivot != col {
			aug[col], aug[pivot] = aug[pivot], aug[col]
		}

		var pivotVal big.Rat
		pivotVal.Set(&aug[col][col])
		for c := col; c < 2*m; c++ {
			aug[col][c].Quo(&aug[col][c], &pivotVal)
		}

		for r := 0; r < m; r++ {
			if r == col {
				continue
			}
			if aug[r][col].Sign() == 0 {
				continue
			}
			var factor big.Rat
			factor.Set(&aug[r][col])
			for c := col; c < 2*m; c++ {
				var prod big.Rat
				prod.Mul(&factor, &aug[col][c])
				aug[r][c].Sub(&aug[r][c], &prod)
			}
		}
	}

	inv := make([][]big.Rat, m)
	for i := 0; i < m; i++ {
		inv[i] = make([]big.Rat, m)
		for j := 0; j < m; j++ {
			inv[i][j].Set(&aug[i][m+j])
		}
	}
	return inv, nil
}

func pivotPressesSum(invB [][]big.Rat, rhs []int64) (int64, bool) {
	m := len(rhs)
	var sumPresses int64

	for r := 0; r < m; r++ {
		var acc big.Rat
		acc.SetInt64(0)
		for c := 0; c < m; c++ {
			if rhs[c] == 0 {
				continue
			}
			var rhsRat big.Rat
			rhsRat.SetInt64(rhs[c])
			var term big.Rat
			term.Mul(&invB[r][c], &rhsRat)
			acc.Add(&acc, &term)
		}

		if !acc.IsInt() {
			return 0, false
		}
		if acc.Sign() < 0 {
			return 0, false
		}
		num := acc.Num()
		if num.BitLen() > 62 {
			return 0, false
		}
		v := num.Int64()
		if sumPresses > math.MaxInt64-v {
			return 0, false
		}
		sumPresses += v
	}

	return sumPresses, true
}

func rrefAugmented(A [][]int64, targets []uint16) (mat [][]big.Rat, pivots []int, err error) {
	m := len(A)
	if m == 0 {
		return nil, nil, ErrInvalidMachine
	}
	k := len(A[0])

	mat = make([][]big.Rat, m)
	for r := 0; r < m; r++ {
		mat[r] = make([]big.Rat, k+1)
		for c := 0; c < k; c++ {
			mat[r][c].SetInt64(A[r][c])
		}
		mat[r][k].SetInt64(int64(targets[r]))
	}

	row := 0
	for col := 0; col < k && row < m; col++ {
		pivot := -1
		for r := row; r < m; r++ {
			if mat[r][col].Sign() != 0 {
				pivot = r
				break
			}
		}
		if pivot == -1 {
			continue
		}
		if pivot != row {
			mat[row], mat[pivot] = mat[pivot], mat[row]
		}

		var pivotVal big.Rat
		pivotVal.Set(&mat[row][col])
		for c := col; c < k+1; c++ {
			mat[row][c].Quo(&mat[row][c], &pivotVal)
		}

		for r := 0; r < m; r++ {
			if r == row {
				continue
			}
			if mat[r][col].Sign() == 0 {
				continue
			}
			var factor big.Rat
			factor.Set(&mat[r][col])
			for c := col; c < k+1; c++ {
				var prod big.Rat
				prod.Mul(&factor, &mat[row][c])
				mat[r][c].Sub(&mat[r][c], &prod)
			}
		}

		pivots = append(pivots, col)
		row++
	}

	// Check consistency: 0 = non-zero.
	for r := 0; r < m; r++ {
		allZero := true
		for c := 0; c < k; c++ {
			if mat[r][c].Sign() != 0 {
				allZero = false
				break
			}
		}
		if allZero && mat[r][k].Sign() != 0 {
			return nil, nil, ErrUnsolvable
		}
	}

	return mat, pivots, nil
}

func canUseDenseBFS(targets []uint16, maxStates uint64) bool {
	product := uint64(1)
	for _, t := range targets {
		base := uint64(t) + 1
		if base == 0 {
			return false
		}
		if product > maxStates/base {
			return false
		}
		product *= base
	}
	return product <= maxStates
}

func buildRadix(targets []uint16) (mults []uint64, bases []uint16, stateCount int, goalID uint64, err error) {
	mults = make([]uint64, len(targets))
	bases = make([]uint16, len(targets))

	product := uint64(1)
	for i, t := range targets {
		base := uint64(t) + 1
		if base == 0 {
			return nil, nil, 0, 0, ErrOverflow
		}
		bases[i] = uint16(base)
		mults[i] = product

		if product > math.MaxUint64/base {
			return nil, nil, 0, 0, ErrOverflow
		}
		product *= base

		goalID += uint64(t) * mults[i]
	}
	if product > math.MaxInt32 {
		return nil, nil, 0, 0, ErrProblemTooLarge
	}
	return mults, bases, int(product), goalID, nil
}

func compileButtons(buttons []button, mults []uint64, bases []uint16) ([]bfsButton, error) {
	out := make([]bfsButton, 0, len(buttons))
	for _, b := range buttons {
		var deltaID uint64
		for _, e := range b.entries {
			if e.delta == 0 {
				continue
			}
			base := bases[e.idx]
			if e.delta >= base {
				return nil, ErrInvalidMachine
			}
			deltaID += uint64(e.delta) * mults[e.idx]
		}
		if deltaID == 0 {
			continue
		}
		out = append(out, bfsButton{entries: b.entries, deltaID: deltaID})
	}
	return out, nil
}

func applyButton(state uint64, b bfsButton, mults []uint64, bases []uint16) (uint64, bool) {
	for _, e := range b.entries {
		base := uint64(bases[e.idx])
		digit := (state / mults[e.idx]) % base
		if digit+uint64(e.delta) >= base {
			return 0, false
		}
	}
	return state + b.deltaID, true
}

type stateKey [4]uint64

const (
	digitBits     = 16
	digitsPerWord = 64 / digitBits
	digitMask     = uint64((1 << digitBits) - 1)
)

func packTargets(targets []uint16) (stateKey, error) {
	var k stateKey
	if len(targets) > len(k)*digitsPerWord {
		return stateKey{}, ErrProblemTooLarge
	}
	for i, v := range targets {
		setDigit(&k, i, v)
	}
	return k, nil
}

func getDigit(k stateKey, idx int) uint16 {
	word := idx / digitsPerWord
	shift := uint((idx % digitsPerWord) * digitBits)
	return uint16((k[word] >> shift) & digitMask)
}

func setDigit(k *stateKey, idx int, val uint16) {
	word := idx / digitsPerWord
	shift := uint((idx % digitsPerWord) * digitBits)
	mask := digitMask << shift
	k[word] = (k[word] & ^mask) | (uint64(val) << shift)
}

func incState(s stateKey, b button, targets []uint16) (stateKey, bool) {
	next := s
	for _, e := range b.entries {
		cur := getDigit(next, e.idx)
		if uint32(cur)+uint32(e.delta) > uint32(targets[e.idx]) {
			return stateKey{}, false
		}
		setDigit(&next, e.idx, cur+e.delta)
	}
	return next, true
}

func decState(s stateKey, b button) (stateKey, bool) {
	next := s
	for _, e := range b.entries {
		cur := getDigit(next, e.idx)
		if cur < e.delta {
			return stateKey{}, false
		}
		setDigit(&next, e.idx, cur-e.delta)
	}
	return next, true
}

func minPressesBidirectionalSparse(targets []uint16, buttons []button) (int, error) {
	start := stateKey{}
	goal, err := packTargets(targets)
	if err != nil {
		return 0, err
	}
	if start == goal {
		return 0, nil
	}

	distF := map[stateKey]uint32{start: 0}
	distB := map[stateKey]uint32{goal: 0}
	frontF := []stateKey{start}
	frontB := []stateKey{goal}
	depthF := uint32(0)
	depthB := uint32(0)

	best := uint32(math.MaxUint32)

	// Safety valve: if we blow past this, the instance is likely too large for
	// graph search and needs a different approach.
	const maxVisited = 30_000_000

	for len(frontF) > 0 && len(frontB) > 0 {
		// Expand the smaller frontier to reduce work.
		expandForward := len(frontF) <= len(frontB)
		if expandForward {
			next := make([]stateKey, 0, len(frontF)*len(buttons))
			for _, s := range frontF {
				d := distF[s]
				if d != depthF {
					// Each frontier is kept as a single BFS layer.
					continue
				}
				if best != math.MaxUint32 && d+1 >= best {
					continue
				}
				for _, b := range buttons {
					ns, ok := incState(s, b, targets)
					if !ok {
						continue
					}
					nd := d + 1
					if od, ok := distB[ns]; ok {
						if cand := nd + od; cand < best {
							best = cand
						}
					}
					if _, seen := distF[ns]; seen {
						continue
					}
					distF[ns] = nd
					next = append(next, ns)
				}
			}
			frontF = next
			depthF++
		} else {
			next := make([]stateKey, 0, len(frontB)*len(buttons))
			for _, s := range frontB {
				d := distB[s]
				if d != depthB {
					continue
				}
				if best != math.MaxUint32 && d+1 >= best {
					continue
				}
				for _, b := range buttons {
					ns, ok := decState(s, b)
					if !ok {
						continue
					}
					nd := d + 1
					if od, ok := distF[ns]; ok {
						if cand := nd + od; cand < best {
							best = cand
						}
					}
					if _, seen := distB[ns]; seen {
						continue
					}
					distB[ns] = nd
					next = append(next, ns)
				}
			}
			frontB = next
			depthB++
		}

		if best != math.MaxUint32 && depthF+depthB >= best {
			break
		}

		if len(distF)+len(distB) > maxVisited {
			return 0, ErrProblemTooLarge
		}
	}

	if best == math.MaxUint32 {
		return 0, ErrUnsolvable
	}
	if uint64(best) > uint64(math.MaxInt) {
		return 0, ErrOverflow
	}
	return int(best), nil
}
