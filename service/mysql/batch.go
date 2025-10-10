package mysql

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
)

func batchInsert(db *gorm.DB, table string, originalFieldMap map[string][]any, start, end int) error {
	if len(originalFieldMap) == 0 || end == start {
		return nil
	}
	fieldMap := make(map[string][]any)
	for k, tmp := range originalFieldMap {
		fieldMap[k] = tmp[start:end]
	}
	fields := make([]string, 0)
	values := make([]any, 0)
	questions := make([]string, 0)
	for k := range fieldMap {
		fields = append(fields, k)
		questions = append(questions, "?")
	}
	placement := fmt.Sprintf("(%s)", strings.Join(questions, ","))
	placements := make([]string, 0)
	count := end - start
	for i := 0; i < count; i++ {
		placements = append(placements, placement)
		for _, f := range fields {
			values = append(values, fieldMap[f][i])
		}
	}
	query := fmt.Sprintf("INSERT IGNORE INTO %s (%s) VALUES %s;", table, strings.Join(fields, ","), strings.Join(placements, ","))
	r := db.Exec(query, values...)
	if r.Error != nil {
		return fmt.Errorf("fail to insert ignore: %v", r.Error)
	}
	return nil
}

func BatchInsert(db *gorm.DB, table string, fieldMap map[string][]any, maxBatchSize int) error {
	if len(fieldMap) == 0 {
		return nil
	}
	var count int
	for _, v := range fieldMap {
		count = len(v)
		break
	}
	var err error
	if count > maxBatchSize {
		batchCount := count/maxBatchSize + 1
		for i := 0; i < batchCount; i++ {
			start := i * maxBatchSize
			end := (i + 1) * maxBatchSize
			if i == batchCount-1 {
				end = count
			}
			if start < end {
				err = batchInsert(db, table, fieldMap, start, end)
				if err != nil {
					return err
				}
			}
		}
	} else {
		err = batchInsert(db, table, fieldMap, 0, count)
		if err != nil {
			return err
		}
	}
	return nil
}
