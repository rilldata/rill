<script lang="ts">
  import AlignedNumber from "./aligned-number.svelte";
  import ColorPicker from "svelte-awesome-color-picker";
  import { numberLists as numberListsUnprocessed } from "./number-samples";
  import {
    formatterFactories,
    NumberFormatter,
    NumPartPxWidthLookupFn,
    RichFormatNumber,
  } from "./number-to-string-formatters";
  import { onMount } from "svelte";
  import LayeredContainer from "./layered-container.svelte";
  import RichNumberBipolarBar from "./rich-number-bipolar-bar.svelte";
  import SampleOptions from "./option-menus/sample-options.svelte";

  // FORMATTER SELECTION
  export let defaultFormatterIndex = 1;
  let selectedFormatter = formatterFactories[defaultFormatterIndex];
  let selectedFormatterForSamples: { [colName: string]: NumberFormatter };

  // SHARED DISPLAY OPTIONS
  export let alignDecimalPoints = true;
  export let alignSuffixes = true;
  let suffixPadding = 2;
  let lowerCaseEForEng = true;
  // let minimumSignificantDigits = 3;
  // let maximumSignificantDigits = 5;
  export let zeroHandling: "exactZero" | "noSpecial" | "zeroDot" = "exactZero";
  export let showMagSuffixForZero = false;

  // NEW HUMANIZER OPTIONS
  let onlyUseLargestMagnitude = false;
  let usePlainNumsForThousands = true;
  let usePlainNumsForThousandsOneDecimal = false;
  let usePlainNumForThousandths = true;
  let usePlainNumForThousandthsPadZeros = false;

  let truncateThousandths = true;
  let truncateTinyOrdersIfBigOrderExists = true;

  let maxTotalDigits = 6;
  let maxDigitsLeft = 3;
  let maxDigitsRight = 3;
  let minDigitsNonzero = 1;
  let nonIntegerHandling: "none" | "oneDigit" | "trailingDot" = "trailingDot";

  // BARS OPTIONS
  let showBars = false;
  let absoluteValExtentsIfPosAndNeg = true;
  let absoluteValExtentsAlways = false;
  let barPosition: "left" | "behind" | "right" = "right";
  let barContainerWidth = 30;
  let barOffset = 10;

  let negativeColor = "#c5deff";
  let positiveColor = "#c5deff";
  let barBackgroundColor = "#f8f8f8";

  let showBaseline = true;
  let baselineColor = "#e2e2e2";

  // TABLE FORMAT OPTIONS
  let tableGutterWidth = 30;

  let worstCaseStringWidth = 79 + suffixPadding;
  $: {
    if (pxWidthLookupFn !== undefined) {
      if (magnitudeStrategy === "unlimitedDigitTarget") {
        let int = pxWidthLookupFn("0") * maxDigitsRight;
        let dot = pxWidthLookupFn(".");
        let frac = pxWidthLookupFn("0") * maxDigitsLeft;
        let suffix = pxWidthLookupFn("e-200");
        worstCaseStringWidth = int + dot + frac + suffix + suffixPadding;
      } else if (magnitudeStrategy === "largestWithDigitTarget") {
        let int = pxWidthLookupFn("0") * 3;
        let dot = pxWidthLookupFn(".");
        let frac = pxWidthLookupFn("0") * (digitTarget - 3);
        let suffix = pxWidthLookupFn("e-200");
        console.log({ int, dot, frac, suffix });
        worstCaseStringWidth = int + dot + frac + suffix + suffixPadding;
      }
    } else {
      worstCaseStringWidth = 79 + suffixPadding;
    }
  }

  $: layerContainerWidth =
    worstCaseStringWidth +
    (showBars && barPosition != "behind" ? barContainerWidth + barOffset : 0);

  $: usePlainNumForThousandthsPadZeros =
    usePlainNumForThousandths && usePlainNumForThousandthsPadZeros;

  $: formatterOptions = {
    // minimumSignificantDigits,
    // maximumSignificantDigits,
    magnitudeStrategy,
    digitTarget,
    digitTargetPadWithInsignificantZeros,
    usePlainNumsForThousands,
    usePlainNumsForThousandsOneDecimal,
    usePlainNumForThousandths,
    usePlainNumForThousandthsPadZeros,
    truncateThousandths,
    truncateTinyOrdersIfBigOrderExists,
    zeroHandling,
    maxTotalDigits,
    maxDigitsLeft,
    maxDigitsRight,
    minDigitsNonzero,
    nonIntegerHandling,
  };

  let samplePreprocessing: "none" | "round" | "currencyRoundCent" = "none";
  let sortSamples: "none" | "asc" | "desc" = "none";

  let magnitudeStrategy:
    | "unlimited"
    | "unlimitedDigitTarget"
    | "largestWithDigitTarget" = "unlimitedDigitTarget";
  let digitTargetPadWithInsignificantZeros = false;
  let digitTarget = 5;

  const blue100 = "#dbeafe";
  const grey100 = "#f5f5f5";

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
    "e",
    "E",
  ];
  for (let i = 0; i <= 9; i++) {
    numFormattingWidthLookupKeys.push(i + "");
  }

  let numFormattingWidthLookup: { [key: string]: number } = {};

  let charMeasuringDiv: HTMLDivElement;

  let pxWidthLookupFn: NumPartPxWidthLookupFn;

  onMount(() => {
    console.time("charMeasuringDiv");
    numFormattingWidthLookupKeys.forEach((str) => {
      charMeasuringDiv.innerHTML = str;
      let rect = charMeasuringDiv.getBoundingClientRect();
      numFormattingWidthLookup[str] = rect.right - rect.left;
    });

    console.timeEnd("charMeasuringDiv");

    pxWidthLookupFn = (str: string) => {
      return str
        .split("")
        .map((char) => numFormattingWidthLookup[char])
        .reduce((a, b) => a + b, 0);
    };
  });

  $: numberLists = numberListsUnprocessed.map((nl) => {
    let sample = nl.sample.map((x) => {
      switch (samplePreprocessing) {
        case "currencyRoundCent":
          return Math.round(x * 100) / 100;
        case "round":
          return Math.round(x);
        default:
          return x;
      }
    });

    sample =
      sortSamples === "none"
        ? sample
        : sample.sort((a, b) => (sortSamples === "desc" ? b - a : a - b));

    return {
      sample,
      desc: nl.desc,
    };
  });

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
    <SampleOptions bind:samplePreprocessing bind:sortSamples />

    <h2>Layout options (applies to all formatters)</h2>

    <div>
      <label>
        <input type="checkbox" bind:checked={alignDecimalPoints} />
        align decimal points
      </label>
    </div>
    <div>
      <label>
        <input type="checkbox" bind:checked={alignSuffixes} />
        align suffixes
      </label>
      <div>
        <label>
          suffix padding:
          <input
            class="number-input"
            type="number"
            min="0"
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

    <h3>Zero handling</h3>
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
  </div>

  <div style="padding-left: 40px;">
    <h2>new humanizer shared options</h2>

    <label>
      <input
        type="checkbox"
        bind:checked={digitTargetPadWithInsignificantZeros}
      />
      pad with insignificant zeros (after last significant digit)
    </label>

    <h2>new humanizer strategy (and strategy-specific options)</h2>
    <form>
      <div class="option-box">
        <label>
          <input
            type="radio"
            bind:group={magnitudeStrategy}
            name="largestWithDigitTarget"
            value={"largestWithDigitTarget"}
          />
          <b>only use largest magnitude</b>
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
          <b>multiple magnitudes v1</b>
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

      <div class="option-box">
        <label>
          <input
            type="radio"
            bind:group={magnitudeStrategy}
            name="unlimitedDigitTarget"
            value={"unlimitedDigitTarget"}
          />
          <b>multiple magnitudes v2 (digit targets) </b></label
        >

        <div class="option-box">
          <label>
            max total digits
            <input
              class="number-input"
              type="number"
              min="3"
              max="12"
              bind:value={maxTotalDigits}
              on:change={() => {
                if (maxDigitsLeft >= maxTotalDigits) {
                  maxDigitsLeft = maxTotalDigits;
                }
                if (maxDigitsRight >= maxTotalDigits) {
                  maxDigitsRight = maxTotalDigits;
                }
              }}
            />
          </label>
          <br />

          <label>
            max digits left of decimal point
            <input
              class="number-input"
              type="number"
              min="3"
              max="12"
              bind:value={maxDigitsLeft}
              on:change={() => {
                if (maxDigitsLeft >= maxTotalDigits) {
                  maxTotalDigits = maxDigitsLeft;
                }
              }}
            />
          </label>
          <br />

          <label>
            max digits right of decimal point
            <input
              class="number-input"
              type="number"
              min="0"
              max="12"
              bind:value={maxDigitsRight}
              on:change={() => {
                if (maxDigitsRight >= maxTotalDigits) {
                  maxTotalDigits = maxDigitsRight;
                }
                if (maxDigitsRight <= minDigitsNonzero) {
                  minDigitsNonzero = maxDigitsRight;
                }
              }}
            />
          </label>
          <br />

          <label>
            min non-zero digits for fractional vals
            <input
              class="number-input"
              type="number"
              min="0"
              max={maxDigitsRight}
              bind:value={minDigitsNonzero}
            />
          </label>

          <br />

          <b>Presets</b>
          <div class="option-box">
            <button
              title="better for alignment; truncate small nums entirely"
              on:click|preventDefault={() => {
                maxTotalDigits = 6;
                maxDigitsLeft = 3;
                maxDigitsRight = 3;
                minDigitsNonzero = 0;
              }}>1</button
            >
            &nbsp;
            <button
              title="better for alignment; truncate small nums mostly"
              on:click|preventDefault={() => {
                maxTotalDigits = 6;
                maxDigitsLeft = 3;
                maxDigitsRight = 3;
                minDigitsNonzero = 1;
              }}>2</button
            >
            &nbsp;
            <button
              title="not optimal for alignment; truncate small nums entirely"
              on:click|preventDefault={() => {
                maxTotalDigits = 6;
                maxDigitsLeft = 6;
                maxDigitsRight = 5;
                minDigitsNonzero = 0;
              }}>3</button
            >
            &nbsp;
            <button
              title="not optimal for alignment; truncate small nums mostly"
              on:click|preventDefault={() => {
                maxTotalDigits = 6;
                maxDigitsLeft = 6;
                maxDigitsRight = 5;
                minDigitsNonzero = 1;
              }}>4</button
            >
            &nbsp;
          </div>

          <b>handling of non-ints that truncate to the e0 digit</b>
          <div class="option-box">
            <form>
              <label>
                <input
                  type="radio"
                  bind:group={nonIntegerHandling}
                  name="none"
                  value={"none"}
                />
                truncate without trailing "." <br /> ex: 1403.35 -> "1403"
              </label>
              <br />

              <label>
                <input
                  type="radio"
                  bind:group={nonIntegerHandling}
                  name="trailingDot"
                  value={"trailingDot"}
                />
                leave a trailing "." <br /> ex: 1403.35 -> "1403."
              </label>
              <br />

              <label>
                <input
                  type="radio"
                  bind:group={nonIntegerHandling}
                  name="oneDigit"
                  value={"oneDigit"}
                />
                roll over to next magnitude <br /> ex: in 4 digit budget, "1403.35"
                -> "1.403 k"
              </label>
            </form>
          </div>
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
        <input type="range" min="10" max="300" bind:value={barContainerWidth} />
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
                  barOffset={showBars ? barOffset : 0}
                >
                  <AlignedNumber
                    slot="foreground"
                    containerWidth={worstCaseStringWidth}
                    {richNum}
                    alignSuffix={alignSuffixes}
                    {alignDecimalPoints}
                    {lowerCaseEForEng}
                    {zeroHandling}
                    {suffixPadding}
                    {showMagSuffixForZero}
                  />
                  <div
                    slot="background"
                    style="width: {showBars ? barContainerWidth : 0}px;"
                  >
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
    padding-left: 15px;
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
