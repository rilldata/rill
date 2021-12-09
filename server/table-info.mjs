import { justRunIt, testConnection } from './db.mjs';

let testDB;

export function getInputTables(queryPlan) {
	let inputTables = [];
	let materializedTables = [];

	if (queryPlan)
		queryPlan.forEach((plan) => {
			if (plan.detail.startsWith('MATERIALIZE') || plan.detail.startsWith('CO-ROUTINE')) {
				let materializedTable = plan.detail
					.replace('MATERIALIZE', '')
					.replace('CO-ROUTINE', '')
					.trim()
					.split(' ')[0];
				materializedTables.push(materializedTable);
			}
			if (plan.detail.startsWith('SCAN') || plan.detail.startsWith('SEARCH')) {
				let str = plan.detail.replace('SCAN', '').replace('SEARCH', '').trim().split(' ')[0];
				inputTables.push(str);
			}
		});
	let inputSet = new Set(inputTables);
	let materializedSet = new Set(materializedTables);
	let difference = new Set([...inputSet].filter((x) => !materializedSet.has(x)));
	inputTables = [...new Set(difference)];
	materializedTables = [...materializedSet];
	return inputTables;
}

export function head(db, table, n = 5) {
	return justRunIt(db, `SELECT * from ${table} LIMIT 5;`);
}

export function validQuery(db, query) {
	try {
		db.prepare(query).run();
		return true;
	} catch (err) {
		return false;
	}
}

export function getInputTableInformation(db, query) {
	// this is probbaly expensive but whatever.
	if (!validQuery(db, query)) return undefined;

	const results = justRunIt(db, `EXPLAIN QUERY PLAN ${query}`);
	// get input tables
	const inputTables = getInputTables(results);
	// for each input table, get table structure.
	return inputTables.map((table) => ({
		table,
		info: justRunIt(db, `PRAGMA table_info(${table})`),
		head: head(db, table)
	}));
}

export function cheapFirstN(db, query, n = 20) {
	let output = [];
	let i = 0;
	for (const row of db.prepare(query).iterate()) {
		output.push(row);
		if (i > n) break;
		i += 1;
	}
	return output;
}
