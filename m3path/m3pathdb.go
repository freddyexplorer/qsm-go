package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3point"
)

const (
	PointsTable = "points"
	PathContextsTable = "path_contexts"
	PathNodesTable    = "path_nodes"
)

func init() {
	m3db.AddTableDef(createPointsTableDef())
	m3db.AddTableDef(createPathContextsTableDef())
	m3db.AddTableDef(creatPathNodesTableDef())
}

const (
	FindPointIdPerCoord = 0
	SelectPointPerId    = 1
)

var pathEnv *m3db.QsmEnvironment

func GetPathEnv() *m3db.QsmEnvironment {
	if pathEnv == nil || pathEnv.GetConnection() == nil {
		pathEnv = m3db.GetDefaultEnvironment()
	}
	return pathEnv
}

func createPointsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PointsTable
	res.DdlColumns = "(id bigserial PRIMARY KEY," +
		" x integer NOT NULL, y integer NOT NULL, z integer NOT NULL," +
		" UNIQUE (x,y,z))"
	res.Insert = "(x,y,z) values ($1,$2,$3) returning id"
	res.SelectAll = "not to call select all on points"
	res.ExpectedCount = -1
	res.Queries = make([]string, 2)
	res.Queries[FindPointIdPerCoord] = fmt.Sprintf("select id from %s where x=$1 and y=$2 and z=$3", PointsTable)
	res.Queries[SelectPointPerId] = fmt.Sprintf("select x,y,z from %s where id=$1", PointsTable)

	return &res
}

func createPathContextsTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PathContextsTable
	res.DdlColumns = fmt.Sprintf("(id serial PRIMARY KEY," +
		" trio_ctx_id smallint NOT NULL REFERENCES %s (id),"+
		" trio_offset smallint NOT NULL," +
		" path_builders_id smallint NOT NULL REFERENCES %s (id))",
		m3point.GrowthContextsTable, m3point.PathBuildersTable)
	res.Insert = "(trio_ctx_id,trio_offset,path_builders_id) values ($1,$2,$3) returning id"
	res.SelectAll = fmt.Sprintf("select id, trio_ctx_id, trio_offset path_builders_id from %s", PathContextsTable)
	res.ExpectedCount = -1
	return &res
}

const (
	SelectPathNodesById = 0
	UpdatePathNode = 1
	SelectPathNodeByCtxAndDistance = 2
	SelectPathNodeByCtx = 3
)

func creatPathNodesTableDef() *m3db.TableDefinition {
	res := m3db.TableDefinition{}
	res.Name = PathNodesTable
	res.DdlColumns = fmt.Sprintf("(id serial PRIMARY KEY," +
		" path_ctx_id integer NOT NULL REFERENCES %s (id)," +
		" path_builders_id smallint NOT NULL REFERENCES %s (id)," +
		" trio_id smallint NOT NULL REFERENCES %s (id)," +
		" point_id bigint NOT NULL REFERENCES %s (id)," +
		" d integer NOT NULL DEFAULT 0," +
		" from1 integer NULL REFERENCES %s (id), from2 integer NULL REFERENCES %s (id), from3 integer NULL REFERENCES %s (id), " +
		" next1 integer NULL REFERENCES %s (id), next2 integer NULL REFERENCES %s (id), next3 integer NULL REFERENCES %s (id))",
		PathContextsTable, m3point.PathBuildersTable, m3point.TrioDetailsTable, PointsTable,
		PathNodesTable, PathNodesTable, PathNodesTable,
		PathNodesTable, PathNodesTable, PathNodesTable)
	res.Insert = "(path_ctx_id, path_builders_id, trio_id, point_id, d," +
		" from1, from2, from3," +
		" next1, next2, next3)" +
		" values ($1,$2,$3,$4,$5," +
		" $6,$7,$8," +
		" $9,$10,$11) returning id"
	res.SelectAll = "not to call select all on node path"
	res.ExpectedCount = -1
	res.Queries = make([]string, 4)
	res.Queries[SelectPathNodesById] = fmt.Sprintf("select path_ctx_id, path_builders_id, trio_id, point_id, d," +
		" from1, from2, from3," +
		" next1, next2, next3 from %s where id = $1", PathNodesTable)
	res.Queries[UpdatePathNode] = fmt.Sprintf("update %s set from1 = $2, from2 = $3, from3 = $4," +
		" next1 = $5, next2 = $6, next3 = $7 where id = $1", PathNodesTable)
	res.Queries[SelectPathNodeByCtxAndDistance] = fmt.Sprintf("select id, path_builders_id, trio_id, point_id," +
		" from1, from2, from3," +
		" next1, next2, next3 from %s where path_ctx_id = $1 and d = $2", PathNodesTable)
	res.Queries[SelectPathNodeByCtx] = fmt.Sprintf("select id, path_builders_id, trio_id, point_id, d," +
		" from1, from2, from3," +
		" next1, next2, next3 from %s where path_ctx_id = $1", PathNodesTable)
	return &res
}

func createTables() {
	env := GetPathEnv()
	_, err := env.GetOrCreateTableExec(PointsTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", PointsTable, err)
		return
	}
	_, err = env.GetOrCreateTableExec(PathContextsTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", PathContextsTable, err)
		return
	}
	_, err = env.GetOrCreateTableExec(PathNodesTable)
	if err != nil {
		Log.Fatalf("could not create table %s due to %v", PathNodesTable, err)
		return
	}
}