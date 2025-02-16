/*
Copyright 2022 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schemadiff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vitess.io/vitess/go/vt/sqlparser"
)

func TestDiffTables(t *testing.T) {
	tt := []struct {
		name    string
		from    string
		to      string
		diff    string
		action  string
		isError bool
	}{
		{
			name: "identical",
			from: "create table t(id int primary key)",
			to:   "create table t(id int primary key)",
		},
		{
			name:   "change of columns",
			from:   "create table t(id int primary key)",
			to:     "create table t(id int primary key, i int)",
			diff:   "alter table t add column i int",
			action: "alter",
		},
		{
			name:   "create",
			to:     "create table t(id int primary key)",
			diff:   "create table t (\n\tid int primary key\n)",
			action: "create",
		},
		{
			name:   "drop",
			from:   "create table t(id int primary key)",
			diff:   "drop table t",
			action: "drop",
		},
		{
			name: "none",
		},
	}
	hints := &DiffHints{}
	for _, ts := range tt {
		t.Run(ts.name, func(t *testing.T) {
			var fromCreateTable *sqlparser.CreateTable
			if ts.from != "" {
				fromStmt, err := sqlparser.Parse(ts.from)
				assert.NoError(t, err)
				var ok bool
				fromCreateTable, ok = fromStmt.(*sqlparser.CreateTable)
				assert.True(t, ok)
			}
			var toCreateTable *sqlparser.CreateTable
			if ts.to != "" {
				toStmt, err := sqlparser.Parse(ts.to)
				assert.NoError(t, err)
				var ok bool
				toCreateTable, ok = toStmt.(*sqlparser.CreateTable)
				assert.True(t, ok)
			}
			// Testing two paths:
			// - one, just diff the "CREATE TABLE..." strings
			// - two, diff the CreateTable constructs
			// Technically, DiffCreateTablesQueries calls DiffTables,
			// but we expose both to users of this library. so we want to make sure
			// both work as expected irrespective of any relationship between them.
			dq, dqerr := DiffCreateTablesQueries(ts.from, ts.to, hints)
			d, err := DiffTables(fromCreateTable, toCreateTable, hints)
			switch {
			case ts.isError:
				assert.Error(t, err)
				assert.Error(t, dqerr)
			case ts.diff == "":
				assert.NoError(t, err)
				assert.NoError(t, dqerr)
				assert.Nil(t, d)
				assert.Nil(t, dq)
			default:
				assert.NoError(t, err)
				require.NotNil(t, d)
				require.False(t, d.IsEmpty())
				diff := d.StatementString()
				assert.Equal(t, ts.diff, diff)
				action, err := DDLActionStr(d)
				assert.NoError(t, err)
				assert.Equal(t, ts.action, action)

				// let's also check dq, and also validate that dq's statement is identical to d's
				assert.NoError(t, dqerr)
				require.NotNil(t, dq)
				require.False(t, dq.IsEmpty())
				diff = dq.StatementString()
				assert.Equal(t, ts.diff, diff)
			}
		})
	}
}

func TestDiffViews(t *testing.T) {
	tt := []struct {
		name    string
		from    string
		to      string
		diff    string
		action  string
		isError bool
	}{
		{
			name: "identical",
			from: "create view v1 as select a, b, c from t",
			to:   "create view v1 as select a, b, c from t",
		},
		{
			name:   "change of column list, qualifiers",
			from:   "create view v1 (col1, `col2`, `col3`) as select `a`, `b`, c from t",
			to:     "create view v1 (`col1`, col2, colother) as select a, b, `c` from t",
			diff:   "alter view v1(col1, col2, colother) as select a, b, c from t",
			action: "alter",
		},
		{
			name:   "create",
			to:     "create view v1 as select a, b, c from t",
			diff:   "create view v1 as select a, b, c from t",
			action: "create",
		},
		{
			name:   "drop",
			from:   "create view v1 as select a, b, c from t",
			diff:   "drop view v1",
			action: "drop",
		},
		{
			name: "none",
		},
	}
	hints := &DiffHints{}
	for _, ts := range tt {
		t.Run(ts.name, func(t *testing.T) {
			var fromCreateView *sqlparser.CreateView
			if ts.from != "" {
				fromStmt, err := sqlparser.Parse(ts.from)
				assert.NoError(t, err)
				var ok bool
				fromCreateView, ok = fromStmt.(*sqlparser.CreateView)
				assert.True(t, ok)
			}
			var toCreateView *sqlparser.CreateView
			if ts.to != "" {
				toStmt, err := sqlparser.Parse(ts.to)
				assert.NoError(t, err)
				var ok bool
				toCreateView, ok = toStmt.(*sqlparser.CreateView)
				assert.True(t, ok)
			}
			// Testing two paths:
			// - one, just diff the "CREATE TABLE..." strings
			// - two, diff the CreateTable constructs
			// Technically, DiffCreateTablesQueries calls DiffTables,
			// but we expose both to users of this library. so we want to make sure
			// both work as expected irrespective of any relationship between them.
			dq, dqerr := DiffCreateViewsQueries(ts.from, ts.to, hints)
			d, err := DiffViews(fromCreateView, toCreateView, hints)
			switch {
			case ts.isError:
				assert.Error(t, err)
				assert.Error(t, dqerr)
			case ts.diff == "":
				assert.NoError(t, err)
				assert.NoError(t, dqerr)
				assert.Nil(t, d)
				assert.Nil(t, dq)
			default:
				assert.NoError(t, err)
				require.NotNil(t, d)
				require.False(t, d.IsEmpty())
				diff := d.StatementString()
				assert.Equal(t, ts.diff, diff)
				action, err := DDLActionStr(d)
				assert.NoError(t, err)
				assert.Equal(t, ts.action, action)

				// let's also check dq, and also validate that dq's statement is identical to d's
				assert.NoError(t, dqerr)
				require.NotNil(t, dq)
				require.False(t, dq.IsEmpty())
				diff = dq.StatementString()
				assert.Equal(t, ts.diff, diff)
			}
		})
	}
}

func TestDiffSchemas(t *testing.T) {
	tt := []struct {
		name        string
		from        string
		to          string
		diffs       []string
		expectError string
	}{
		{
			name: "identical tables",
			from: "create table t(id int primary key)",
			to:   "create table t(id int primary key)",
		},
		{
			name: "change of table columns",
			from: "create table t(id int primary key)",
			to:   "create table t(id int primary key, i int)",
			diffs: []string{
				"alter table t add column i int",
			},
		},
		{
			name: "create table",
			to:   "create table t(id int primary key)",
			diffs: []string{
				"create table t (\n\tid int primary key\n)",
			},
		},
		{
			name: "create table (2)",
			from: ";;; ; ;    ;;;",
			to:   "create table t(id int primary key)",
			diffs: []string{
				"create table t (\n\tid int primary key\n)",
			},
		},
		{
			name: "drop table",
			from: "create table t(id int primary key)",
			diffs: []string{
				"drop table t",
			},
		},
		{
			name: "create, alter, drop tables",
			from: "create table t1(id int primary key); create table t2(id int primary key); create table t3(id int primary key)",
			to:   "create table t4(id int primary key); create table t2(id bigint primary key); create table t3(id int primary key)",
			diffs: []string{
				"drop table t1",
				"alter table t2 modify column id bigint primary key",
				"create table t4 (\n\tid int primary key\n)",
			},
		},
		{
			name: "identical views",
			from: "create table t(id int); create view v1 as select * from t",
			to:   "create table t(id int); create view v1 as select * from t",
		},
		{
			name: "modified view",
			from: "create table t(id int); create view v1 as select * from t",
			to:   "create table t(id int); create view v1 as select id from t",
			diffs: []string{
				"alter view v1 as select id from t",
			},
		},
		{
			name: "drop view",
			from: "create table t(id int); create view v1 as select * from t",
			to:   "create table t(id int);",
			diffs: []string{
				"drop view v1",
			},
		},
		{
			name: "create view",
			from: "create table t(id int)",
			to:   "create table t(id int); create view v1 as select id from t",
			diffs: []string{
				"create view v1 as select id from t",
			},
		},
		{
			name:        "create view: unresolved dependencies",
			from:        "create table t(id int)",
			to:          "create table t(id int); create view v1 as select id from t2",
			expectError: ErrViewDependencyUnresolved.Error(),
		},
		{
			name: "convert table to view",
			from: "create table t(id int); create table v1 (id int)",
			to:   "create table t(id int); create view v1 as select * from t",
			diffs: []string{
				"drop table v1",
				"create view v1 as select * from t",
			},
		},
		{
			name: "convert view to table",
			from: "create table t(id int); create view v1 as select * from t",
			to:   "create table t(id int); create table v1 (id int)",
			diffs: []string{
				"drop view v1",
				"create table v1 (\n\tid int\n)",
			},
		},
		{
			name:        "unsupported statement",
			from:        "create table t(id int)",
			to:          "drop table t",
			expectError: ErrUnsupportedStatement.Error(),
		},
		{
			name: "create, alter, drop tables and views",
			from: "create view v1 as select * from t1; create table t1(id int primary key); create table t2(id int primary key); create view v2 as select * from t2; create table t3(id int primary key);",
			to:   "create view v0 as select * from v2, t2; create table t4(id int primary key); create view v2 as select id from t2; create table t2(id bigint primary key); create table t3(id int primary key)",
			diffs: []string{
				"drop table t1",
				"drop view v1",
				"alter table t2 modify column id bigint primary key",
				"create table t4 (\n\tid int primary key\n)",
				"alter view v2 as select id from t2",
				"create view v0 as select * from v2, t2",
			},
		},
	}
	hints := &DiffHints{}
	for _, ts := range tt {
		t.Run(ts.name, func(t *testing.T) {
			diffs, err := DiffSchemasSQL(ts.from, ts.to, hints)
			if ts.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), ts.expectError)
			} else {
				assert.NoError(t, err)
				statements := []string{}
				for _, d := range diffs {
					statement := sqlparser.String(d.Statement())
					statements = append(statements, statement)
				}
				if ts.diffs == nil {
					ts.diffs = []string{}
				}
				assert.Equal(t, ts.diffs, statements)
			}
		})
	}
}
