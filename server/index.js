// create server function
import express from 'express';
import cors from 'cors';

import { connect, justRunIt } from './sqlite3.mjs';
import { getInputTableInformation, validQuery, cheapFirstN } from './sqlite3-info.mjs';

const app = express();
app.use(express.json());
app.use(cors());

const db = connect();

app.options('/query', cors());

app.post('/query', cors(), (req, res) => {
	const query = req.body.query;
	try {
		const results = justRunIt(db, query);
		const queryInfo = getInputTableInformation(db, query); //console.log('ok!', queryInfo); //justRunIt(`EXPLAIN QUERY PLAN ${query}`);
		res.json(JSON.stringify({ results, queryInfo }));
	} catch (err) {
		console.log(err);
	}
});

app.post('/table-definitions', (req, res) => {
	const query = req.body.query;
	let output = { status: 'QUERY_RUNNING' };
	if (!validQuery(db, query)) {
		output.status = 'QUERY_INVALID';
		res.json(JSON.stringify(output));
		return;
	}
	try {
		output.queryInfo = getInputTableInformation(db, query);
		output.query = query;
	} catch (err) {
		console.log('--- ', err);
	}
	try {
		output.results = cheapFirstN(db, query);
	} catch (err) {
		console.log('hmm', err);
	}
	res.json(JSON.stringify(output));
});

app.listen(8081);
