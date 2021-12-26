export function columnize(rowObjects) {
	// get keys
	const keys = Object.keys(rowObjects[0]);
	const columns = keys.reduce((obj, key) => {
		obj[key] = [];
		return obj;
	}, {});
	rowObjects.forEach((row) => {
		keys.forEach((k) => {
			columns[k].push(row[k]);
		});
	});
	return columns;
}

// cheapest cardinality
export function cardinality(column) {
	return [...new Set(column)].length;
}

export function stats(dataset) {
	const columns = columnize(dataset);
	return Object.entries(columns).map(([key, column]) => {
		return {
			key,
			cardinality: cardinality(column)
		};
	});
}
