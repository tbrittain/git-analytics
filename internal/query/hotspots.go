package query

import (
	"database/sql"
	"time"
)

// FileHotspot represents aggregated churn for a single file path.
type FileHotspot struct {
	Path         string `json:"path"`
	LinesChanged int    `json:"lines_changed"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	Commits      int    `json:"commits"`
}

// FileHotspots returns per-file churn (additions + deletions) and commit counts
// for commits between from (inclusive) and to (exclusive), ordered by
// lines_changed descending.
func FileHotspots(db *sql.DB, from, to time.Time) ([]FileHotspot, error) {
	rows, err := db.Query(
		`SELECT fs.file_path,
		        SUM(fs.additions + fs.deletions) AS lines_changed,
		        SUM(fs.additions) AS additions,
		        SUM(fs.deletions) AS deletions,
		        COUNT(DISTINCT fs.commit_hash) AS commits
		 FROM file_stats fs
		 JOIN commits c ON c.hash = fs.commit_hash
		 WHERE c.committed_at >= ? AND c.committed_at < ?
		 GROUP BY fs.file_path
		 ORDER BY lines_changed DESC`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []FileHotspot
	for rows.Next() {
		var h FileHotspot
		if err := rows.Scan(&h.Path, &h.LinesChanged, &h.Additions, &h.Deletions, &h.Commits); err != nil {
			return nil, err
		}
		result = append(result, h)
	}
	return result, rows.Err()
}
