import sqlite3 from 'better-sqlite3';

export function connect() {
	return sqlite3('./microfiche.db', { readonly: true });
}

export function testConnection() {
	return sqlite3(':memory:');
}

export function justRunIt(db, query) {
	return db.prepare(query).all();
}
