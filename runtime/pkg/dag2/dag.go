package dag2

import "fmt"

// DAG implements a directed acyclic graph.
// Its implementation tracks unresolved references and resolves them if possible when new vertices are added.
// The implementation is not thread safe and panics if used incorrectly.
// It's based on the implementation in runtime/pkg/dag and should replace it when that package is no longer used.
type DAG[K comparable, V any] struct {
	hash     func(V) K
	vertices map[K]*vertex[K, V]
}

// New initializes a DAG.
func New[K comparable, V any](hash func(V) K) DAG[K, V] {
	return DAG[K, V]{
		hash:     hash,
		vertices: make(map[K]*vertex[K, V]),
	}
}

type vertex[K comparable, V any] struct {
	val      V
	present  bool // True if added directly, false if only referenced by other vertices.
	parents  map[K]*vertex[K, V]
	children map[K]*vertex[K, V]
}

// Add adds a vertex to the DAG.
// It returns false if adding the vertex would create a cycle.
// It panics if the vertex is already present.
func (d DAG[K, V]) Add(val V, parentVals ...V) bool {
	k := d.hash(val)
	v, ok := d.vertices[k]
	if !ok {
		v = &vertex[K, V]{val: val}
	}

	if v.present {
		panic(fmt.Errorf("dag: vertex is already present: %v", val))
	}

	// If no parents, no need to link or check for cyclic refs.
	if len(parentVals) == 0 {
		v.present = true
		return true
	}

	// Build parents map partially (not linked yet).
	// In the optimistic case, this avoids allocating a redundant parent map for the cycles check.
	parents := make(map[K]*vertex[K, V], len(parentVals))
	for _, pv := range parentVals {
		pk := d.hash(pv)
		p, ok := d.vertices[pk]
		if !ok {
			p = &vertex[K, V]{val: pv}
		}
		// Note: not adding to d.vertices until after checks.
		parents[pk] = p
	}

	// Check for cycles (there may be existing phantom references to it).
	if len(v.children) > 0 {
		visited := make(map[K]bool, len(v.children))
		found := d.visit(v, visited, func(ck K, c V) bool {
			_, ok := parents[ck]
			return ok
		})
		if found {
			return false
		}
	}

	// Link everything
	d.vertices[k] = v
	v.present = true
	v.parents = parents
	for pk, p := range parents {
		d.vertices[pk] = p // If it's already there, it's harmless.
		if p.children == nil {
			p.children = make(map[K]*vertex[K, V])
		}
		p.children[k] = v
	}

	return true
}

// Remove removes a vertex from the DAG.
// It panics if the vertex is not present.
func (d DAG[K, V]) Remove(val V) {
	k := d.hash(val)
	v, ok := d.vertices[k]
	if !ok || !v.present {
		panic(fmt.Errorf("dag: vertex not found: %v", val))
	}

	for pk, p := range v.parents {
		if len(p.children) == 1 && !p.present {
			delete(d.vertices, pk)
		} else {
			delete(p.children, k)
		}
	}

	if len(v.children) > 0 {
		v.present = false
	} else {
		delete(d.vertices, k)
	}
}

// Roots return the vertices that have no parents.
// Vertices with only phantom references are not returned.
func (d DAG[K, V]) Roots() []V {
	var roots []V
	for _, v := range d.vertices {
		if !v.present {
			continue
		}
		if len(v.parents) == 0 {
			roots = append(roots, v.val)
		}
	}
	return roots
}

// Parents returns the parents of the given value.
func (d DAG[K, V]) Parents(val V) []V {
	k := d.hash(val)
	v, ok := d.vertices[k]
	if !ok {
		panic(fmt.Errorf("dag: vertex not found: %v", val))
	}

	parents := make([]V, 0, len(v.parents))
	for _, p := range v.parents {
		if p.present {
			parents = append(parents, p.val)
		}
	}

	return parents
}

// Children returns the children of the given value.
func (d DAG[K, V]) Children(val V) []V {
	k := d.hash(val)
	v, ok := d.vertices[k]
	if !ok {
		panic(fmt.Errorf("dag: vertex not found: %v", val))
	}

	children := make([]V, 0, len(v.children))
	for _, c := range v.children {
		// Children can't be phantom references.
		children = append(children, c.val)
	}

	return children
}

// Visit recursively visits the children of the given value.
// If the visitor function returns true, the visit is stopped.
func (d DAG[K, V]) Visit(val V, fn func(K, V) bool) {
	k := d.hash(val)
	v, ok := d.vertices[k]
	if !ok {
		panic(fmt.Errorf("dag: vertex not found: %v", val))
	}

	if len(v.children) == 0 {
		return
	}

	visited := make(map[K]bool, len(v.children))
	d.visit(v, visited, fn)
}

func (d DAG[K, V]) visit(v *vertex[K, V], visited map[K]bool, fn func(K, V) bool) bool {
	for ck, c := range v.children {
		if !visited[ck] {
			stop := fn(ck, c.val)
			if stop {
				return true
			}
			visited[ck] = true
			stop = d.visit(c, visited, fn)
			if stop {
				return true
			}
		}
	}
	return false
}
