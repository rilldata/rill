package catalog

import "github.com/rilldata/rill/runtime/api"

// DAG is a simple implementation of a directed acyclic graph.
type DAG struct {
	NameMap map[string]*DAGNode
}

type DAGNode struct {
	Catalog  *api.CatalogObject
	Parents  []*DAGNode
	Children []*DAGNode
}

func (d *DAG) Add(catalog *api.CatalogObject) error {
	return nil
}

func (d *DAG) Update(catalog *api.CatalogObject) error {
	return nil
}

func (d *DAG) Delete(catalog *api.CatalogObject) error {
	return nil
}
