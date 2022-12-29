package dag

// DAG is a simple implementation of a directed acyclic graph.
type DAG struct {
	NameMap map[string]*Node
}

func NewDAG() *DAG {
	return &DAG{
		NameMap: make(map[string]*Node),
	}
}

// TODO: handle cycles when we support model to model connection

type Node struct {
	Name     string
	Present  bool
	Parents  map[string]*Node
	Children map[string]*Node
}

func (d *DAG) Add(name string, dependants []string) *Node {
	n := d.getNode(name)
	n.Present = true

	dependantMap := make(map[string]bool)
	for _, dependant := range dependants {
		dependantMap[dependant] = true
	}

	for _, parent := range n.Parents {
		ok := dependantMap[parent.Name]
		if ok {
			delete(dependantMap, parent.Name)
		} else {
			d.removeChild(parent, name)
			delete(n.Parents, parent.Name)
		}
	}

	for newParent := range dependantMap {
		n.Parents[newParent] = d.addChild(newParent, n)
	}

	return n
}

func (d *DAG) Delete(name string) {
	n := d.getNode(name)
	n.Present = false
	d.deleteBranch(n)
}

func (d *DAG) GetChildren(name string) []string {
	children := make([]string, 0)
	childMap := make(map[string]bool)

	n, ok := d.NameMap[name]
	if !ok {
		return []string{}
	}

	// we need the immediate children to be loaded 1st.
	for _, child := range n.Children {
		children = append(children, child.Name)
		childMap[child.Name] = true
	}

	// then we load deeper children
	for _, child := range n.Children {
		deepChildren := d.GetChildren(child.Name)
		for _, deepChild := range deepChildren {
			if _, ok := childMap[deepChild]; !ok {
				children = append(children, deepChild)
				childMap[deepChild] = true
			}
		}
	}

	return children
}

func (d *DAG) GetParents(name string) []string {
	n := d.getNode(name)
	parents := make([]string, 0)
	for _, parent := range n.Parents {
		if parent.Present {
			parents = append(parents, parent.Name)
		}
	}
	return parents
}

func (d *DAG) Has(name string) bool {
	_, ok := d.NameMap[name]
	return ok
}

func (d *DAG) addChild(name string, child *Node) *Node {
	n := d.getNode(name)
	n.Children[child.Name] = child
	return n
}

func (d *DAG) removeChild(node *Node, childName string) {
	delete(node.Children, childName)
}

func (d *DAG) getNode(name string) *Node {
	n, ok := d.NameMap[name]
	if !ok {
		n = &Node{
			Name:     name,
			Parents:  make(map[string]*Node),
			Children: make(map[string]*Node),
		}
		d.NameMap[name] = n
	}
	return n
}

func (d *DAG) deleteBranch(n *Node) {
	if n.Present || len(n.Children) > 0 {
		return
	}

	for _, parent := range n.Parents {
		d.removeChild(parent, n.Name)
		d.deleteBranch(parent)
	}

	delete(d.NameMap, n.Name)
}
