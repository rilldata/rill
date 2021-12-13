import duckdb from 'duckdb';

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
	return new duckdb.Database('./microfiche.duckdb', { read_only: true });
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

export async function getInputTables(db, parentNode) {
	const tables = getTableScans(parentNode);
	return Promise.all(
		tables.map(async (tableName) => {
			const info = await getTableInfo(db, tableName);
			const head = await dbAll(db, `SELECT * from ${tableName} LIMIT 1;`);
			return {
				info,
				table: tableName,
				head
			};
		})
	);
}
