package db

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type ValueChange struct {
	OldValue interface{} `json:"old"`
	NewValue interface{} `json:"new"`
}

func QueryAddHistory(tableName string, rowId int64, old map[string]interface{}, changed map[string]interface{}) {
	changes := make([]string, 0, 1)

	diffMap := map[string]ValueChange{}

	if changed != nil && old != nil {
		for k, v := range changed {
			if k == "updated_at" || k == "created_at" {
				continue
			}
			if fmt.Sprintf("%v", old[k]) != fmt.Sprintf("%v", v) {
				diffMap[k] = ValueChange{
					OldValue: old[k],
					NewValue: v,
				}
				changes = append(changes, k)
			}
		}
	}

	if len(diffMap) == 0 {
		return
	}

	rawDiffMap, err := json.Marshal(diffMap)

	if err != nil {
		log.Errorf("QueryAddHistory failed to marshall obj to json: %+v with error: %+v", diffMap, err)
	}

	_, err = Exec(`insert into
update_history (table_name, row_id, changed_values)
values($1, $2::bigint, $3::json)`, tableName, rowId, rawDiffMap)

	if err != nil {
		log.Errorf("QueryAddHistory %s->%d failed: %+v", tableName, rowId, err)
	}
}
