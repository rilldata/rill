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
  import LayeredContainer from "./layered-container.svelte";
  import RichNumberBipolarBar from "./rich-number-bipolar-bar.svelte";

  export let defaultFormatterIndex = 1;
  export let alignDecimalPoints = true;
  export let alignSuffixes = true;
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
  export let zeroHandling: "exactZero" | "noSpecial" | "zeroDot" = "exactZero";
  export let showMagSuffixForZero = false;

  let showBars = false;
  let absoluteValExtentsIfPosAndNeg = true;
  let absoluteValExtentsAlways = false;
  let barPosition: "left" | "behind" | "right" = "behind";
  let barContainerWidth = 81;
  let barOffset = 10;

  let negativeColor = "#ffbebe";
  let positiveColor = "#eaeaea";
  let barBackgroundColor = "#ffffff";

  let showBaseline = true;
  let baselineColor = "#eeeeee";

  let selectedFormatter = formatterFactories[defaultFormatterIndex];
  let selectedFormatterForSamples: { [colName: string]: NumberFormatter };

  let tableGutterWidth = 30;

  $: worstCaseStringWidth = 79 + suffixPadding;
  $: layerContainerWidth =
    worstCaseStringWidth +
    (barPosition === "behind" ? 0 : barContainerWidth + barOffset);

  $: usePlainNumForThousandthsPadZeros =
    usePlainNumForThousandths && usePlainNumForThousandthsPadZeros;

  $: formatterOptions = {
    minimumSignificantDigits,
    maximumSignificantDigits,
    magnitudeStrategy,
    digitTarget,
    digitTargetShowSignificantZeros,
    digitTargetPadWithInsignificantZeros,
    usePlainNumsForThousands,
    usePlainNumsForThousandsOneDecimal,
    usePlainNumForThousandths,
    usePlainNumForThousandthsPadZeros,
    truncateThousandths,
    truncateTinyOrdersIfBigOrderExists,
    zeroHandling,
  };

  let numberInputType;

  let magnitudeStrategy = "largestWithDigitTarget";
  let digitTargetShowSignificantZeros = true;
  let digitTargetPadWithInsignificantZeros = false;
  let digitTarget = 5;

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
      treat numbers as currency (no rounding)
    </label>

    <label>
      <input
        type="radio"
        bind:group={numberInputType}
        name="currency"
        value={"currencyRoundCent"}
      />
      treat numbers as currency (round fracs to nearest cent)
    </label>

    <label>
      <input
        type="radio"
        bind:group={numberInputType}
        name="percentage"
        value={"percentage"}
      />
      display values as percentages
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
  <div style:width="300px">
    <h2>Display options (applies to all formatters)</h2>

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
          /> px
        </label>
      </div>
    </div>
    <div>
      <label>
        <input type="checkbox" bind:checked={lowerCaseEForEng} />
        force lower case "e" for exponential variants
      </label>
    </div>

    Zero handling - zeros are semantically and mathematically important values,
    and should be represented with care, especially when juxtaposed with finite
    precision decimal representation of small numbers (e.g. "0.000" often means
    "a non-zero number, but one that rounds to zero to the third decimal of
    precision", whereas "0" can be reserved for 0
    <em>exactly</em>.)
    <div class="option-box">
      <form>
        <div>
          <label>
            <input
              type="radio"
              bind:group={zeroHandling}
              name="exactZero"
              value={"exactZero"}
            />
            "0" for exact zeros
          </label>
        </div>
        <div>
          <label>
            <input
              type="radio"
              bind:group={zeroHandling}
              name="zeroDot"
              value={"zeroDot"}
            />
            "0." for exact zeros. (Used by legacy dash)
          </label>
        </div>

        <div>
          <label>
            <input
              type="radio"
              bind:group={zeroHandling}
              name="noSpecial"
              value={"noSpecial"}
            />
            no special treament for exact zeros
          </label>
        </div>
      </form>

      <label>
        <input type="checkbox" bind:checked={showMagSuffixForZero} />
        Show order of magnitude suffix for exact zeros (not recommended -- order
        of magnitude is not relevant to 0)
      </label>
    </div>
    <!-- <div>
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
    </div> -->
  </div>

  <div style="padding-left: 40px;">
    <h2>new humanizer shared options</h2>

    <label>
      <input type="checkbox" bind:checked={digitTargetShowSignificantZeros} />
      show significant zeros
    </label>

    <h2>new humanizer strategy (and strategy-specific options)</h2>
    <form>
      <!-- <div class="option-box">
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
      </div> -->

      <div class="option-box">
        <label>
          <input
            type="radio"
            bind:group={magnitudeStrategy}
            name="largestWithDigitTarget"
            value={"largestWithDigitTarget"}
          />
          only use largest magnitude, with digit target
        </label>
        <div class:inactive={magnitudeStrategy !== "largestWithDigitTarget"}>
          <div class="option-box">
            <label>
              target num digits
              <input
                class="number-input"
                type="number"
                min="3"
                max="8"
                bind:value={digitTarget}
              />
            </label>
            <br />
            <label>
              <input
                type="checkbox"
                bind:checked={digitTargetShowSignificantZeros}
              />
              show significant zeros
            </label>
            <br />
            <label>
              <input
                type="checkbox"
                bind:checked={digitTargetPadWithInsignificantZeros}
              />
              pad with insignificant zeros (after last significant digit)
            </label>
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

  <div style="padding-left: 10px;">
    bar options

    <div class="option-box">
      <label>
        <input type="checkbox" bind:checked={showBars} />
        show bars
      </label>
      <div class="option-box">
        <label>
          <input type="checkbox" bind:checked={absoluteValExtentsIfPosAndNeg} />
          use symmetric extent if a sample has pos and neg values
        </label>
        <div class="option-box">
          <label>
            <input type="checkbox" bind:checked={absoluteValExtentsAlways} />
            always use symmetric extents
          </label>
        </div>

        <form>
          <label>
            <input
              type="radio"
              bind:group={barPosition}
              name="left"
              value={"left"}
            />
            left
          </label>
          <label>
            <input
              type="radio"
              bind:group={barPosition}
              name="behind"
              value={"behind"}
            />
            behind numbers
          </label>

          <label>
            <input
              type="radio"
              bind:group={barPosition}
              name="right"
              value={"right"}
            />
            right
          </label>
        </form>
      </div>
      <div class="option-box">
        bar container width
        <input type="range" min="10" max="100" bind:value={barContainerWidth} />
        {barContainerWidth}px
      </div>

      <div class="option-box">
        bar offset (if left or right)
        <input type="range" min="0" max="100" bind:value={barOffset} />
        {barOffset}px
      </div>

      <div class="option-box">
        <ColorPicker bind:hex={negativeColor} label="negative bar color" />
        <ColorPicker bind:hex={positiveColor} label="positive bar color" />

        <ColorPicker
          bind:hex={barBackgroundColor}
          label="bar background color"
        />
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
        </div>
      </div>
    </div>
  </div>

  <div style="padding-left: 40px;">
    table options
    <div class="option-box">
      table gutter width
      <input type="range" min="10" max="100" bind:value={tableGutterWidth} />
      {tableGutterWidth}px
    </div>
  </div>
