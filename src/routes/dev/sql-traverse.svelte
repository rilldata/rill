<script lang="ts">
  import { browser } from "$app/env";
  import Editor from "$lib/components/Editor.svelte";

  let value;

  let location = 0;
  let content = "";

  function scanMatches(str, location) {
    // get all select statements
    const re = /[\(]+[\s]*select[\s]/gi;
    let match;
    let matches = [];
    while ((match = re.exec(str)) != null) {
      matches.push({
        index: match.index,
        content: match[0],
      });
    }

    let matchSet = [];
    matches.forEach(({ index }) => {
      // scan to the right
      let ri = index + 1;
      let nest = 0;
      while (ri < str.length) {
        let char = str[ri];
        if (char === ")" && nest === 0) {
          break;
        }
        if (char === ")") {
          nest -= 1;
        }
        if (char === "(") {
          nest += 1;
        }
        ri += 1;
      }
      const substring = str.slice(index, ri + 1);
      matchSet.push({
        substring,
        start: index,
        end: ri,
        length: substring.length,
      });
      // let's scan the main string from the index.
    });

    let candidates = matchSet.filter(
      (match) => location >= match.start && location <= match.end
    );

    // get the smallest one, since that represents most nested (the most nested can't be a longer length than the next one up.)
    candidates.sort((a, b) => {
      return a.length - b.length;
    });
    let finalMatch = candidates.length
      ? candidates[0]
      : { substring: str, start: 0, end: str.length - 1, length: str.length };

    return finalMatch;
  }

  function getSubquery(str, location) {
    const final = scanMatches(str, location);
    console.log(final.substring.slice(1, -1));
    return {
      view: `${str.slice(
        0,
        final.start
      )}<span style="font-weight:bold; color: black;">${
        final.substring
      }</span>${str.slice(final.end + 1)}`,
    };
  }

  function extractCTEs(string, location) {
    if (!string) return undefined;
    const withExpressionStartPoint = string.toLowerCase().indexOf("with ");
    let si = withExpressionStartPoint + "WITH ".length;
    if (si === -1) return undefined;
    const CTEs = [];
    // set the tape.
    let ri = si;
    // the expression index.
    let ei;
    let nest = 0;
    let inside = false;
    let currentExpression: {
      name?: string;
      start?: number;
      end?: number;
      substring?: string;
    } = {};
    while (ri < string.length) {
      // let's get the name of this thing
      let char = string[ri];
      if (string.slice(si, ri).toLowerCase().endsWith(" as")) {
        currentExpression.name = string
          .slice(si, ri - 3)
          .replace(",", "")
          .trim();
      }

      if (char === "(") {
        nest += 1;
        if (!inside) {
          inside = true;
          ei = ri;
        }
      }
      if (char === ")") {
        nest -= 1;
      }

      if (char === ")" && nest === 0) {
        // we reset.
        currentExpression.start = ei;
        currentExpression.end = ri;
        currentExpression.substring = string
          .slice(ei, ri + 1)
          .slice(1, -1)
          .trim();
        CTEs.push({ ...currentExpression });
        si = ri + 1;

        currentExpression = {};
        nest = 0;
        inside = false;
      }

      ri += 1;
    }
    return CTEs;
  }

  function trackDownFrom(query) {}

  ////
  $: parse = getSubquery(content, location);
  $: ctes = extractCTEs(content, location);
</script>

<div style="width: 1600px;;" class="p-5 grid grid-flow-col gap-8">
  <Editor
    name="traversal test"
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

  <div style="font-size:12px;">
    {#if parse}
      <pre style="color:hsl({1}, 20%, 70%)">{@html parse.view}</pre>
    {/if}
    <!-- {#if parse2}
            {#each parse2 as cte}
                {cte.name}
                <pre>
                    {cte.substring}
                </pre>
            {/each}
        {/if} -->
  </div>
</div>
