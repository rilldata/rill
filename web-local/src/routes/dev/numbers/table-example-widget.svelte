<script lang="ts">
  import { number } from "yup/lib/locale";
  import AlignedNumber from "./aligned-number.svelte";

  import { numberLists } from "./number-samples";
  import {
    formatterFactories,
    NumberFormatter,
    RichFormatNumber,
  } from "./number-to-string-formatters";

  export let defaultFormatterIndex = 1;
  // import { numStrToAlignedNumSpec } from "./num-string-to-aligned-spec";

  let alignDecimalPoints = false;
  let alignSuffixes = false;
  let lowerCaseEForEng = false;
  let minimumSignificantDigits = 3;
  let maximumSignificantDigits = 5;

  // $: formatterOptions = { minimumSignificantDigits, maximumSignificantDigits };

  // $: selectedFormatter = formatterFactories[defaultFormatterIndex];
  // $: selectedFormatterForSamples = Object.fromEntries(
  //   numberLists.map((nl) => {
  //     return [nl.desc, selectedFormatter.fn(nl.sample, formatterOptions)];
  //   })
  // );
  // let formatterOptions;
  let selectedFormatter = formatterFactories[defaultFormatterIndex];
  let selectedFormatterForSamples: { [colName: string]: NumberFormatter };

  $: formatterOptions = { minimumSignificantDigits, maximumSignificantDigits };

  $: {
    console.log("something updated", Date.now());
    selectedFormatterForSamples = Object.fromEntries(
      numberLists.map((nl) => {
        return [nl.desc, selectedFormatter.fn(nl.sample, formatterOptions)];
      })
    );

    console.log({ selectedFormatterForSamples });
  }

  let numericType;

  let formattedColumns: { [colName: string]: RichFormatNumber[] };
  let firstColSample: RichFormatNumber[];

  $: {
    console.log("data columns updated", Date.now());
    formattedColumns = Object.fromEntries(
      numberLists.map((nl) => {
        const formatter = selectedFormatter.fn(nl.sample, formatterOptions);
        return [nl.desc, nl.sample.map(formatter)];
      })
    );
    formattedColumns = formattedColumns;
    firstColSample = formattedColumns[Object.keys(formattedColumns)[0]];
    firstColSample = firstColSample;

    console.log({ formattedColumns });
    console.log({ firstColSample });
  }
</script>

<div>
  <form>
    <label>
      <input
        type="radio"
        bind:group={numericType}
        name="number"
        value={"number"}
      />
      plain numbers (humanize)
    </label>

    <label>
      <input
        type="radio"
        bind:group={numericType}
        name="currency"
        value={"currency"}
      />
      currency
    </label>

    <label>
      <input
        type="radio"
        bind:group={numericType}
        name="percentage"
        value={"percentage"}
      />
      percentages
    </label>
  </form>
</div>

<div>
  base formatter
  <select bind:value={selectedFormatter}>
    {#each formatterFactories as formatFactory}
      <option value={formatFactory}>
        {formatFactory.desc}
      </option>
    {/each}
  </select>
</div>
formatter options
<div>
  <label>
    <input type="checkbox" bind:checked={alignDecimalPoints} />
    align decimal points
  </label>
</div>
<div>
  <label>
    <input type="checkbox" bind:checked={alignSuffixes} />
    align suffixes (requires "aligns decimal points")
  </label>
</div>
<div>
  <label>
    <input type="checkbox" bind:checked={lowerCaseEForEng} />
    force lower case "e" for exponential variants
  </label>
</div>
<div>
  significant digits:
  <label>
    min
    <input
      class="number-input"
      type="number"
      bind:value={minimumSignificantDigits}
    />
  </label>
  <label>
    max
    <input
      class="number-input"
      type="number"
      bind:value={maximumSignificantDigits}
    />
  </label>
</div>

<table class="ui-copy-number fixed-width-cols">
  <thead>
    {#each numberLists as { desc, sample }, _i}
      <td>{desc}</td>
    {/each}
  </thead>
  {#each numberLists[0].sample as _, i}
    <tr>
      {#each numberLists as { desc, sample }}
        {@const richNum = selectedFormatterForSamples[desc](sample[i])}

        <td class="table-body" title={sample[i].toString()}>
          <div class="align-content-right">
            <AlignedNumber
              {richNum}
              alignSuffix={alignSuffixes}
              {alignDecimalPoints}
              {lowerCaseEForEng}
            />
          </div>
        </td>
      {/each}
    </tr>
  {/each}
</table>

<table class="ui-copy-number fixed-width-cols">
  <thead>
    {#each Object.keys(formattedColumns) as sampleName}
      <td>{sampleName}</td>
    {/each}
  </thead>
  {#each firstColSample as richNum_, i (richNum_.number)}
    {@const rowOfValuesNums = Object.values(formattedColumns).map(
      (col) => col[i]
    )}
    <tr>
      {#each rowOfValuesNums as richNum}
        <td class="table-body" title={richNum.number.toString()}>
          <div class="align-content-right">
            <AlignedNumber
              {richNum}
              alignSuffix={alignSuffixes}
              {alignDecimalPoints}
              {lowerCaseEForEng}
            />
          </div>
        </td>
      {/each}
    </tr>
  {/each}
</table>

<!-- 
<table class="ui-copy-number fixed-width-cols">
  <thead>
    {#each numberLists as { desc, sample }, _i}
      <td>{desc}</td>
    {/each}
  </thead>
  {#each numberLists[0].sample as _, i}
    {@const rowOfRichNums = Object.entries(numberLists).map(([desc, sample]) =>
      selectedFormatterForSamples[desc](sample[i])
    )}
    <tr>
      {#each rowOfRichNums as richNum}
        <td class="table-body" title={richNum.number.toString()}>
          <div class="align-content-right">
            <AlignedNumber
              {richNum}
              alignSuffix={alignSuffixes}
              {alignDecimalPoints}
              {lowerCaseEForEng}
            />
          </div>
        </td>
      {/each}
    </tr>
  {/each}
</table> -->
<style>
  thead td {
    text-align: right;
    padding-left: 20px;
    padding-bottom: 3px;

    border-bottom: 1px solid rgb(210, 208, 208);
  }
  td.table-body {
    /* text-align: right; */
    padding-left: 30px;
    white-space: nowrap;
  }

  .align-content-right {
    display: flex;
    justify-content: flex-end;
    align-content: flex-end;
    flex-direction: row;
  }

  table {
    margin-top: 20px;
    margin-bottom: 20px;
  }

  table.fixed-width-cols td {
    width: 120px;
  }

  .number-input {
    width: 40px;
    padding-left: 6px;
    outline: solid black 1px;
  }
</style>
