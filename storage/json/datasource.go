package json

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/cube2222/octosql"
	"github.com/cube2222/octosql/config"
	"github.com/cube2222/octosql/execution"
	"github.com/cube2222/octosql/physical"
	"github.com/pkg/errors"
)

var availableFilters = map[physical.FieldType]map[physical.Relation]struct{}{
	physical.Primary:   make(map[physical.Relation]struct{}),
	physical.Secondary: make(map[physical.Relation]struct{}),
}

type DataSource struct {
	path  string
	alias string
}

func NewDataSourceBuilderFactory(path string) physical.DataSourceBuilderFactory {
	return physical.NewDataSourceBuilderFactory(
		func(filter physical.Formula, alias string) (execution.Node, error) {
			return &DataSource{
				path:  path,
				alias: alias,
			}, nil
		},
		nil,
		availableFilters,
	)
}

// NewDataSourceBuilderFactoryFromConfig creates a data source builder factory using the configuration.
func NewDataSourceBuilderFactoryFromConfig(dbConfig map[string]interface{}) (physical.DataSourceBuilderFactory, error) {
	path, err := config.GetString(dbConfig, "path")
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get path")
	}

	return NewDataSourceBuilderFactory(path), nil
}

func (ds *DataSource) Get(variables octosql.Variables) (execution.RecordStream, error) {
	file, err := os.Open(ds.path)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't open file")
	}
	sc := bufio.NewScanner(file)
	return &RecordStream{
		file:   file,
		sc:     sc,
		isDone: false,
		alias:  ds.alias,
	}, nil
}

type RecordStream struct {
	file   *os.File
	sc     *bufio.Scanner
	isDone bool
	alias  string
}

func (rs *RecordStream) Close() error {
	err := rs.file.Close()
	if err != nil {
		return errors.Wrap(err, "Couldn't close underlying file")
	}

	return nil
}

func (rs *RecordStream) Next() (*execution.Record, error) {
	if rs.isDone {
		return nil, execution.ErrEndOfStream
	}

	if !rs.sc.Scan() {
		rs.isDone = true
		rs.file.Close()
		return nil, execution.ErrEndOfStream
	}

	var record map[octosql.VariableName]interface{}
	err := json.Unmarshal(rs.sc.Bytes(), &record)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't unmarshal json record")
	}

	aliasedRecord := make(map[octosql.VariableName]interface{})
	for k, v := range record {
		aliasedRecord[octosql.VariableName(fmt.Sprintf("%s.%s", rs.alias, k))] = v
	}

	fields := make([]octosql.VariableName, 0)
	for k := range aliasedRecord {
		fields = append(fields, k)
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i] < fields[j]
	})

	return execution.NewRecord(fields, aliasedRecord), nil
}
