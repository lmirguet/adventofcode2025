# Day 10 — Algorithm (Joltage Mode)

Day 10 (joltage mode) is: for each line/machine, find the minimum number of button presses needed to reach exact counter targets, then sum across machines.

## 1) Model: “buttons as columns of a matrix”
For one machine:

- There are `m` counters (from `{t0,t1,...}`), each starts at 0.
- There are `k` buttons (each `(…)` lists which counters it increments by +1 per press).
- Let `x_j ≥ 0` be the integer number of times you press button `j`.

Each press adds a fixed increment vector, so the final counters are:

- `A * x = t`

Where:

- `A` is an `m×k` integer matrix, `A[i][j]` is how much button `j` increments counter `i` per press (usually 0/1).
- `x` is the vector of unknown press counts.
- `t` is the target vector.

Objective:

- minimize `sum(x_j)`.

So this is an integer optimization problem (a small ILP).

In code this is built in `src/day10/compute.go` (`buildMatrix`).

We also compute per-button upper bounds:

- If button `j` touches counter `i` by `A[i][j] > 0`, then `x_j ≤ floor(t_i / A[i][j])`.
- Take the minimum across counters it touches.

This is `buttonUpperBounds`.

## 2) Why BFS explodes
A “state” is the current vector of counter values.

If you do a plain BFS over all states from 0 to `t`, the number of possible states is roughly:

- `∏(t_i + 1)`

With targets like `{154,26,19,142,16,27,114}`, this is enormous, so array-based BFS is impossible.

## 3) The solver pipeline used now
For each machine (`minPresses` in `src/day10/compute.go`), it tries, in order:

### A) Fast exact solve for square invertible cases: “pivot matrix + small enumeration”
This is `minPressesByLinearSolve`.

It tries to pick `m` “pivot” buttons so their `m×m` submatrix `B` is invertible (checked modulo primes first), and treats the remaining `k-m` buttons as “free variables”.

If `k-m` is small (≤ 6), it enumerates all combinations of those free button press counts within their upper bounds. For each choice:

- Compute `rhs = t - A_free * x_free`
- Solve `B * x_pivot = rhs` exactly using rational arithmetic (`big.Rat` via `invertMatrixInt64`)
- Accept only if `x_pivot` is all integers and non-negative
- Minimize total presses `sum(x_free) + sum(x_pivot)`

This works great when:

- the system has full row rank,
- and there exists an invertible square submatrix.

But it can fail when `B` can’t be made invertible (singular structure), even if the overall problem is solvable.

### B) Rank-deficient-safe solve: RREF + enumerate the nullspace degrees
This is `minPressesByRREFEnumeration` + `rrefAugmented`.

Steps:

1) Build the augmented matrix `[A | t]` as rationals.

2) Compute Reduced Row Echelon Form (RREF) via Gauss-Jordan elimination:

- Choose pivot columns where a non-zero entry exists.
- Normalize pivot row so pivot becomes 1.
- Eliminate that column in all other rows.

This yields a system where each pivot variable is expressed as:

- `x_pivot = rhs - Σ (coeff_free * x_free)`

3) Detect inconsistency:

- If a row becomes all zeros in `A` but rhs is non-zero, system is unsatisfiable.

4) Identify free variables:

- Columns without pivots are free variables.

5) If #free variables is small (≤ 6) and the product of their (upperBound+1) isn’t too large:

- Enumerate all assignments to free variables in range `[0..ub]`.
- For each assignment, compute pivot variables from the RREF equations.
- Require:
  - pivot values are integers (`Rat.IsInt()`),
  - pivot values ≥ 0,
  - pivot values ≤ their computed upper bounds.
- Compute total presses = sum(all `x`).
- Keep minimum.

Why this fixes singular cases:

- Even when `k == m`, the matrix can be singular (rank < m). In that case, a direct inversion-based approach fails, but RREF can still solve the system by exposing the remaining degrees of freedom.

### C) Dense BFS (only when the state space fits)
If `∏(t_i+1)` ≤ 30,000,000 (configurable), it uses the mixed-radix array BFS (`minPressesBFS`).

This is fastest for small targets because it uses a flat array for distances.

### D) Sparse bidirectional BFS (fallback)
If dense BFS doesn’t fit, it uses `minPressesBidirectionalSparse`:

- pack the counter vector into a fixed-size key (`stateKey`) with 16-bit digits,
- expand from both start (all zero) and goal (targets), meeting in the middle,
- store visited states in hash maps.

This avoids allocating an array of size `∏(t_i+1)`, but it can still blow up if the reachable space is huge, so it has a visited cap.

## 4) Parallelization
At file level, `Compute` now:

- reads all non-empty lines,
- runs `minPressesForLine` in a worker pool (`min(runtime.NumCPU()/2, 6)`),
- sums results, returning the earliest line error if any.

This parallelizes across machines, which is the natural unit of work.

