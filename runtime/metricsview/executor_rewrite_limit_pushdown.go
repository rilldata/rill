package metricsview

// rewriteLimitsIntoSubqueries rewrites the AST pushing limits and sorts into subqueries where possible.
// It only pushes into sub-queries that make up the "spine" of the query result, i.e. where it does not impact correctness.
// This may help optimize joins in some cases.
func (e *Executor) rewriteLimitsIntoSubqueries(ast *AST) error {
	// There must be a limit and order
	if len(ast.Root.OrderBy) == 0 || ast.Root.Limit == nil {
		return nil
	}

	return e.rewriteLimitsIntoSubqueriesWalk(ast, ast.Root)
}

func (e *Executor) rewriteLimitsIntoSubqueriesWalk(ast *AST, n *SelectNode) error {
	// Skip if doesn't have subqueries
	if n.FromSelect == nil {
		return nil
	}

	// We can only push order and limits down to a node that makes up the "spine" of the query.
	// E.g. if the other subqueries are left-joined to FromSelect, we can only push order and limits down to FromSelect.
	// This code identifies the "spine" node.
	var spineNode *SelectNode
	if n.SpineSelect != nil {
		spineNode = n.SpineSelect
	} else if n.JoinComparisonSelect != nil {
		switch n.JoinComparisonType {
		case JoinTypeLeft:
			spineNode = n.FromSelect
		case JoinTypeRight:
			spineNode = n.JoinComparisonSelect
		default:
			// Can't push order and limits down to either for other join types
			return nil
		}
	} else {
		spineNode = n.FromSelect
	}

	// Apply limits
	rewriteLimitInSubquery(n, spineNode, ast.Root.OrderBy, ast.Root.Limit, ast.Root.Offset)

	// Recurse
	err := e.rewriteLimitsIntoSubqueriesWalk(ast, spineNode)
	if err != nil {
		return err
	}

	return nil
}

func rewriteLimitInSubquery(parent, child *SelectNode, order []OrderFieldNode, limit, offset *int64) {
	// Skip if the node already has order or limit
	if len(child.OrderBy) > 0 || child.Limit != nil || child.Offset != nil {
		return
	}

	// Skip if any of the order fields are not present in the node
	for _, f := range order {
		if !hasName(child, f.Name) {
			return
		}
	}

	// Apply changes to child, and clear offset from parent
	child.OrderBy = order
	child.Limit = limit
	child.Offset = offset
	parent.Offset = nil
}
