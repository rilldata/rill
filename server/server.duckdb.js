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

app.post('/results', async (req, res) => {
	const query = req.body.query;
	let output = { status: 'QUERY_RUNNING' };
	const isValid = await validQuery(db, query);
	if (!(isValid === true)) {
		output.status = 'QUERY_INVALID';
		if (isValid.message !== 'No statement to prepare!') {
			console.log(isValid.message);
			output.error = isValid.message;
		}
		res.json(JSON.stringify(output));
		return;
	}
	try {
		let newQuery = `${query.replace(';', '')} LIMIT 50;`;
		try {
			output.results = await dbAll(db, newQuery);
		} catch (err) {
			console.log('error');
			console.log(err);
			output.results = [];
		}

		const file = JSON.parse(fs.readFileSync('./server/output.json').toString());
		output.queryInfo = await getInputTables(db, file);
		output.query = query;
	} catch (err) {
		console.error('hmm', err);
	}
	res.json(JSON.stringify(output));
});

app.listen(8081);
