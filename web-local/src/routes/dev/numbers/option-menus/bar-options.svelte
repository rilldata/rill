<script lang="ts">
  export let showBars = true;
  export let absoluteValExtentsIfPosAndNeg = true;
  export let absoluteValExtentsAlways = false;
  export let reflectNegativeBars = false;

  export let barPosition: "left" | "behind" | "right" = "right";
  export let barContainerWidth = 30;
  export let barOffset = 10;

  // https://tailwindcss.com/docs/customizing-colors

  const blue100 = "#dbeafe";
  const blue200 = "#bfdbfe";
  // const blue300 = "#93c5fd";

  const red100 = "#fee2e2";
  const red200 = "#fecaca";
  // const red300 = "#fca5a5";

  const white = "#ffffff";

  const grey50 = "#f9fafb";
  const grey75 = "#f8f8f8";
  const grey100 = "#f3f4f6";
  const grey200 = "#e5e7eb";
  const grey300 = "#d1d5db";
  const grey400 = "#9ca3af";
  // const grey500 = "#6b7280";

  export let negativeColor = red200;
  export let positiveColor = blue200;
  export let barBackgroundColor = grey75;

  export let showBaseline = true;
  export let baselineColor = grey200;
</script>

<div style="padding-left: 10px;">
  <h3>bar options</h3>
  <label>
    <input type="checkbox" bind:checked={showBars} />
    show bars
  </label>

  <button
    on:click={() => {
      barPosition = "behind";
      barContainerWidth = 300;
    }}>leaderboard-ish</button
  >
  &nbsp;<button
    on:click={() => {
      barPosition = "right";
      barContainerWidth = 35;
    }}>tabular</button
  >
  <h3>bar position</h3>

  <div class="option-box">
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
  </div>
  <h3>bar direction</h3>
  <div class="option-box">
    <form>
      <label>
        <input
          type="radio"
          bind:group={reflectNegativeBars}
          name="false"
          value={false}
        />
        diverging bars
        <div class="option-box">
          <label>
            <input
              type="checkbox"
              bind:checked={absoluteValExtentsIfPosAndNeg}
            />
            use symmetric extents if sample has pos and neg values
          </label>
          <div class="option-box">
            <label>
              <input type="checkbox" bind:checked={absoluteValExtentsAlways} />
              always use symmetric extents
            </label>
          </div>
        </div>
      </label>

      <label>
        <input
          type="radio"
          bind:group={reflectNegativeBars}
          name="true"
          value={true}
        />
        reflect negative bars
      </label>
    </form>
  </div>

  <h3>bar colors</h3>
  <div class="option-box">
    <!-- <div><ColorPicker bind:hex={negativeColor} label="negative" /></div> -->
    <div class="color-picker-row">
      negative: {negativeColor} &nbsp;
      <div class="color-picker-wrapper">
        <input type="color" bind:value={negativeColor} />
      </div>
      <button on:click={() => (negativeColor = red100)}>r-100</button>
      &nbsp;
      <button on:click={() => (negativeColor = red200)}>r-200</button>
      &nbsp;
      <button on:click={() => (negativeColor = grey200)}>gy-200</button>
    </div>

    <div class="color-picker-row">
      positive: {positiveColor} &nbsp;
      <div class="color-picker-wrapper">
        <input type="color" bind:value={positiveColor} />
      </div>
      <button on:click={() => (positiveColor = blue100)}>b-100</button>
      &nbsp;
      <button on:click={() => (positiveColor = blue200)}>b-200</button>
      &nbsp;
      <button on:click={() => (positiveColor = grey200)}>gy-200</button>
      &nbsp;
      <button on:click={() => (positiveColor = grey300)}>gy-300</button>
    </div>

    <div class="color-picker-row">
      background: {barBackgroundColor} &nbsp;
      <div class="color-picker-wrapper">
        <input type="color" bind:value={barBackgroundColor} />
      </div>
      <button on:click={() => (barBackgroundColor = white)}>w</button>
      &nbsp;
      <button on:click={() => (barBackgroundColor = grey50)}>gy-50</button>
      &nbsp;
      <button on:click={() => (barBackgroundColor = grey75)}>gy-75</button>
      &nbsp;
      <button on:click={() => (barBackgroundColor = grey100)}>gy-100</button>
      &nbsp;
      <button on:click={() => (barBackgroundColor = grey200)}>gy-200</button>
    </div>

    <div class="color-picker-row">
      baseline: {baselineColor}&nbsp;
      <div class="color-picker-wrapper">
        <input type="color" bind:value={baselineColor} />
      </div>
      <button on:click={() => (baselineColor = grey100)}>gy-100</button>
      &nbsp;
      <button on:click={() => (baselineColor = grey200)}>gy-200</button>
      &nbsp;
      <button on:click={() => (baselineColor = grey400)}>gy-400</button>
      &nbsp;
      <label>
        <input type="checkbox" bind:checked={showBaseline} />
        show
      </label>
    </div>
  </div>
</div>

<style>
  .option-box {
    padding-left: 15px;
  }
  .color-picker-row {
    /* width: 100px; */
    display: flex;
    flex-direction: row;

    align-items: center;
  }

  .color-picker-wrapper {
    /* width: 100px; */
    display: inline-block;
  }

  button {
    outline: 1px solid #ddd;
    background-color: #f2f2f2;
    padding: 3px;
    border-radius: 5px;
    margin-left: 5px;
  }
  input[type="color"] {
    width: 30px;
    height: 30px;
    margin-right: 8px;
  }
</style>
