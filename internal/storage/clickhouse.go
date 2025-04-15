package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"errors"

	_ "github.com/ClickHouse/clickhouse-go"
)

type ClickHouseService struct {
	db *sql.DB
}

func NewClickHouseService() *ClickHouseService {
	return &ClickHouseService{}
}

func (s *ClickHouseService) Connect(host, port, database, user, jwtToken string, secure bool) error {
	protocol := "http"
	if secure {
		protocol = "https"
	}
	
	dsn := fmt.Sprintf("%s://%s:%s?username=%s&database=%s&jwt=%s",
		protocol, host, port, user, database, jwtToken)
	
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return err
	}
	
	s.db = db
	return db.Ping()
}

func (s *ClickHouseService) GetTables() ([]string, error) {
	rows, err := s.db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (s *ClickHouseService) GetTableSchema(table string) ([]string, error) {
	rows, err := s.db.Query(fmt.Sprintf("DESCRIBE %s", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var name, typ, defaultType, defaultExpr, comment, codec string
		if err := rows.Scan(&name, &typ, &defaultType, &defaultExpr, &comment, &codec); err != nil {
			return nil, err
		}
		columns = append(columns, name)
	}
	return columns, nil
}

func (s *ClickHouseService) ExportData(table string, columns []string, batchSize int) (<-chan []map[string]interface{}, <-chan error) {
	dataChan := make(chan []map[string]interface{})
	errChan := make(chan error, 1)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ","), table)
		rows, err := s.db.Query(query)
		if err != nil {
			errChan <- err
			return
		}
		defer rows.Close()

		columnNames, err := rows.Columns()
		if err != nil {
			errChan <- err
			return
		}

		var batch []map[string]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columnNames))
			valuePtrs := make([]interface{}, len(columnNames))
			for i := range columnNames {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				errChan <- err
				return
			}

			entry := make(map[string]interface{})
			for i, col := range columnNames {
				entry[col] = values[i]
			}

			batch = append(batch, entry)
			if len(batch) >= batchSize {
				dataChan <- batch
				batch = nil
			}
		}

		if len(batch) > 0 {
			dataChan <- batch
		}
	}()

	return dataChan, errChan
}

func (s *ClickHouseService) ImportData(table string, columns []string, data []map[string]interface{}) (int, error) {
	if s.db == nil {
        return 0, errors.New("not connected to ClickHouse")
    }

    if len(data) == 0 {
        return 0, nil
    }

    // Begin transaction
    tx, err := s.db.Begin()
    if err != nil {
        return 0, fmt.Errorf("failed to begin transaction: %v", err)
    }
    defer func() {
        if tx != nil {
            tx.Rollback()
        }
    }()

    // Prepare the INSERT statement
    placeholders := make([]string, len(columns))
    for i := range placeholders {
        placeholders[i] = "?"
    }

    query := fmt.Sprintf(
        "INSERT INTO %s (%s) VALUES (%s)",
        table,
        strings.Join(columns, ","),
        strings.Join(placeholders, ","),
    )

    stmt, err := tx.Prepare(query)
    if err != nil {
        return 0, fmt.Errorf("failed to prepare statement: %v", err)
    }
    defer stmt.Close()

    // Execute for each row
    insertedRows := 0
    for _, row := range data {
        // Build values slice in column order
        values := make([]interface{}, len(columns))
        for i, col := range columns {
            values[i] = row[col]
        }

        _, err := stmt.Exec(values...)
        if err != nil {
            return insertedRows, fmt.Errorf("failed to insert row: %v", err)
        }
        insertedRows++
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return insertedRows, fmt.Errorf("failed to commit transaction: %v", err)
    }

    // Set tx to nil so defer doesn't try to rollback after successful commit
    tx = nil

    return insertedRows, nil
}