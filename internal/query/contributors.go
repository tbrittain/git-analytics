package query

import (
	"database/sql"
	"time"
)

// Contributor represents aggregated commit activity for a single author.
type Contributor struct {
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	Commits     int    `json:"commits"`
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
}

// Contributors returns per-author commit counts, additions, and deletions
// for commits between from (inclusive) and to (exclusive), ordered by
// commits descending. Files matching any of the excludeGlobs patterns are
// excluded from the additions/deletions totals but commits still count.
func Contributors(db *sql.DB, from, to time.Time, excludeGlobs []string) ([]Contributor, error) {
	excludeSQL, excludeArgs := buildExcludeClauses("fs.file_path", excludeGlobs)

	q := `SELECT c.author_email,
	        MAX(c.author_name) AS author_name,
	        COUNT(DISTINCT c.hash) AS commits,
	        COALESCE(SUM(fs.additions), 0) AS additions,
	        COALESCE(SUM(fs.deletions), 0) AS deletions
	 FROM commits c
	 LEFT JOIN file_stats fs ON fs.commit_hash = c.hash` + excludeSQL + `
	 WHERE c.committed_at >= ? AND c.committed_at < ?
	 GROUP BY c.author_email
	 ORDER BY commits DESC`

	args := make([]any, 0, len(excludeArgs)+2)
	args = append(args, excludeArgs...)
	args = append(args, from, to)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Contributor
	for rows.Next() {
		var c Contributor
		if err := rows.Scan(&c.AuthorEmail, &c.AuthorName, &c.Commits, &c.Additions, &c.Deletions); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
