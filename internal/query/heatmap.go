package query

import (
	"database/sql"
	"time"
)

// HeatmapDay represents one day's commit count.
type HeatmapDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// CommitHeatmap returns per-day commit counts between from (inclusive) and to
// (exclusive). If email is non-empty, results are filtered to that author.
// Only days with commits are returned (sparse).
func CommitHeatmap(db *sql.DB, from, to time.Time, email string) ([]HeatmapDay, error) {
	var rows *sql.Rows
	var err error

	if email == "" {
		rows, err = db.Query(
			`SELECT SUBSTR(committed_at, 1, 10) AS day, COUNT(*) AS count
			 FROM commits
			 WHERE committed_at >= ? AND committed_at < ?
			 GROUP BY day ORDER BY day`,
			from, to,
		)
	} else {
		rows, err = db.Query(
			`SELECT SUBSTR(committed_at, 1, 10) AS day, COUNT(*) AS count
			 FROM commits
			 WHERE committed_at >= ? AND committed_at < ?
			   AND author_email = ?
			 GROUP BY day ORDER BY day`,
			from, to, email,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []HeatmapDay
	for rows.Next() {
		var d HeatmapDay
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}
