// some tests for table-info
import { getInputTableInformation } from './table-info.mjs';
import { connect, justRunIt } from './db.mjs';

const db = connect();

const query01 = `
WITH events_count AS (SELECT count(*) as count, date(datetime(pages.createdAt / 1000, 'unixepoch')) as dt from events 
    JOIN pages ON events.pageId = pages.pageId 
    GROUP BY dt),
  page_visit_count AS (SELECt COUNT(*) as count, date(datetime(createdAt / 1000, 'unixepoch')) as dt from pages GROUP BY dt),
  article_count AS (SELECT COUNT(*) as count, date(datetime(pages.createdAt / 1000, 'unixepoch')) as dt 
  FROM articles JOIN pages ON pages.pageId = articles.pageId GROUP BY dt)
  SELECT 
    page_visit_count.count AS page_count,
    events_count.count AS event_count,
    article_count.count AS article_count,
    events_count.dt
  FROM events_count
  JOIN page_visit_count ON page_visit_count.dt = events_count.dt
  JOIN article_count ON article_count.dt = events_count.dt;
`;

console.log(JSON.stringify(getInputTableInformation(db, query01), null, 2));
