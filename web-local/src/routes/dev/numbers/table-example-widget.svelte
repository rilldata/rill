<script lang="ts">
  import { number } from "yup/lib/locale";
  import AlignedNumber from "./aligned-number.svelte";
  import ColorPicker from "svelte-awesome-color-picker";
  // import { HsvPicker } from "svelte-color-picker";
  // import {  } from "svelte";
  import { numberLists } from "./number-samples";
  import {
    formatterFactories,
    NumberFormatter,
    NumPartPxWidthLookupFn,
    RichFormatNumber,
  } from "./number-to-string-formatters";
  import { onMount } from "svelte";
  import { writable } from "svelte/store";

  export let defaultFormatterIndex = 1;
  let alignDecimalPoints = true;
  let alignSuffixes = true;
  let suffixPadding = 2;

  let lowerCaseEForEng = true;
  let minimumSignificantDigits = 3;
  let maximumSignificantDigits = 5;

  let onlyUseLargestMagnitude = false;
  let usePlainNumsForThousands = true;
  let usePlainNumsForThousandsOneDecimal = false;
  let usePlainNumForThousandths = true;
  let usePlainNumForThousandthsPadZeros = false;

  let truncateThousandths = true;
  let truncateTinyOrdersIfBigOrderExists = true;
  let zeroHandling: "exactZero" | "noSpecial" | "zeroDot" = "exactZero";

  let showBars = true;

  let negativeColor = "#ffbebe";
  let positiveColor = "#eaeaea";

  let showBaseline = true;
  let baselineColor = "#eeeeee";

  let selectedFormatter = formatterFactories[defaultFormatterIndex];
  let selectedFormatterForSamples: { [colName: string]: NumberFormatter };

  $: usePlainNumForThousandthsPadZeros =
    usePlainNumForThousandths && usePlainNumForThousandthsPadZeros;

  $: formatterOptions = {
    minimumSignificantDigits,
    maximumSignificantDigits,
    magnitudeStrategy,
    usePlainNumsForThousands,
    usePlainNumsForThousandsOneDecimal,
    usePlainNumForThousandths,
    usePlainNumForThousandthsPadZeros,
    truncateThousandths,
    truncateTinyOrdersIfBigOrderExists,
    zeroHandling,
  };

  let numberInputType;
  let magnitudeStrategy = "largest";

  const blue100 = "#dbeafe";
  const grey100 = "#f5f5f5";

  // const numberAlignmentStores = numberLists.map(() =>
  //   writable({ int: 0, dot: 0, frac: 0, suffix: 0 })
  // );

  let numFormattingWidthLookupKeys = [
    ".",
    "-",
    "$",
    "%",
    "k",
    "M",
    "B",
    "T",
    "Q",
  ];
  for (let i = -79; i < 308; i++) {
    numFormattingWidthLookupKeys.push("e" + i);
    numFormattingWidthLookupKeys.push("E" + i);
  }
  for (let i = 0; i < 20; i++) {
    let thisManyZerosString = "0".repeat(i);
    numFormattingWidthLookupKeys.push(thisManyZerosString);
  }
  let numFormattingWidthLookup: { [key: string]: number } = {};

  let charMeasuringDiv: HTMLDivElement;

  let pxWidthLookupFn: NumPartPxWidthLookupFn;

  onMount(() => {
    console.time("charMeasuringDiv");
    numFormattingWidthLookupKeys.forEach((str) => {
      charMeasuringDiv.innerHTML = str;
      let rect = charMeasuringDiv.getBoundingClientRect();
      // console.log(str, cw);
      numFormattingWidthLookup[str] = rect.right - rect.left;
    });

    console.timeEnd("charMeasuringDiv");

    pxWidthLookupFn = (str: string, isNumStr: boolean) => {
      let out = 0;
      if (isNumStr) {
        let len = str.length;
        if (str !== "" && str[0] === "-") {
          out =
            numFormattingWidthLookup["-"] +
            numFormattingWidthLookup["0".repeat(len - 1)];
        } else {
          out = numFormattingWidthLookup["0".repeat(len)];
        }
      } else {
        out = numFormattingWidthLookup[str];
      }
      return isNaN(out) ? 0 : out;
    };
  });
  // console.log({ numFormattingWidthLookup });

  $: {
    if (pxWidthLookupFn !== undefined) {
      // window.pxWidthLookupFn = pxWidthLookupFn;

      selectedFormatterForSamples = Object.fromEntries(
        numberLists.map((nl) => {
          return [
            nl.desc,
            selectedFormatter.fn(nl.sample, pxWidthLookupFn, formatterOptions),
          ];
        })
      );
    }
  }
</script>

<div class="outer">
  <div class="inner ui-copy-number" bind:this={charMeasuringDiv}>CONTENT</div>
</div>

<div>
  <form>
    <label>
      <input
        type="radio"
        bind:group={numberInputType}
        name="number"
        value={"number"}
      />
      real numbers (no special treament)
    </label>

    <label>
      <input
        type="radio"
        bind:group={numberInputType}
        name="number"
        value={"number"}
      />
      integers (inputs rounded)
    </label>

    <label>
      <input
        type="radio"
        bind:group={numberInputType}
        name="currency"
        value={"currency"}
      />
      currency
    </label>

    <label>
      <input
        type="radio"
        bind:group={numberInputType}
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

