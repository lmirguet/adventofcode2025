package day4

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrNoGrid          = errors.New("no grid provided")
	ErrNonRectangular  = errors.New("grid is not rectangular")
	ErrInvalidCellRune = errors.New("invalid cell character")
)

type Result struct {
	TotalRemoved int
}

// Compute odczytuje siatkę znaków '.' i '@' i wielokrotnie usuwa 'kulki' ('@'),
// które mają mniej niż 4 sąsiadów '@' w 8 sąsiednich pozycjach, aż nie będzie
// można usunąć więcej kulek. Zwraca łączną liczbę usuniętych kulek.
func Compute(r io.Reader) (Result, error) {
	// Odczyt wiersz po wierszu wejścia, z pominięciem pustych linii.
	// Używamy bufio.Scanner z powiększonym buforem, aby obsłużyć duże
	// siatki, jeśli to konieczne.
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	// `grid` zawiera siatkę jako slice wierszy ([]byte)
	var grid [][]byte
	// `width` przechowuje oczekiwaną szerokość dla każdej niepustej linii
	width := -1
	// licznik wierszy do czytelnych komunikatów o błędach
	line := 0

	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		if width == -1 {
			width = len(raw)
		}
		if len(raw) != width {
			return Result{}, fmt.Errorf("line %d: %w", line, ErrNonRectangular)
		}

		// Konwertuje linię na []byte i sprawdza poprawność znaków
		row := []byte(raw)
		for col, ch := range row {
			if ch != '.' && ch != '@' {
				return Result{}, fmt.Errorf("line %d col %d: %w: %q", line, col+1, ErrInvalidCellRune, ch)
			}
		}
		grid = append(grid, row)
	}

	if err := scanner.Err(); err != nil {
		return Result{}, err
	}
	if len(grid) == 0 {
		return Result{}, ErrNoGrid
	}
	if width == 0 {
		return Result{}, ErrNoGrid
	}

	// wysokość i szerokość siatki
	h := len(grid)
	w := width

	// `present` oznacza, czy komórka nadal zawiera '@'
	present := make([]bool, h*w)
	// `degree` przechowuje aktualną liczbę sąsiadów '@' dla każdej komórki
	degree := make([]int, h*w)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if grid[y][x] == '@' {
				present[y*w+x] = true
			}
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			// Jeśli komórka nie jest obecna, pomijamy obliczenia
			if !present[idx] {
				continue
			}
			// Zlicza sąsiadów '@' wokół komórki (8 kierunków)
			neighbors := 0
			for dy := -1; dy <= 1; dy++ {
				ny := y + dy
				if ny < 0 || ny >= h {
					continue
				}
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					nx := x + dx
					if nx < 0 || nx >= w {
						continue
					}
					if present[ny*w+nx] {
						neighbors++
					}
				}
			}
			degree[idx] = neighbors
		}
	}

	// Buduje początkową kolejkę komórek, które mają mniej niż 4 sąsiadów i
	// dlatego powinny zostać usunięte w iteracji.
	queue := make([]int, 0, h*w)
	for idx, ok := range present {
		if ok && degree[idx] < 4 {
			queue = append(queue, idx)
		}
	}

	// Przetwarzamy kolejkę, usuwając kwalifikujące się komórki i aktualizując
	// stopnie ich sąsiadów — jeśli sąsiad spadnie poniżej 4, zostaje dodany
	// do kolejki. To przypomina algorytm obliczania k-core (tutaj k = 4)
	// dla grafu.
	removed := 0
	for len(queue) > 0 {
		idx := queue[0]
		queue = queue[1:]
		// Sprawdza, czy komórka nadal jest obecna i wymaga usunięcia
		if !present[idx] || degree[idx] >= 4 {
			continue
		}
		// Usuwa komórkę i zwiększa licznik
		present[idx] = false
		removed++
		// Aktualizuje sąsiadów
		y := idx / w
		x := idx % w
		for dy := -1; dy <= 1; dy++ {
			ny := y + dy
			if ny < 0 || ny >= h {
				continue
			}
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx := x + dx
				if nx < 0 || nx >= w {
					continue
				}
				nidx := ny*w + nx
				if !present[nidx] {
					continue
				}
				degree[nidx]--
				// Si ce voisin tombe sous le seuil, il sera supprimé aussi
				if degree[nidx] < 4 {
					queue = append(queue, nidx)
				}
			}
		}
	}

	return Result{TotalRemoved: removed}, nil
}
