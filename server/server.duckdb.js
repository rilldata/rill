// create server function
import express from 'express';
import cors from 'cors';
import fs from 'fs';

import { validQuery, connect, dbAll, getInputTables } from './duckdb.mjs';

const app = express();
app.use(express.json());
app.use(cors());

const db = connect();

await dbAll(db, 'PRAGMA enable_profiling="json";');
await dbAll(db, "PRAGMA profile_output='./server/output.json';");

app.options('/query', cors());

/**
 * Wraps a query as a TEMPORARY VIEW, which
 * we can then treat as a table in subsequent queries.
 * @param {string} query
 * @returns {string}
 */
function wrapQueryAsTemporaryView(query) {
	return `CREATE OR REPLACE TEMPORARY VIEW tmp AS (
	${query.replace(';', '')}
);`;
}

// intercept CREATE queries and return an error.

function hasCreateStatement(query) {
	return query.toLowerCase().startsWith('create')
		? `Query has a CREATE statement. 
	Let us handle that for you!
	Just use SELECT and we'll do the rest.
	`
		: false;
}

function containsMultipleQueries(query) {
	return query.split().filter((character) => character == ';').length > 1
		? 'Keep it to a single query please!'
		: false;
}

function validateQuery(query, ...validators) {
	return validators.map((validator) => validator(query)).filter((validation) => validation);
}

app.post('/results', async (req, res) => {
	const query = req.body.query;
	let output = { status: 'QUERY_RUNNING' };
	const isValid = await validQuery(db, query);
	if (!(isValid === true)) {
		output.status = 'QUERY_INVALID';
		if (isValid.message !== 'No statement to prepare!') {
			output.error = isValid.message;
		}
		res.json(JSON.stringify(output));
		return;
	}

	const validation = validateQuery(query, hasCreateStatement, containsMultipleQueries);
	if (validation.length) {
		res.json(JSON.stringify({ ...output, error: validation[0] }));
	}
	try {
		try {
			db.exec(wrapQueryAsTemporaryView(query));
			// get the preview.
			output.results = await dbAll(db, 'SELECT * from tmp LIMIT 25;');
		} catch (err) {
			console.log('error');
			console.log(err);
			output.results = [];
		}
		// exec the statement to get the profiling information.
		await new Promise((resolve) => db.run(query, resolve));
		const file = JSON.parse(fs.readFileSync('./server/output.json').toString());
		// get the profile.
		output.queryInfo = await getInputTables(db, file);
		// TEST: let's try getting the destination profile.
		// can I calculate rollup factors?
		output.destinationInfo = {};
		output.destinationInfo.info = await dbAll(db, `PRAGMA show(tmp);`);

		output.costSummary = {};
		// destination table size
		const [outputSize] = await dbAll(db, 'SELECT count(*) AS cardinality from tmp;');
		output.destinationInfo.size = outputSize.cardinality;
		output.query = query;
	} catch (err) {
		console.error('hmm', err);
	}
	res.json(JSON.stringify(output));
});

app.listen(8081);
