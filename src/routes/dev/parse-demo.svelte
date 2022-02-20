<script>
import Editor from "$lib/components/Editor.svelte";
import { extractCTEs, getCoreQuerySelectStatements } from "$lib/util/model-structure";
let content = `
WITH dataset AS (
	SELECT epoch(created_date) as created_date FROM './scripts/nyc311-reduced.parquet'
), S AS (
	SELECT 
		min(created_date) as minVal,
		max(created_date) as maxVal,
		(max(created_date) - min(created_date)) as range
		FROM dataset
), values AS (
	SELECT created_date as value from dataset
	WHERE created_date IS NOT NULL
), buckets AS (
	SELECT
		range as bucket,
		(range - 1) * (select range FROM S) / 40 + (select minVal from S) as low,
		(range) * (select range FROM S) / 40 + (select minVal from S) as high
	FROM range(0, 40, 1)
)
, histogram_stage AS (
	SELECT
		bucket,
		low,
		high,
		count(values.value) as count
	FROM buckets
	LEFT JOIN values ON (values.value BETWEEN low and high)
	GROUP BY bucket, low, high
	ORDER BY BUCKET
)
SELECT 
	bucket,
	low,
	high,
	CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + 1 ELSE count END AS count
	FROM histogram_stage;`
let location;

$: ctes = extractCTEs(content);
$: selects = getCoreQuerySelectStatements(content || '');
</script>

<div class='grid grid-cols-2'>
<Editor 
content={`
WITH dataset AS (
	SELECT epoch(created_date) as created_date FROM './scripts/nyc311-reduced.parquet'
), S AS (
	SELECT 
		min(created_date) as minVal,
		max(created_date) as maxVal,
		(max(created_date) - min(created_date)) as range
		FROM dataset
), values AS (
	SELECT created_date as value from dataset
	WHERE created_date IS NOT NULL
), buckets AS (
	SELECT
		range as bucket,
		(range - 1) * (select range FROM S) / 40 + (select minVal from S) as low,
		(range) * (select range FROM S) / 40 + (select minVal from S) as high
	FROM range(0, 40, 1)
)
, histogram_stage AS (
	SELECT
		bucket,
		low,
		high,
		count(values.value) as count
	FROM buckets
	LEFT JOIN values ON (values.value BETWEEN low and high)
	GROUP BY bucket, low, high
	ORDER BY BUCKET
)
SELECT 
	bucket,
	low,
	high,
	CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + 1 ELSE count END AS count
	FROM histogram_stage;
`}
on:cursor-location={(event) => {
    location = event.detail.location;
    content = event.detail.content;
}}
/>

<div>

{#if ctes}
<div class="p-3">
    <div>ctes: {ctes.length}</div>
    {#each ctes as cte}
        <div class="text-ellipsis overflow-hidden whitespace-nowrap ">
            <b>{cte.name}</b> = <i>{cte.substring}</i>
        </div>
    {/each}
</div>
{/if}

{#if selects}
<div class="p-3">
    <div>selects: {selects.length}</div>
    {#each selects as select}
        <div class="text-ellipsis overflow-hidden whitespace-nowrap">
            <b>{select.name}</b> = <i>{select.expression}</i>
        </div>
    {/each}
</div>
{/if}

</div>

</div>

