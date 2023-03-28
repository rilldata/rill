<script lang="ts">
  import AlignedNumber from "../aligned-number.svelte";

  import type { NumberFormatter } from "../number-to-string-formatters";
  // import LayeredContainer from "../layered-container.svelte";
  // import RichNumberBipolarBar from "../rich-number-bipolar-bar.svelte";
  import type { FormatterOptionsV1 } from "../formatter-options";

  // FORMATTER SELECTION
  export let formattersDescriptionsAndOptions: [
    formatter: NumberFormatter,
    description: string,
    options: FormatterOptionsV1,
    pxWidth: number
  ][];
  export let sample: number[];
  export let tableGutterWidth: number;

  $: console.log({ formattersDescriptionsAndOptions });
</script>

<div class="table-container">
  <table class="ui-copy-number fixed-width-cols">
    <thead>
      {#each formattersDescriptionsAndOptions as [_formatter, desc, _options, pxWidth], _i}
        <td
          style="padding-left: {tableGutterWidth}px; width: {pxWidth}px; min-width: {pxWidth}px; padding-bottom: 0px;"
        >
          <div class="column-title">{desc}</div></td
        >
      {/each}
    </thead>
    {#each sample as x, i}
      <tr>
        {#each formattersDescriptionsAndOptions as [formatter, _desc, options, pxWidth], _i}
          {@const richNum = formatter(x)}
          {@const {
            alignSuffixes,
            alignDecimalPoints,
            lowerCaseEForEng,
            zeroHandling,
            suffixPadding,
            showMagSuffixForZero,
          } = options}

          <td
            style="padding-left: {tableGutterWidth}px; width: {pxWidth}px; min-width: {pxWidth}px;"
            class="table-body"
            title={sample[i].toString()}
          >
            <div class="align-content-right">
              <!-- <LayeredContainer
                containerWidth={layerContainerWidth}
                {barPosition}
                barOffset={showBars ? barOffset : 0}
              > -->
              <!-- FIXME: if bars are added back,AlignedNumber will need slot="foreground" -->
              <AlignedNumber
                containerWidth={pxWidth}
                {richNum}
                alignSuffix={alignSuffixes}
                {alignDecimalPoints}
                {lowerCaseEForEng}
                {zeroHandling}
                {suffixPadding}
                {showMagSuffixForZero}
              />
              <!-- <div
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
                </div> -->
              <!-- </LayeredContainer> -->
            </div>
          </td>
        {/each}
      </tr>
    {/each}
  </table>
</div>

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
