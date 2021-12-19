import fs from "fs";
import {
	validQuery,
	connect,
	dbAll,
	getInputTables,
	getSourceTableSizes,
	getQuerySizeInBytes,
    validateQuery,
    containsMultipleQueries,
    hasCreateStatement
} from './duckdb.mjs';

const db = connect();

// console.log('copying db for profiling etc');
// fs.copyFileSync('./scripts/nyc311-reduced.duckdb', './scripts/nyc311-reduced-copy.duckdb');

// console.log('calculating sources');
// const SOURCE_TABLES = getSourceTableSizes();

await dbAll(db, 'PRAGMA enable_profiling="json";');
await dbAll(db, "PRAGMA profile_output='./server/output.json';");

function wrapQueryAsTemporaryView(query) {
	return `CREATE OR REPLACE TEMPORARY VIEW tmp AS (
	${query.replace(';', '')}
);`;
}

export async function checkQuery(query) {
	const output = {};
	const isValid = await validQuery(db, query);
	if (!(isValid === true)) {
		output.status = 'QUERY_INVALID';
		if (isValid.message !== 'No statement to prepare!') {
			output.error = isValid.message;
		}
		console.log('error', output.error);
        output.status = 'ERROR';
		return output;
	}

	const validation = validateQuery(query, hasCreateStatement, containsMultipleQueries);
	if (validation.length) {
        output.error = validation[0];
        output.status = 'ERROR';
		return output;
	}
	// no error;
	return output;
}

export async function wrapQueryAsView(query) {
	try {
		db.exec(wrapQueryAsTemporaryView(query));
	} catch (err) {
		return err.message;
	}
	return true;
}

export async function createPreview(query) {
    // this should only return a preview object.
	const preview = {}
    try {
		try {
			// get the preview.
			preview.results = await dbAll(db, 'SELECT * from tmp LIMIT 25;');
		} catch (err) {
			console.log('error');
			console.log(err);
			preview.error = err.message;
			preview.results = [];
		}
	} catch (err) {
        preview.error = err.message;
		console.error('hmm', err);
	}
    return preview;
}

export async function createSourceProfile(query) {
	await new Promise((resolve) => db.run(query, resolve));
	const file = JSON.parse(fs.readFileSync('./server/output.json').toString());
	return await getInputTables(db, file, {});
}

export async function calculateDestinationCardinality(query) {
	const [outputSize] = await dbAll(db, 'SELECT count(*) AS cardinality from tmp;');
	return outputSize.cardinality;
}

export async function createDestinationProfile(query) {
	return await dbAll(db, `PRAGMA show(tmp);`);
}