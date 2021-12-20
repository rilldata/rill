import duckdb from 'duckdb';
import { execSync } from 'child_process';
import fs from 'fs';
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

export function connect() {
	return new duckdb.Database('./scripts/nyc311-reduced.duckdb', { read_only: true });
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