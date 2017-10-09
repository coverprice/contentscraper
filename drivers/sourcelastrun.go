package drivers

// SourceLastRunService assists scrapers by storing when scraping was last successfully
// run on a given source.

import (
	"database/sql"
	"time"
)

// Record to store when a scraping was last done on a source.
type SourceLastRun struct {
	SourceConfigId SourceConfigId
	// 0 means not run at all.
	DateLastRun int64
}

type SourceLastRunService struct {
	dbconn *sql.DB
	// The # of seconds prior to now that
	// a missing "Last Run" record is presumed to be.
	DefaultLastRunInterval_s int64
}

func NewSourceLastRunService(
	dbconn *sql.DB,
) (sourceLastRunService *SourceLastRunService, err error) {
	sourceLastRunService = &SourceLastRunService{
		dbconn: dbconn,
		DefaultLastRunInterval_s: int64(7 * 24 * 60 * 60),
	}
	if err = sourceLastRunService.initTables(); err != nil {
		return nil, err
	}
	return
}

func (this *SourceLastRunService) initTables() (err error) {
	_, err = this.dbconn.Exec(`
        CREATE TABLE IF NOT EXISTS source_last_run
            ( id TEXT PRIMARY KEY
            , last_run INTEGER NOT NULL
        ) WITHOUT ROWID
    `)
	return
}

func (this *SourceLastRunService) GetSourceLastRunFromId(
	id SourceConfigId,
) (lastRun *SourceLastRun, err error) {
	lastRun = &SourceLastRun{
		SourceConfigId: id,
		DateLastRun:    0,
	}

	err = this.dbconn.QueryRow(`
        SELECT last_run
        FROM source_last_run
        WHERE id = $a`,
		string(id),
	).Scan(&lastRun.DateLastRun)
	if err == sql.ErrNoRows {
		lastRun.DateLastRun = int64(time.Now().Unix()) - this.DefaultLastRunInterval_s
		err = nil
	}
	return
}

func (this *SourceLastRunService) UpsertLastRun(
	lastRun *SourceLastRun,
) (err error) {
	_, err = this.dbconn.Exec(`
        INSERT OR REPLACE INTO source_last_run
            ( id
            , last_run
        ) VALUES
            ( $a
            , $b
        )`,
		string(lastRun.SourceConfigId),
		int64(lastRun.DateLastRun),
	)
	return
}