<div class="options-container-row">
  <div>
    generic formatter options
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
      <div>
        <label>
          suffix padding:
          <input
            class="number-input"
            type="number"
            bind:value={suffixPadding}
          />
        </label>
      </div>
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
  </div>

  <div style="padding-left: 40px;">
    new humanizer shared options

    <div class="option-box">
      <form>
        <label>
          <input
            type="radio"
            bind:group={zeroHandling}
            name="noSpecial"
            value={"noSpecial"}
          />
          no special treament for exact zeros
        </label>
        <label>
          <input
            type="radio"
            bind:group={zeroHandling}
            name="exactZero"
            value={"exactZero"}
          />
          "0" for exact zeros
        </label>

        <label>
          <input
            type="radio"
            bind:group={zeroHandling}
            name="zeroDot"
            value={"zeroDot"}
          />
          "0." for exact zeros
        </label>
      </form>
    </div>

    new humanizer strategy
    <form>
      <div class="option-box">
        <label>
          <input
            type="radio"
            bind:group={magnitudeStrategy}
            name="largest"
            value={"largest"}
          />
          only use largest magnitude (like current humanizer)
        </label>
        <div class:inactive={magnitudeStrategy !== "largest"}>
          <div class="option-box">
            <label>
              <input type="checkbox" bind:checked={usePlainNumsForThousands} />
              for samples in interval (-1e6,1e6), just show plain number
            </label>
            <div class="option-box">
              <label>
                <input
                  type="checkbox"
                  bind:checked={usePlainNumsForThousandsOneDecimal}
                />
                show one digit after the decimal point (to indicate non-integer sample)
              </label>
            </div>
          </div>
          <div class="option-box">
            <label>
              <input type="checkbox" bind:checked={usePlainNumForThousandths} />
              show a plain number if the largest order of magnitude is thousandths
            </label>

            <div class="option-box" class:inactive={!usePlainNumForThousandths}>
              <label>
                <input
                  type="checkbox"
                  bind:checked={usePlainNumForThousandthsPadZeros}
                />
                pad with zeros
              </label>
            </div>
          </div>
        </div>
      </div>

      <div class="option-box">
        <label>
          <input
            type="radio"
            bind:group={magnitudeStrategy}
            name="unlimited"
            value={"unlimited"}
          />
          allow as many magnitudes as needed
        </label>
        <div class="option-box">
          <label>
            <input type="checkbox" bind:checked={truncateThousandths} />
            truncate and render thousandths without suffix
          </label>
        </div>
        <div class="option-box">
          <label>
            <input
              type="checkbox"
              bind:checked={truncateTinyOrdersIfBigOrderExists}
            />
            truncate tiny numbers if sample has any non-tiny numbers
          </label>
        </div>
      </div>
    </form>
  </div>
  <div style="padding-left: 40px;">
    <div class="option-box">
      <label>
        <input type="checkbox" bind:checked={showBars} />
        show bars
      </label>
      <div class="option-box">
        <ColorPicker bind:hex={negativeColor} label="negative bar color" />
        <ColorPicker bind:hex={positiveColor} label="positive bar color" />
        set positive bar color
        <button on:click={() => (positiveColor = blue100)}
          >blue-100 (like `main`)</button
        >
        &nbsp;
        <button on:click={() => (positiveColor = grey100)}>grey-100</button>
        &nbsp;
        <button on:click={() => (positiveColor = "#eeeeee")}>grey-200</button>
      </div>
      <div class="option-box">
        <label>
          <input type="checkbox" bind:checked={showBaseline} />
          show baseline
        </label>
        <div class="option-box">
          <ColorPicker bind:hex={baselineColor} label="baseline color" />
          <!-- set positive bar color
          <button on:click={() => (positiveColor = blue100)}
            >blue-100 (like `main`)</button
          >
          &nbsp;
          <button on:click={() => (positiveColor = grey100)}>grey-100</button>
          &nbsp;
          <button on:click={() => (positiveColor = "#eeeeee")}>grey-200</button> -->
        </div>
      </div>
    </div>
  </div>
</div>

{#if selectedFormatterForSamples !== undefined}
  <div class="table-container">
    <table class="ui-copy-number fixed-width-cols">
      <thead>
        {#each numberLists as { desc, sample }, _i}
          <td>{desc}</td>
        {/each}
      </thead>
      {#each numberLists[0].sample as _, i}
        <tr>
          {#each numberLists as { desc, sample }, j}
            {@const richNum = selectedFormatterForSamples[desc](sample[i])}

            <td class="table-body" title={sample[i].toString()}>
              <div class="align-content-right">
                <AlignedNumber
                  containerWidth={100}
                  {richNum}
                  alignSuffix={alignSuffixes}
                  {alignDecimalPoints}
                  {lowerCaseEForEng}
                  {zeroHandling}
                  {showBars}
                  {negativeColor}
                  {positiveColor}
                  {showBaseline}
                  {baselineColor}
                  {suffixPadding}
                />
              </div>
            </td>
          {/each}
        </tr>
      {/each}
    </table>
  </div>
{/if}

<style>
  div.table-container {
    width: 100%;
    overflow-x: scroll;
  }

  thead td {
    text-align: right;
    padding-left: 20px;
    padding-bottom: 3px;

    border-bottom: 1px solid rgb(210, 208, 208);
  }
  td.table-body {
    /* text-align: right; */
    padding: 0 0 0 0;
    white-space: nowrap;
  }

  .options-container-row {
    display: flex;
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
    min-width: 120px;
  }

  .option-box {
    padding-left: 30px;
  }

  .inactive {
    color: rgb(144, 144, 144);
    pointer-events: none;
  }

  .number-input {
    width: 40px;
    padding-left: 6px;
    outline: solid black 1px;
  }

  button {
    outline: 1px solid #ddd;
    background-color: #f2f2f2;
    padding: 3px;
    border-radius: 5px;
  }

  .outer {
    overflow: hidden;
    position: relative;
  }
  .inner {
    position: absolute;
    right: -50px;
    top: 50px;
    width: fit-content;
  }
</style>
