import { connect } from './db.mjs';
import { cheapFirstN } from './table-info.mjs';

const db = connect();

console.log(cheapFirstN(db, 'SELECt * from events;'));

//console.log());
