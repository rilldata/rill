// @ts-nocheck
import fs from "fs";
import duckdb from 'duckdb';
import { execSync } from 'child_process';
import glob from 'glob';


export function dbAll(db, query) {
	return new Promise((resolve, reject) => {
		try {
			db.prepare(query).all((err, res) => {
				if (err !== null) {
					reject(err);
				} else {
					resolve(res);
				}
			});
		} catch (err) {
			reject(err);
		}
	});
}

export function dbRun(query) { 
	return new Promise((resolve, reject) => {
		db.run(query, (err) => {
				if (err !== null) reject(false);
				resolve(true);
			}
		)
	})
}

export function connect() {
	//return new duckdb.Database('./scripts/nyc311-reduced.duckdb', { read_only: true });
	return new duckdb.Database(':memory:');
}

export function testConnection() {
	return duckdb(':memory:');
}

export async function validQuery(db, query) {
	return new Promise((resolve) => {
		db.prepare(query).run((err) => {
			if (err !== null) {
				resolve(err);
			} else {
				resolve(true);
			}
		});
	});
}

function recursiveGetTableScans(parent) {
	if (parent.name === 'SEQ_SCAN') {
		return [parent.extra_info.split('\n')[0]];
	}
	if (parent.children.length) {
		return [parent.children.map(recursiveGetTableScans)];
	}
	return undefined;
}

export function getTableScans(parent) {
	return [...new Set(recursiveGetTableScans(parent).flat(Infinity))];
}

export async function getTableInfo(db, table) {
	return new Promise((resolve) => {
		return dbAll(db, `PRAGMA show('${table}');`).then(resolve);
	});
}

export async function getInputTables(db, parentNode, tableSizes = {}) {
	const tables = getTableScans(parentNode);
	return Promise.all(
		tables.map(async (tableName) => {
			const info = await getTableInfo(db, tableName);
			const head = await dbAll(db, `SELECT * from ${tableName} LIMIT 1;`);
			const [cardinality] = await dbAll(db, `select count(*) as count FROM ${tableName};`);
			return {
				info,
				table: tableName,
				head,
				size: tableSizes[tableName]?.size || 0,
				cardinality: cardinality.count
			};
		})
	);
}

/** Some operations just can't happen with the node library, which is woefully outdated and incomplete.
 * Nothing a little child_process can't fix!
 * Let's make sure to download the duckdb cli via an npm command.
 */

export function runQueryWithDuckDBCLI(query, db = './scripts/nyc311-reduced-copy.duckdb') {
	execSync(`echo "${query}" | ./server/duckdb ${db}`);
}

export function dumpDBToParquet() {
	runQueryWithDuckDBCLI(`EXPORT DATABASE './tmp-source' (FORMAT PARQUET);`);
}

export function getSourceTableSizes() {
	dumpDBToParquet();
	// get all values next:
	const files = glob.sync('./tmp-source/*.parquet').reduce((acc, file) => {
		const table = file
			.split('/')
			.slice(-1)[0]
			.split('_')
			.slice(-1)[0]
			.split('.parquet')
			.slice(0)[0];
		acc[table] = {
			// yikes on this part!
			table,
			file,
			size: fs.statSync(file).size
		};
		return acc;
	}, {});
	fs.rmdirSync('./tmp-source', { recursive: true });
	return files;
}

export async function getQuerySizeInBytes(query, location = './tmp.parquet') {
	query = query.replace(';', '');
	runQueryWithDuckDBCLI(`COPY (${query}) TO '${location}' WITH (FORMAT PARQUET)`);
	const stats = fs.statSync(location);
	/** delete the temporary file */
	fs.unlinkSync(location);
	return stats.size;
}

export function hasCreateStatement(query) {
	return query.toLowerCase().startsWith('create')
		? `Query has a CREATE statement. 
	Let us handle that for you!
	Just use SELECT and we'll do the rest.
	`
		: false;
}

export function containsMultipleQueries(query) {
	return query.split().filter((character) => character == ';').length > 1
		? 'Keep it to a single query please!'
		: false;
}

export function validateQuery(query, ...validators) {
	return validators.map((validator) => validator(query)).filter((validation) => validation);
}

const db = connect();

// console.log('copying db for profiling etc');
// fs.copyFileSync('./scripts/nyc311-reduced.duckdb', './scripts/nyc311-reduced-copy.duckdb');

// console.log('calculating sources');
// const SOURCE_TABLES = getSourceTableSizes();

await dbAll(db, 'PRAGMA enable_profiling="json";');
await dbAll(db, "PRAGMA profile_output='./last-query-output.json';");


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
		console.log('"check query" error', isValid.message);
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
	console.log('running the wrap query here')
	return new Promise((resolve, reject) => {
		db.run(wrapQueryAsTemporaryView(query), (err) => {
			if (err !== null) reject(false);
			resolve(true);
		})
	})
}

export async function createPreview(query) {
    // this should only return a preview object.
	const preview = {}
    try {
		try {
			// get the preview.
			console.log('trying to wrap query');
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
	console.log('successfully created preview')
    return preview;
}

export async function createSourceProfile(query) {
	await new Promise((resolve) => db.run(query, resolve));
	const file = JSON.parse(fs.readFileSync('./last-query-output.json').toString());
	return await getInputTables(db, file, {});
}

export async function createSourceProfileFromParquet(query) {
	// capture output from parquet query.
	let re = /'.*\.parquet'/g;
	const matches = query.match(re);
	const tables = matches === null ? [] : await Promise.all(matches.map(async (match) => {
		const info = await new Promise((resolve, reject) => db.all(`select * from parquet_schema(${match});`, (err, res) => {
			if (err !== null) reject(err);
			const output = res.map((r) => {
				return {
					Type: r.type,
					Field: r.name
				}
			})
			resolve(output);
		}));
		const head = await dbAll(db, `SELECT * from ${match} LIMIT 1;`);
		const [cardinality] = await dbAll(db, `select count(*) as count FROM ${match};`);
		const output = await dbAll(db, `SELECT total_compressed_size from parquet_metadata(${match})`);
		return {
			info: info.filter(i => i.Field !== 'duckdb_schema'),
			head, 
			cardinality: cardinality.count,
			table: match,
			size: output.reduce((acc,v) => acc + v.total_compressed_size, 0)
		}
	}))
	return tables;
}

export async function getDestinationSize(path) {
	if (fs.existsSync(path)) {
		const size = await dbAll(db, `SELECT total_compressed_size from parquet_metadata('${path}')`);
		return size.reduce((acc,v) => acc + v.total_compressed_size, 0)
	}
	return undefined;
}

export async function calculateDestinationCardinality(query) {
	const [outputSize] = await dbAll(db, 'SELECT count(*) AS cardinality from tmp;');
	return outputSize.cardinality;
}

export async function createDestinationProfile(query) {
	return await dbAll(db, `PRAGMA show(tmp);`);
}

export async function exportToParquet(query, output) {
	// generate export just in case.
	if (!fs.existsSync('./export')) {
		fs.mkdirSync('./export');
	}
	const exportQuery = `COPY (${query.replace(';', '')}) TO '${output}' (FORMAT 'parquet')`;
	return dbRun(exportQuery);
}