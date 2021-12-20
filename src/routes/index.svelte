<script>
import { setContext } from "svelte";
import { createStore } from '$lib/app-store';
import { browser } from "$app/env";

import AddIcon from "$lib/components/icons/AddIcon.svelte";
import RefreshIcon from "$lib/components/icons/RefreshIcon.svelte";
import Logo from "$lib/components/Logo.svelte";

import EditorPane from "./_panes/EditorPane.svelte";
import InspectorPane from "./_panes/InspectorPane.svelte";

`
--SELECT events.pageId from events
--JOIN pages ON pages.pageId = events.pageId
--JOIN articles ON articles.pageId = events.pageId
--LIMIT 100;
WITH 
events_count AS (
  SELECT 
    COUNT(*) as count, 
  date(datetime(pages.createdAt / 1000, 'unixepoch')) AS dt 
  FROM events 
  JOIN pages ON events.pageId = pages.pageId 
  GROUP BY dt),
page_visit_count AS (
  SELECT COUNT(*) as count, 
  date(datetime(createdAt / 1000, 'unixepoch')) AS dt 
  FROM pages 
  GROUP BY dt),
article_count AS (
  SELECT 
    COUNT(*) as count, 
    date(datetime(pages.createdAt / 1000, 'unixepoch')) as dt 
FROM articles JOIN pages ON pages.pageId = articles.pageId GROUP BY dt)
SELECT 
  page_visit_count.count AS page_count,
  events_count.count AS event_count,
  COALESCE(article_count.count, 0) AS article_count,
  events_count.dt
FROM events_count
LEFT OUTER JOIN page_visit_count ON events_count.dt = page_visit_count.dt
LEFT OUTER JOIN article_count ON events_count.dt = article_count.dt;
`


const another = `
SELECT
    pageId,
    nextTimestamp - timestamp AS duration,
    timestamp as startTime
    FROM
    (SELECT 
    timestamp,
    eventType,
    pageId,
    LEAD(timestamp, 1) OVER (ORDER BY pageId, timestamp) AS nextTimestamp,
    LEAD(eventType, 1) OVER (ORDER BY pageId, timestamp) AS nextEvent,
    LEAD(pageId, 1) OVER (ORDER BY pageId, timestamp) AS nextPageId
        FROM
    -- subquery to order all events by page ID and timestamp first.
        (SELECT *
        FROM events
        ORDER BY pageId, timestamp
        )
)
WHERE pageId = nextPageId
AND (
    (eventType = 'attention-start' and nextEvent = 'attention-stop') OR
    (eventType = 'attention-start' and nextEvent = 'page-visit-stop')
)
`

const lastOne = `
WITH events_count AS (
  SELECT 
    COUNT(*) as count, 
  strftime(epoch_ms(pages.createdAt), '%Y-%m-%d') AS dt 
  FROM events 
  JOIN pages ON events.pageId = pages.pageId 
  GROUP BY dt),
page_visit_count AS (
  SELECT COUNT(*) as count, 
  strftime(epoch_ms(pages.createdAt), '%Y-%m-%d') AS dt 
  FROM pages 
  GROUP BY dt),
article_count AS (
  SELECT 
    COUNT(*) as count, 
    strftime(epoch_ms(pages.createdAt), '%Y-%m-%d') AS dt 
FROM articles JOIN pages ON pages.pageId = articles.pageId GROUP BY dt)
SELECT 
  page_visit_count.count AS page_count,
  events_count.count AS event_count,
  COALESCE(article_count.count, 0) AS article_count,
  events_count.dt
FROM events_count
LEFT OUTER JOIN page_visit_count ON events_count.dt = page_visit_count.dt
LEFT OUTER JOIN article_count ON events_count.dt = article_count.dt;
`


let resultset;
let queryInfo;
let query;
let destinationInfo;

// FIXME: this is out of control :(
let destinationSize;

let store;

if (browser) {
  store = createStore();
  setContext('rill:app:store', store);
}



</script>

<header class="header">
  <h1><Logo /></h1>
  <button  on:click={() => store.action("addQuery")}><AddIcon size={18} /></button>
  <button on:click={() => store.action('reset')}>
      <RefreshIcon size={18} />
  </button>
</header>
<div class='body'>
  <div class="pane inputs">
    <EditorPane bind:destinationSize bind:queryInfo bind:resultset bind:query bind:destinationInfo />
  </div>

  <div class='pane outputs'>
    <InspectorPane />
    </div>
  </div>

<style>
.body {
  width: calc(100vw);
  display: grid;
  grid-template-columns: calc(100vw - var(--right-sidebar-width, 450px)) auto;
  align-content: stretch;
  min-height: calc(100vh - var(--header-height));
}

header {
  box-sizing: border-box;
  margin:0;
  background: linear-gradient(to right, hsl(300, 30%, 14%), hsl(300, 60%, 18%));
  color: white;
  height: var(--header-height);
  display: grid;
  justify-items: left;
  justify-content: start;
  align-items: stretch;
  align-content: stretch;
  grid-auto-flow: column;
}

header h1 {
  font-size:13px;
  font-weight: normal;
  margin:0;
  padding:0;
  display: grid;
  place-items: center;
  padding: 0px 12px;
  padding-left: 2px;
  margin-left: 1rem;
}

header button {
  color: white;
  background-color: transparent;
  display: grid;
  place-items: center;
  padding: 0px 12px;
  border:none;
  font-size: 1.5rem;
}

header button:hover {
  background-color: hsla(var(--hue), var(--sat), var(--lgt), .1);
}

.inputs {
  --hue: 217;
  --sat: 20%;
  --lgt: 95%;
  --bg: hsl(var(--hue), var(--sat), var(--lgt));
  --bg-transparent: hsla(var(--hue), var(--sat), var(--lgt), .8);
  background-color: var(--bg);
  height: calc(100vh - var(--header-height));
  overflow-y: auto;
}


.pane {
  box-sizing: border-box;
}

.outputs {
  /* padding: 1rem; */
}

.pane:first-child {
  border-right: 1px solid #ddd;
}

.pane.outputs {
  height: calc(100vh - var(--header-height));
  overflow-y: auto;
  overflow-x: hidden;
}

</style>