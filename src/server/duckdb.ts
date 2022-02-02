/**
 * A single-process duckdb engine.
 */


// @ts-nocheck
import fs from "fs";
import duckdb from 'duckdb';
import { default as glob } from 'glob';
// we will use d3's binning for now until we can
// debug my histogram query
import { bin } from "d3-array";
import { sanitizeQuery } from "../lib/util/sanitize-query.js";
import { guidGenerator } from "../lib/util/guid.js";
import { rollupQuery } from "./explore-api.js"

interface DB {
	all: Function;
	exec: Function;
	run: Function
}

export function connect() : DB {
	return new duckdb.Database(':memory:');
}

const db:DB = connect();
db.exec(`
PRAGMA threads=32;
PRAGMA log_query_path='./log';
`);

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
		db.prepare(query, (err) => {
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

export function runDataModelerValidationQueries(query:string, ...validators:Function[]) {
	return validators.map((validator) => validator(query)).filter((validation) => validation);
}

export async function validateQuery(query:string) : Promise<void> {
	const output = {};
	const isValid = await validQuery(db, query);
	if (!(isValid.value)) {
		throw Error(isValid.message);
	}
	const validation = runDataModelerValidationQueries(query, hasCreateStatement, containsMultipleQueries);
	if (validation.length) {
		throw Error(validation[0])
	}
}

function wrapQueryAsTemporaryView(query:string, newTableName:string) {
	return `
-- wrapQueryAsTemporaryView
CREATE OR REPLACE TEMPORARY VIEW ${newTableName} AS (
	${query.replace(';', '')}
);`;
}

export async function wrapQueryAsView(query:string, newTableName:string) {
	return new Promise((resolve, reject) => {
		db.run(wrapQueryAsTemporaryView(query, newTableName), (err) => {
			if (err !== null) reject(err);
			resolve(true);
		})
	})
}

export async function getPreviewDataset(query:string, table:string) {
    // FIXME: sort out the type here
	let preview:any;
    try {
		try {
			// get the preview.
			preview = await dbAll(db, `SELECT * from ${table} LIMIT 25;`);
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

export async function materializeTable(tableName:string, query:string) {
	// check for table
	await dbAll(db, `-- wrapQueryAsTemporaryView
DROP TABLE IF EXISTS ${tableName}`);
	const sanitizedQuery = sanitizeQuery(query);
	return dbAll(db, `-- wrapQueryAsTemporaryView
CREATE TABLE ${tableName} AS ${sanitizedQuery}`);
}

export async function createViewOfQuery(tableName:string, query:string) {
	// check for table
	await dbAll(db, `-- createViewOfQuery
DROP VIEW IF EXISTS ${tableName}`);
	const sanitizedQuery = sanitizeQuery(query);
	return dbAll(db, `-- createViewOfQuery
CREATE TEMP VIEW ${tableName} AS ${sanitizedQuery}`);
}

export async function parquetToDBTypes(parquetFile:string) {
	const guid = guidGenerator().replace(/-/g, '_');
    await dbAll(db, `-- parquetToDBTypes
	CREATE TEMP TABLE tbl_${guid} AS (
        SELECT * from '${parquetFile}' LIMIT 1
    );
	`);
	const tableDef = await dbAll(db, `-- parquetToDBTypes
PRAGMA table_info(tbl_${guid});`)
	await dbAll(db, `DROP TABLE tbl_${guid};`);
    return tableDef;
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
	const matches = query.match(re);
	if (matches === null) { return null };
	return matches.map(match => match.replace(/'/g, ''));;
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

export async function getTransformRowCardinality(query:string, table:string) {
	const [outputSize] = await dbAll(db, `SELECT count(*) AS cardinality from ${table};`) as any[];
	return outputSize.cardinality;
}

export async function createDestinationProfile(table:string) {
	const info = await dbAll(db, `PRAGMA table_info(${table});`);
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

export function toDistributionSummary(field:string) {
	//const quotedField = `'${field}'`;
	return [
		`min(${field}) as min`,
		`reservoir_quantile(${field}, 0.25) as q25`,
		`reservoir_quantile(${field}, 0.5)  as q50`,
		`reservoir_quantile(${field}, 0.75) as q75`,
		`max(${field}) as max`,
		`avg(${field})::FLOAT as mean`,
		`stddev_pop(${field}) as sd`,
	]
}

function topK(tablePath, field:string, func = 'count(*)') {
	//const quotedField = `'${field}'`;
	return `SELECT ${field} as value, ${func} AS count from ${tablePath}
GROUP BY ${field}
ORDER BY count desc
LIMIT 50;`
}


export async function getTopKAndCardinality(tablePath, column, func = 'count(*)', dbEngine = db) {
	const topKValues = await dbAll(dbEngine, topK(tablePath, column, func));
	const [cardinality] = await dbAll(dbEngine = db, `SELECT approx_count_distinct(${column}) as count from ${tablePath};`);
	return {
		column,
		topK: topKValues,
		cardinality: cardinality.count
	}
}

export async function getNullCount(tablePath:string, field:string, dbEngine = db) {
	const [nullity] = await dbAll(dbEngine, `
		SELECT COUNT(*) as count FROM ${tablePath} WHERE ${field} IS NULL;
	`);
	return nullity.count;
}

export async function getNullCounts(tablePath:string, fields:any, dbEngine = db) {
	const [nullities] = await dbAll(dbEngine, `
		SELECT
		${fields.map(field => {
			return `COUNT(CASE WHEN ${field.name} IS NULL THEN 1 ELSE NULL END) as ${field.name}`
		}).join(',\n')}
		FROM ${tablePath};
	`);
	return nullities;
}

export async function descriptiveStatistics(tablePath:string, field:string, fieldType: string, dbEngine = db) {
	const query = `SELECT ${toDistributionSummary(field)} FROM ${tablePath};`;
	const [results] = await dbAll(dbEngine, query);
	return results;
}

export async function getTimeRange(tablePath:string, field:any, dbEngine = db) {
	const [ranges] = await dbAll(dbEngine, `
	SELECT
		min(${field}) as min, max(${field}) as max, 
		max(${field}) - min(${field}) as interval
		FROM ${tablePath};
	`)
	return ranges;
}

// FIXME: this doesn't work as expected.
export async function numericHistogram(tablePath:string, field:string, fieldType:string, dbEngine = db) {
	// if the field type is an integer and the total number of values is low, can't we just use
	// first check a sample to see how many buckets there are for this value.
	//const quotedField = `'${field}'`;

	const buckets = await dbAll(dbEngine, `SELECT count(*) as count, ${field} FROM ${tablePath} WHERE ${field} IS NOT NULL GROUP BY ${field} USING SAMPLE reservoir(1000 ROWS);`)
	const bucketSize = Math.min(40, buckets.length);
	return dbAll(dbEngine, `
	WITH dataset AS (
		SELECT ${fieldType === 'TIMESTAMP' ? `epoch(${field})` : `${field}::DOUBLE`} as ${field} FROM ${tablePath}
	) , S AS (
		SELECT 
			min(${field}) as minVal,
			max(${field}) as maxVal,
			(max(${field}) - min(${field})) as range
			FROM dataset
	), values AS (
		SELECT ${field} as value from dataset
		WHERE ${field} IS NOT NULL
	), buckets AS (
		SELECT
			range as bucket,
			(range) * (select range FROM S) / ${bucketSize} + (select minVal from S) as low,
			(range + 1) * (select range FROM S) / ${bucketSize} + (select minVal from S) as high
		FROM range(0, ${bucketSize}, 1)
	)
	, histogram_stage AS (
		SELECT
			bucket,
			low,
			high,
			count(values.value) as count
		FROM buckets
		LEFT JOIN values ON values.value BETWEEN low and high
		GROUP BY bucket, low, high
		ORDER BY BUCKET
	)
	SELECT 
		bucket,
		low,
		high,
		CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + 1 ELSE count END AS count
		FROM histogram_stage;
	
	`)
}

export async function numericHistogram_d3(tablePath:string, field:string, fieldType:string, dbEngine = db) {
	// const buckets = await dbAll(dbEngine, `SELECT count(*) as count, ${field} FROM ${tablePath} WHERE ${field} IS NOT NULL GROUP BY ${field} USING SAMPLE reservoir(1000 ROWS);`)
	// const bucketSize = Math.min(40, buckets.length);
	const results = dbAll(dbEngine, `SELECT ${fieldType === 'TIMESTAMP' ? `epoch(${field})` : `${field}::DOUBLE`} as ${field} FROM ${tablePath}`);
	const binFcn = bin();
	// binFcn.
	const binned = binFcn(results.map(result => result[field]));
	return binned.map((binnedData, i:number) => {
		return {
			bucket: i,
			count: binnedData.length,
			low: binnedData.x0,
			high: binnedData.x1
		};
	})
}

export async function generateExploreTimeseries({ table, metrics, dimensions = [], timeField, timeGrain}) {
	const query = rollupQuery({ table, timeField, timeGrain, metrics, dimensions });
	const results = await dbAll(db, query);
	return results;
}

export function generateExploreLeaderboard({
	table, leaderboardMetric, dimension, timeField, timeMin = undefined, timeMax = undefined
}) {
	const lowWhere = timeMin && `${timeField} >= TIMESTAMP '${timeMin.toISOString()}'`;
    const highWhere = timeMax && `${timeField} <= TIMESTAMP '${timeMax.toISOString()}'`;
    let whereClauses = []
    if (lowWhere) whereClauses.push(lowWhere)
    if (highWhere) whereClauses.push(highWhere)
	dimensions.forEach((slice) => {
		whereClauses.push(`${slice.field} = ${typeof slice.value === 'string' ? `'${slice.value}'` : slice.value}`);
	})
    const whereClause = lowWhere || highWhere ? `WHERE ${whereClauses.join(' AND ')}` : '';
	const query = `SELECT ${dimension} as value, ${leaderboardMetric} AS count from ${table}
${whereClause}
GROUP BY ${dimension}
ORDER BY count desc
LIMIT 50;`
	return dbAll(db, query);
}
