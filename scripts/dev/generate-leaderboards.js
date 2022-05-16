/**
 * generate-leaderboards.js
 * ------------------------
 * 
 * Usage:
 * node scripts/generate_leaderboards path/to/a-duckdb-database.db
 * 
 * This is a script to aid in the development of components for developers.
 * The goal is to pass in a duckdb database and output a number of JSON
 * files representing leaderboards with a count(*) metric.
 * ./src/routes/dev/leaderboard/.
 * 
 */

import duckdb from "duckdb";
import fs from "fs";

async function dbAll(db, query) {
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

if (!process.argv[2]) {
    throw Error('Please supply a duckdb database.')
}

const db = new duckdb.Database(process.argv[2]);

function fixUpName(inputString) {
    let str = inputString.toLowerCase();
    str = str.replace(/_/g, ' ');
    return str.split(' ').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ');
}

const tables = await dbAll(db, 'PRAGMA show_tables;');
if (fs.existsSync('./src/routes/dev/leaderboard/data/')) {
    fs.rmSync('./src/routes/dev/leaderboard/data/',  { recursive: true });
}
fs.mkdirSync('./src/routes/dev/leaderboard/data/');
for (let { name } of tables) {
    // get column names.
    const columns = await dbAll(db, `PRAGMA table_info('${name}');`)
    // select out only VARCHAR column names.
    const onlyVarchars = columns.filter(column => column.type.toUpperCase() === 'VARCHAR');
    // generate a leaderboard per.
    const [ total ] = await dbAll(db, `SELECt count(*) as c from ${name}`);
    let leaderboards = {
        displayName: fixUpName(name),
        leaderboards: [],
        total: total.c
    };
    for (let column of onlyVarchars) {

        // run the query here.
        const data = await dbAll(db, `
            SELECT count(*) AS value, "${column.name}" as label
            FROM "${name}"
            WHERE label IS NOT NULL
            GROUP BY "${column.name}"
            ORDER BY value desc
            LIMIT 15;
        `)
        const [nulls] = await dbAll(db, `
            SELECT count(*) as c from "${name}"
            WHERE "${column.name}" IS NULL;
        `)
        leaderboards.leaderboards.push({
            values: data,
            nullCount: nulls.c,
            displayName: fixUpName(column.name),
        })
    }
    // save to disk.
    fs.writeFileSync(`./src/routes/dev/leaderboard/data/${name}.json`, JSON.stringify(leaderboards, null, 2));
    console.log(`wrote ${name}.json`)
}