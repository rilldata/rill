// @ts-nocheck
import fs from "fs";
import duckdb from 'duckdb';
import { default as glob } from 'glob';

interface DB {
	all: Function;
	exec: Function;
	run: Function
}

export function connect() : DB {
	return new duckdb.Database(':memory:');
}

const db:DB = connect();

let onCallback;
let offCallback;

/** utilize these for setting the "running" and "not running" state in the frontend */
export function registerDBRunCallbacks(onCall:Function, offCall:Function) {
	onCallback = onCall;
	offCallback = offCall;
}

function dbAll(db:DB, query:string) {
	if (onCallback) {
		onCallback();
	}
	return new Promise((resolve, reject) => {
		try {
			db.all(query, (err, res) => {
				if (err !== null) {
					reject(err);
				} else {
					if (offCallback) offCallback();
					resolve(res);
				}
			});
		} catch (err) {
			reject(err);
		}
	});
};

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

export async function createSourceProfile(parquetFile:string) {
	return await dbAll(db, `select * from parquet_schema('${parquetFile}');`) as any[];
}

export async function getCardinality(parquetFile:string) {
	const [cardinality] =  await dbAll(db, `select count(*) as count FROM '${parquetFile}';`);
	return cardinality.count;
}

export async function getFirstN(table, n=1) {
	return  dbAll(db, `SELECT * from ${table} LIMIT ${n};`);
}

export function extractParquetFilesFromQuery(query:string) {
	let re = /'[^']*\.parquet'/g;
	const matches = query.match(re).map(match => match.replace(/'/g, ''));
	return matches;
}

export async function createSourceProfileFromQuery(query:string) {
	// capture output from parquet query.
	const matches = extractParquetFilesFromQuery(query);
	const tables = (matches === null) ? [] : await Promise.all(matches.map(async (strippedMatch) => {
		//let strippedMatch = match.replace(/'/g, '');
		let match = `'${strippedMatch}'`;
		const info = await createSourceProfile(strippedMatch);
		const head = await getFirstN(match);
		const cardinality = await getCardinality(strippedMatch);
		const sizeInBytes = await getDestinationSize(strippedMatch);
		return {
			profile: info.filter(i => i.name !== 'duckdb_schema'),
			head, 
			cardinality,
			table: strippedMatch,
			sizeInBytes,
			path: strippedMatch,
			name: strippedMatch.split('/').slice(-1)[0]
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

export async function getParquetFilesInRoot() {
	return new Promise((resolve, reject) => {
		glob.glob('./**/*.parquet', {ignore: ['./node_modules/', './.svelte-kit/', './build/', './src/', './tsc-tmp']},
			(err, output) => {
				if (err!==null) reject(err);
				resolve(output);
			}
		)
	});
}
/**
 * getSummary
 * number: five number summary + mean
 * date: max, min, total time between the two
 * categorical: cardinality
 */

 export function toDistributionSummary(column:string) {
	return [
		`min(${column}) as min_${column}`,
		`approx_quantile(${column}, 0.25) as q25_${column}`,
		`approx_quantile(${column}, 0.5)  as q50_${column}`,
		`approx_quantile(${column}, 0.75) as q75_${column}`,
		`max(${column}) as max_${column}`,
		`avg(${column}) as mean_${column}`,
		`stddev_pop(${column}) as sd_${column}`,
	]
}

export async function getDistributionSummary(parquetFilePath:string, column:string) {
	const [point] = await dbAll(db, `
SELECT 
	min(${column}) as min, 
	approx_quantile(${column}, 0.25) as q25, 
	approx_quantile(${column}, 0.5)  as q50,
	approx_quantile(${column}, 0.75) as q75,
	max(${column}) as max,
	avg(${column}) as mean,
	stddev_pop(${column}) as sd
	FROM '${parquetFilePath}';`);
	return point;
}