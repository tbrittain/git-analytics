package query

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// CoChangePair represents two files that frequently change together.
type CoChangePair struct {
	FileA         string  `json:"file_a"`
	FileB         string  `json:"file_b"`
	CoChangeCount int     `json:"co_change_count"`
	CommitsA      int     `json:"commits_a"`
	CommitsB      int     `json:"commits_b"`
	CouplingRatio float64 `json:"coupling_ratio"`
}

// CoChanges returns file pairs that frequently appear in the same commits
// between from (inclusive) and to (exclusive), ordered by co-change count
// descending. Only pairs with at least minCount shared commits are returned.
// Files matching any of the excludeGlobs patterns are omitted.
func CoChanges(db *sql.DB, from, to time.Time, minCount int, limit int, excludeGlobs []string) ([]CoChangePair, error) {
	excludeA, excludeArgsA := buildExcludeClauses("a.file_path", excludeGlobs)
	excludeB, excludeArgsB := buildExcludeClauses("b.file_path", excludeGlobs)
	excludeFS, excludeArgsFS := buildExcludeClauses("fs.file_path", excludeGlobs)

	var b strings.Builder
	b.WriteString(`WITH pairs AS (
    SELECT a.file_path AS file_a, b.file_path AS file_b,
           COUNT(DISTINCT a.commit_hash) AS co_change_count
    FROM file_stats a
    JOIN file_stats b ON a.commit_hash = b.commit_hash
         AND a.file_path < b.file_path
    JOIN commits c ON c.hash = a.commit_hash
    WHERE c.committed_at >= ? AND c.committed_at < ?`)
	b.WriteString(excludeA)
	b.WriteString(excludeB)
	b.WriteString(fmt.Sprintf(`
    GROUP BY a.file_path, b.file_path
    HAVING co_change_count >= ?
),
file_commits AS (
    SELECT fs.file_path, COUNT(DISTINCT fs.commit_hash) AS commit_count
    FROM file_stats fs
    JOIN commits c ON c.hash = fs.commit_hash
    WHERE c.committed_at >= ? AND c.committed_at < ?%s
    GROUP BY fs.file_path
)
SELECT p.file_a, p.file_b, p.co_change_count,
       fa.commit_count, fb.commit_count
FROM pairs p
JOIN file_commits fa ON fa.file_path = p.file_a
JOIN file_commits fb ON fb.file_path = p.file_b
ORDER BY p.co_change_count DESC
LIMIT ?`, excludeFS))

	args := make([]any, 0, 6+len(excludeArgsA)+len(excludeArgsB)+len(excludeArgsFS))
	// pairs CTE args
	args = append(args, from, to)
	args = append(args, excludeArgsA...)
	args = append(args, excludeArgsB...)
	args = append(args, minCount)
	// file_commits CTE args
	args = append(args, from, to)
	args = append(args, excludeArgsFS...)
	// LIMIT
	args = append(args, limit)

	rows, err := db.Query(b.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []CoChangePair
	for rows.Next() {
		var p CoChangePair
		if err := rows.Scan(&p.FileA, &p.FileB, &p.CoChangeCount, &p.CommitsA, &p.CommitsB); err != nil {
			return nil, err
		}
		minCommits := p.CommitsA
		if p.CommitsB < minCommits {
			minCommits = p.CommitsB
		}
		if minCommits > 0 {
			p.CouplingRatio = float64(p.CoChangeCount) / float64(minCommits)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
