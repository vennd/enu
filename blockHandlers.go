package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	"github.com/vennd/enu/log"
)

func GetBlocks(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {

	var blocks enulib.Blocks
	requestId := c.Value(consts.RequestIdKey).(string)
	blocks.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	//	 Query DB
	database.Init()
	stmt, err := database.Db.Prepare("select * from blocks order by blockId desc limit 10")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s\n", err.Error())
		panic(err.Error())
	}
	defer stmt.Close()

	//	 Get rows
	var rows *sql.Rows
	rows, err = stmt.Query()
	if err != nil {
		panic(err.Error())
	}

	// Iterate through last 10 blocks and return

	i := 1
	for rows.Next() {
		var rowId int64
		var blockId int64
		var status string
		var duration int64

		if err := rows.Scan(&rowId, &blockId, &status, &duration); err != nil {
			log.FluentfContext(consts.LOGERROR, c, err.Error())
		}

		block := enulib.Block{BlockId: blockId, Status: status, Duration: duration}

		log.FluentfContext(consts.LOGINFO, c, "Blockid: %d, Status: %s, Duration: %d\n", block.BlockId, block.Status, block.Duration)

		blocks.Allblocks = append(blocks.Allblocks, block)

		// Maximum of 10 rows
		i++
		if i == 10 {
			break
		}
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(blocks); err != nil {
		panic(err)
	}
	return nil
}
