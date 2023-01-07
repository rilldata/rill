<script lang="ts">
  import {
    humanizeDataType,
    humanizeGroupValues,
    NicelyFormattedTypes,
  } from "@rilldata/web-local/lib/util/humanize-numbers";
  import AlignedNumber from "./aligned-number.svelte";
  import TableExampleWidget from "./table-example-widget.svelte";

  import { numberLists } from "./number-samples";
  import { formatterFactories } from "./number-to-string-formatters";

  // let selectedFormatterFactory = formatterFactories[0];
  // let alignDecimalPoints = false;
  // let alignSuffixes = false;
  // // $: alignDecimalPoints =
  // //   alignDecimalPoints === false ? alignSuffixes : alignDecimalPoints;
  // $: selectedFormatterForSamples = Object.fromEntries(
  //   numberLists.map((nl) => {
  //     return [nl.desc, selectedFormatter.fn(nl.sample)];
  //   })
  // );

  // let lowerCaseEForEng = false;
</script>

<h1 class="pb-4">Tabular / columnar number formatting</h1>

<TableExampleWidget />
<TableExampleWidget defaultFormatterIndex={2} />

<!-- 
<h1>Tables by formatter</h1>

{#each formatterFactories as formatFactory}
  {@const formattersPerSample = Object.fromEntries(
    numberLists.map((nl) => {
      return [nl.desc, formatFactory.fn(nl.sample)];
    })
  )}
  <h2>
    {formatFactory.desc}
  </h2>
  <table class="ui-copy-number fixed-width-cols">
    <thead>
      {#each numberLists as { desc, sample }, i}
        <td>{desc}</td>
      {/each}
    </thead>
    {#each numberLists[0].sample as num, i}
      <tr>
        {#each numberLists as { desc, sample }}
          {@const formattedNum = formattersPerSample[desc](sample[i])}
          <td class="table-body">{formattedNum}</td>
        {/each}
      </tr>
    {/each}
  </table>
{/each}

<h1>Tables by distribution</h1>

{#each numberLists as { desc, sample }, i}
  <h2>
    {desc}
  </h2>
  <table class="ui-copy-number">
    <thead>
      {#each formatters as formatFactory}
        <td>{formatFactory.desc}</td>
        <td>{formatFactory.desc}<br /> dec. aligned</td>
        <td>{formatFactory.desc}<br /> suff. aligned</td>
      {/each}
    </thead>
    {#each sample as num}
      <tr>
        {#each formatters as formatFactory}
          <td class="table-body"
            ><AlignedNumber {...alignedSpecFn(num)} alignSuffix />
          </td>
        {/each}
      </tr>
    {/each}
  </table>
{/each} -->
<style>
  thead td {
    text-align: right;
    padding-left: 20px;
  }
  td.table-body {
    /* text-align: right; */
    padding-left: 30px;
    white-space: nowrap;
  }

  .align-content-right {
    display: flex;
    justify-content: flex-end;
    /* align-content: end; */
    flex-direction: row;
  }

  table {
    margin-top: 20px;
    margin-bottom: 20px;
  }

  table.fixed-width-cols td {
    width: 120px;
  }
  /* td.dot-aligned {
    display: flex;
    width: 300px;

    padding-right: 30px;
  } */

  /* .before-dot {
    width: 120px;
  }
  .suffix {
    padding-left: 4px;
  } */

  h2 {
    padding-top: 30px;
    padding-bottom: 10px;
  }

  h1 {
    /* padding-top: 50px; */
    font-size: large;
  }
</style>
