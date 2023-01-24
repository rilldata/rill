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
  let alignDecimalPoints = true;
  let alignSuffixes = true;
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

  $: {
    selectedFormatterForSamples = Object.fromEntries(
      numberLists.map((nl) => {
        return [nl.desc, selectedFormatter.fn(nl.sample, formatterOptions)];
      })
    );
  }

  let numberInputType;
  let magnitudeStrategy = "largest";
</script>

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

      <!-- <div class="option-box">
        <label>
          <input
            type="radio"
            bind:group={magnitudeStrategy}
            name="outliers"
            value={"outliers"}
          />
          allow a limited number of magnitudes (for outliers) NOT YET IMPLEMENTED
        </label>
      </div> -->
    </form>

    <!-- <div class="option-box">
      <label>
        <input type="checkbox" bind:checked={onlyUseLargestMagnitude} />
        only use largest magnitude
      </label>
    </div> -->
  </div>
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
              {zeroHandling}
            />
          </div>
        </td>
      {/each}
    </tr>
  {/each}
</table>

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
</style>
