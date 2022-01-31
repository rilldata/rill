import duckdb from 'duckdb';
import {jest} from '@jest/globals';

import { sanitizeQuery } from "../../lib/util/sanitize-query.js";
import { rollupQuery, topKQuery } from "../explore-api";
import type { Metric, DimensionSlice } from "../explore-api";
import { dbRun } from '../duckdb.js';
jest.setTimeout(30000);
const metrics:Metric[] = [
    {
        name: 'Total Volume',
        function: 'count',
        description: 'The sales volume description',
        field: '*',
        aka: 'total_volume'
    },
    {
        name: 'Average Value',
        function: 'avg',
        description: 'The average value description',
        field: 'latitude',
    },
]

async function createDB() {
    const db = new duckdb.Database('./test.db');

    function dbAll(query) {
        return new Promise((resolve, reject) => {
            db.all(query, (err, res) => {
                if (err !== null) {reject(err)};
                resolve(res);
            })
        })
    }
    async function checkForTables() : Promise<any> {
        return dbAll(`PRAGMA show_tables`);
    }
    
    const tables = (await checkForTables()).map(table => table.name);
    
    if (!tables.includes('test')) {
        console.time('ingestion');
        await dbAll(`CREATE TABLE test AS SELECT * FROM './scripts/nyc311-reduced.parquet';`)
        console.timeEnd('ingestion');
    }

    if (!tables.includes('transformed')) {
        console.time('materialized transformation')
        await dbAll(`
        CREATE TABLE transformed AS
        SELECT
            created_date,
            incident_zip,
            agency,
            complaint_type,
            latitude,
            longitude,
            status,
            closed_date,
            resolution_action_updated_date,
            taxi_pick_up_location
        FROM test;
        `)
        console.timeEnd('materialized transformation');
        console.time('create index')
        await dbAll(`CREATE INDEX agency_idx ON transformed (agency);`)
        console.timeEnd('create index')
    }
    return {
        async run(query) {
            return dbAll(query);
        }
    }
}


describe('rollupQuery', () => {
    let db;
    beforeAll(async () => {
        //jest.setTimeout(30000);
        db = await createDB();
    })


    it('will create a query with no dimension slices', () => {
        const query  = rollupQuery({
            metrics, timeField: 'created_date', timeGrain: 'day', table: 'transformed'
        })
        const s = sanitizeQuery(query);
        expect(s).toBe(`select count(*) as total_volume, avg(latitude) as latitude, date_trunc('day', created_date) as _ts from transformed group by date_trunc('day', created_date)`)
    })
    it('will create a query with some dimension slices', () => {
        const query1  = rollupQuery({
            metrics, 
            timeField: 'created_date',
            timeGrain: 'day', 
            table: 'transformed',
            dimensions: [
                { field: 'agency', value: 'nyc'},
                { field: 'zip', value: 90210}
            ]
        })
        const query2  = rollupQuery({
            metrics, 
            timeField: 'created_date',
            timeGrain: 'hour', 
            table: 'transformed',
            dimensions: [
                { field: 'agency', value: 'nyc'},
                { field: 'zip', value: 90210}
            ]
        })
        const s1 = sanitizeQuery(query1);
        const s2 = sanitizeQuery(query2);
        expect(s1).toBe(`select count(*) as total_volume, avg(latitude) as latitude, date_trunc('day', created_date) as _ts from transformed where agency = 'nyc' and zip = 90210 group by date_trunc('day', created_date)`)
        expect(s2).toBe(`select count(*) as total_volume, avg(latitude) as latitude, date_trunc('hour', created_date) as _ts from transformed where agency = 'nyc' and zip = 90210 group by date_trunc('hour', created_date)`)
    })
    // let's integrate duckdb
    it('returns the right resultset from a small dataset in duckdb', async () => {
        console.time('daily rollup')
        const results = await db.run(rollupQuery({
            metrics, 
            timeField: 'created_date',
            timeGrain: 'day', 
            table: 'transformed'
        }))
        console.timeEnd('daily rollup')
        console.log(results.length);
        // check for uniqueness of date strings.
        console.time('hourly rollup')
        const results2 = await db.run(rollupQuery({
            metrics, 
            timeField: 'created_date',
            timeGrain: 'hour', 
            table: 'transformed'
        }))
        console.timeEnd('hourly rollup')
        console.log(results2.length);
        
    })
})

describe('topKQuery', () => {
    let db;
    beforeAll(async () => {
        //jest.setTimeout(30000);
        db = await createDB();
    })

    it('will create a top k query with no dimension slices', async () => {
        const query = topKQuery({
            metrics: [metrics[0]], table: 'transformed', field: 'agency'
        })
        const s = sanitizeQuery(query);
        expect(s).toBe(`select count(*) as total_volume, agency from transformed group by agency`);
        console.time('top-n (28)')
        const results = await db.run(query);
        console.timeEnd('top-n (28)')
        const query2 = topKQuery({
            metrics: [metrics[0]], table: 'transformed', field: 'resolution_action_updated_date'
        })
        console.time('top-n (3 mil)')
        const results2 = await db.run(query2);
        console.timeEnd('top-n (3 mil)')
    })
    it('will create a top k query with some dimension slices', () => {
        const query = topKQuery({
            metrics, table: 'transformed', field: 'agency', dimensions: [
                {field:'zip', value: 90210}
            ]
        });
        const s = sanitizeQuery(query);
        expect(s).toBe(`select count(*) as total_volume, avg(latitude) as latitude, agency from transformed where zip = 90210 group by agency`);
    })
})