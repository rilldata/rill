package duckdbsql

func (a *AST) correctSampleClause(node astNode) astNode {
	if node == nil {
		return nil
	}
	sampleSize := toNode(node, astKeySampleSize)
	if sampleSize == nil {
		return node
	}
	newSampleSize, err := createStandaloneValue(sampleSize[astKeyValue])
	if err != nil {
		return nil
	}
	node[astKeySampleSize] = newSampleSize
	return node
}
