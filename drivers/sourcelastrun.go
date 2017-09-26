package drivers

// SourceLastRunService assists scrapers by storing when scraping was last successfully
// run on a given source.

import (
	"fmt"
	"github.com/coverprice/contentscraper/database"
	"time"
)

// Record to store when a scraping was last done on a source.
type SourceLastRun struct {
	SourceConfigId SourceConfigId
	// 0 means not run at all.
	DateLastRun uint64
}

type SourceLastRunService struct {
	dbconn *database.DbConn
}

func NewSourceLastRunService(
	dbconn *database.DbConn,
) (sourceLastRunService *SourceLastRunService, err error) {
	sourceLastRunService = &SourceLastRunService{
		dbconn: dbconn,
	}
	if err = sourceLastRunService.initTables(); err != nil {
		return nil, err
	}
	return
}

func (this *SourceLastRunService) initTables() (err error) {
	err = this.dbconn.ExecSql(`
        CREATE TABLE IF NOT EXISTS source_last_run
            ( id TEXT PRIMARY KEY
            , last_run INTEGER NOT NULL
        )
        ;
        CREATE TABLE IF NOT EXISTS post
            ( id TEXT PRIMARY KEY
            , time_created INTEGER NOT NULL
            , score INTEGER NOT NULL
            , is_published INTEGER NOT NULL
        )
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

	var row *database.Row
	row, err = this.dbconn.GetFirstRow(`
        SELECT last_run
        FROM source_last_run
        WHERE id = $a`,
		id,
	)
	if err != nil {
		return nil, err
	}
	if row != nil {
		var ok bool
		if lastRun.DateLastRun, ok = (*row)["last_run"].(uint64); !ok {
			return nil, fmt.Errorf("Could not interpret last_run column as uint64")
		}
	}
	if lastRun.DateLastRun == 0 {
		// No row (or the value was 0), so fill in a default value
		// TODO: make the default value configurable at runtime.
		lastRun.DateLastRun = uint64(time.Now().Unix()) - 7*24*60*60
	}
	return lastRun, nil
}

func (this *SourceLastRunService) UpsertLastRun(
	lastRun *SourceLastRun,
) (err error) {
	err = this.dbconn.ExecSql(`
        INSERT OR REPLACE INTO source_last_run
            ( id
            , last_run
        ) VALUES
            ( $a
            , $b
        )`,
		lastRun.SourceConfigId,
		lastRun.DateLastRun,
	)
	return
}