</div>

{#if selectedFormatterForSamples !== undefined}
  <div class="table-container">
    <table class="ui-copy-number fixed-width-cols">
      <thead>
        {#each numberLists as { desc, sample }, _i}
          <td
            style="padding-left: {tableGutterWidth}px; width: {layerContainerWidth}px; min-width: {layerContainerWidth}px; padding-bottom: 0px;"
          >
            <div class="column-title">{desc}</div></td
          >
        {/each}
      </thead>
      {#each numberLists[0].sample as _, i}
        <tr>
          {#each numberLists as { desc, sample }, j}
            {@const richNum = selectedFormatterForSamples[desc](sample[i])}

            <td
              style="padding-left: {tableGutterWidth}px; width: {layerContainerWidth}px; min-width: {layerContainerWidth}px;"
              class="table-body"
              title={sample[i].toString()}
            >
              <div class="align-content-right">
                <LayeredContainer
                  containerWidth={layerContainerWidth}
                  {barPosition}
                  {barOffset}
                >
                  <AlignedNumber
                    slot="foreground"
                    containerWidth={79 + suffixPadding}
                    {richNum}
                    alignSuffix={alignSuffixes}
                    {alignDecimalPoints}
                    {lowerCaseEForEng}
                    {zeroHandling}
                    {suffixPadding}
                    {showMagSuffixForZero}
                  />
                  <div slot="background" style="width: {barContainerWidth}px;">
                    {#if showBars}
                      <RichNumberBipolarBar
                        containerWidth={barContainerWidth}
                        {richNum}
                        {positiveColor}
                        {negativeColor}
                        {showBaseline}
                        {baselineColor}
                        {barBackgroundColor}
                        {absoluteValExtentsIfPosAndNeg}
                        {absoluteValExtentsAlways}
                      />
                    {/if}
                  </div>
                </LayeredContainer>
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
    /* padding-left: 20px; */
    /* padding-bottom: 3px; */
    vertical-align: bottom;

    /* border-bottom: 1px solid rgb(210, 208, 208); */
  }

  thead td div.column-title {
    /* text-align: right; */
    /* padding-left: 20px; */
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

  /* table.fixed-width-cols td {
    width: 120px;
    min-width: 120px;
  } */

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
