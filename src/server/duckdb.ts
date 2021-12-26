// @ts-nocheck
import fs from "fs";
import duckdb from 'duckdb';

interface DB {
	all: Function;
	exec: Function;
	run: Function
}

export function connect() : DB {
	return new duckdb.Database(':memory:');
}

const db:DB = connect();

export function dbAll(db:DB, query:string) {
	return new Promise((resolve, reject) => {
		try {
			db.all(query, (err, res) => {
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

export function dbRun(query:string) { 
	return new Promise((resolve, reject) => {
		db.run(query, (err) => {
				if (err !== null) reject(false);
				resolve(true);
			}
		)
	})
}

export async function validQuery(db:DB, query:string): Promise<{value: boolean, message?: string}> {
	return new Promise((resolve) => {
		db.run(query, (err) => {
			if (err !== null) {
				resolve({
					value: false,
					message: err.message
				});
			} else {
				resolve({ value: true});
			}
		});
	});
}

export function hasCreateStatement(query:string) {
	return query.toLowerCase().startsWith('create')
		? `Query has a CREATE statement. 
	Let us handle that for you!
	Just use SELECT and we'll do the rest.
	`
		: false;
}

export function containsMultipleQueries(query:string) {
	return query.split().filter((character) => character == ';').length > 1
		? 'Keep it to a single query please!'
		: false;
}

export function validateQuery(query:string, ...validators:Function[]) {
	return validators.map((validator) => validator(query)).filter((validation) => validation);
}

function wrapQueryAsTemporaryView(query:string) {
	return `CREATE OR REPLACE TEMPORARY VIEW tmp AS (
	${query.replace(';', '')}
);`;
}

export async function checkQuery(query:string) : Promise<void> {
	const output = {};
	const isValid = await validQuery(db, query);
	if (!(isValid.value)) {
		throw Error(isValid.message);
	}
	const validation = validateQuery(query, hasCreateStatement, containsMultipleQueries);
	if (validation.length) {
		throw Error(validation[0])
	}
}

export async function wrapQueryAsView(query:string) {
	return new Promise((resolve, reject) => {
		db.run(wrapQueryAsTemporaryView(query), (err) => {
			if (err !== null) reject(err);
			resolve(true);
		})
	})
}

export async function createPreview(query:string) {
    // FIXME: sort out the type here
	let preview:any;
    try {
		try {
			// get the preview.
			preview = await dbAll(db, 'SELECT * from tmp LIMIT 25;');
		} catch (err) {
			throw Error(err);
		}
	} catch (err) {
		throw Error(err)
	}
    return preview;
}

export async function createSourceProfile(query:string) {
	// capture output from parquet query.
	let re = /'.*\.parquet'/g;
	const matches = query.match(re);
	const tables = (matches === null) ? [] : await Promise.all(matches.map(async (match) => {
		const info = await dbAll(db, `select * from parquet_schema(${match});`) as any[];
		const head = await dbAll(db, `SELECT * from ${match} LIMIT 1;`);
		// @ts-ignore
		const [cardinality] = await dbAll(db, `select count(*) as count FROM ${match};`);
		const output = await dbAll(db, `SELECT total_compressed_size from parquet_metadata(${match})`) as any[];
		return {
			info: info.filter(i => i.name !== 'duckdb_schema'),
			head, 
			cardinality: cardinality.count,
			table: match,
			size: output.reduce((acc:number,v) => acc + v.total_compressed_size, 0)
		}
	}))
	return tables;
}

export async function getDestinationSize(path:string) {
	if (fs.existsSync(path)) {
		const size = await dbAll(db, `SELECT total_compressed_size from parquet_metadata('${path}')`) as any[];
		return size.reduce((acc:number, v:object) => acc + v.total_compressed_size, 0)
	}
	return undefined;
}

export async function calculateDestinationCardinality(query:string) {
	const [outputSize] = await dbAll(db, 'SELECT count(*) AS cardinality from tmp;') as any[];
	return outputSize.cardinality;
}

export async function createDestinationProfile(query:string) {
	const info = await dbAll(db, `PRAGMA table_info(tmp);`);
	return info;
}

export async function exportToParquet(query:string, output:string) {
	// generate export just in case.
	if (!fs.existsSync('./export')) {
		fs.mkdirSync('./export');
	}
	const exportQuery = `COPY (${query.replace(';', '')}) TO '${output}' (FORMAT 'parquet')`;
	return dbRun(exportQuery);
}