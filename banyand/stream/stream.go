// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Package stream implements a time-series-based storage which is consists of a sequence of element.
// Each element drops in a arbitrary interval. They are immutable, can not be updated or overwritten.
package stream

import (
	"context"

	databasev1 "github.com/apache/skywalking-banyandb/api/proto/banyandb/database/v1"
	"github.com/apache/skywalking-banyandb/banyand/tsdb"
	"github.com/apache/skywalking-banyandb/banyand/tsdb/index"
	"github.com/apache/skywalking-banyandb/pkg/logger"
)

// a chunk is 1MB.
const chunkSize = 1 << 20

type stream struct {
	db          tsdb.Supplier
	l           *logger.Logger
	schema      *databasev1.Stream
	indexWriter *index.Writer
	name        string
	group       string
	indexRules  []*databasev1.IndexRule
	shardNum    uint32
}

func (s *stream) GetSchema() *databasev1.Stream {
	return s.schema
}

func (s *stream) GetIndexRules() []*databasev1.IndexRule {
	return s.indexRules
}

func (s *stream) Close() error {
	return nil
}

func (s *stream) parseSpec() {
	s.name, s.group = s.schema.GetMetadata().GetName(), s.schema.GetMetadata().GetGroup()
}

type streamSpec struct {
	schema     *databasev1.Stream
	indexRules []*databasev1.IndexRule
}

func openStream(shardNum uint32, db tsdb.Supplier, spec streamSpec, l *logger.Logger) *stream {
	sm := &stream{
		shardNum:   shardNum,
		schema:     spec.schema,
		indexRules: spec.indexRules,
		l:          l,
	}
	sm.parseSpec()
	ctx := context.WithValue(context.Background(), logger.ContextKey, l)

	if db == nil {
		return sm
	}
	sm.db = db
	sm.indexWriter = index.NewWriter(ctx, index.WriterOptions{
		DB:                db,
		ShardNum:          shardNum,
		Families:          spec.schema.TagFamilies,
		IndexRules:        spec.indexRules,
		EnableGlobalIndex: true,
	})
	return sm
}
