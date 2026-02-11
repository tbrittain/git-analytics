package query

import (
	"database/sql"
	"sort"
	"time"
)

// FileOwnership represents per-file ownership analysis showing the dominant
// contributors by lines changed (additions + deletions).
type FileOwnership struct {
	Path              string  `json:"path"`
	TopAuthorName     string  `json:"top_author_name"`
	TopAuthorEmail    string  `json:"top_author_email"`
	TopAuthorPct      float64 `json:"top_author_pct"`
	SecondAuthorName  string  `json:"second_author_name"`
	SecondAuthorEmail string  `json:"second_author_email"`
	SecondAuthorPct   float64 `json:"second_author_pct"`
	ContributorCount  int     `json:"contributor_count"`
	TotalLines        int     `json:"total_lines"`
}

// FileOwnerships returns per-file ownership analysis for commits between from
// (inclusive) and to (exclusive). Results are sorted by top_author_pct descending
// (highest concentration of ownership first). Files matching any of the
// excludeGlobs patterns are omitted.
func FileOwnerships(db *sql.DB, from, to time.Time, excludeGlobs []string) ([]FileOwnership, error) {
	excludeSQL, excludeArgs := buildExcludeClauses("fs.file_path", excludeGlobs)

	q := `WITH file_author AS (
    SELECT fs.file_path, c.author_email, MAX(c.author_name) AS author_name,
           SUM(fs.additions + fs.deletions) AS lines_changed
    FROM file_stats fs
    JOIN commits c ON c.hash = fs.commit_hash
    WHERE c.committed_at >= ? AND c.committed_at < ?` + excludeSQL + `
    GROUP BY fs.file_path, c.author_email
)
SELECT file_path, author_email, author_name, lines_changed,
       SUM(lines_changed) OVER (PARTITION BY file_path) AS total_lines,
       COUNT(*) OVER (PARTITION BY file_path) AS contributor_count
FROM file_author
ORDER BY file_path, lines_changed DESC`

	args := make([]any, 0, len(excludeArgs)+2)
	args = append(args, from, to)
	args = append(args, excludeArgs...)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []FileOwnership
	var current *FileOwnership
	authorIdx := 0

	for rows.Next() {
		var filePath, email, name string
		var linesChanged, totalLines, contribCount int

		if err := rows.Scan(&filePath, &email, &name, &linesChanged, &totalLines, &contribCount); err != nil {
			return nil, err
		}

		pct := 0.0
		if totalLines > 0 {
			pct = float64(linesChanged) / float64(totalLines) * 100
		}

		if current == nil || current.Path != filePath {
			// New file — start a new entry.
			result = append(result, FileOwnership{
				Path:             filePath,
				TotalLines:       totalLines,
				ContributorCount: contribCount,
				TopAuthorName:    name,
				TopAuthorEmail:   email,
				TopAuthorPct:     pct,
			})
			current = &result[len(result)-1]
			authorIdx = 1
		} else if authorIdx == 1 {
			// Second author for this file.
			current.SecondAuthorName = name
			current.SecondAuthorEmail = email
			current.SecondAuthorPct = pct
			authorIdx++
		}
		// Skip authors beyond the second.
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Sort by top_author_pct descending.
	sort.Slice(result, func(i, j int) bool {
		return result[i].TopAuthorPct > result[j].TopAuthorPct
	})

	return result, nil
}
