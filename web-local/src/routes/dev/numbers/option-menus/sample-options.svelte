<script lang="ts">
  export let samplePreprocessing: "none" | "round" | "currencyRoundCent" =
    "none";
  export let sortSamples: "none" | "asc" | "desc" | "abs_asc" | "abs_desc" =
    "none";
</script>

<h2>Sample preprocessing</h2>
<div style="padding-left: 10px;">
  <form>
    <label>
      <input
        type="radio"
        bind:group={samplePreprocessing}
        name="none"
        value={"none"}
      />
      none (samples as described)
    </label>
    <br />

    <label>
      <input
        type="radio"
        bind:group={samplePreprocessing}
        name="round"
        value={"round"}
      />
      round to ints
    </label>
    <br />
    <label>
      <input
        type="radio"
        bind:group={samplePreprocessing}
        name="currencyRoundCent"
        value={"currencyRoundCent"}
      />
      round to 2 decimal places (like currency)
    </label>
  </form>

  <div style="display: flex; flex-direction: row;">
    <b>Sort</b>

    &nbsp; &nbsp;
    <br />

    <form>
      <label>
        <input
          type="radio"
          bind:group={sortSamples}
          name="none"
          value={"none"}
        />
        none
      </label>
      <label>
        <input type="radio" bind:group={sortSamples} name="asc" value={"asc"} />
        asc
      </label>
      <label>
        <input
          type="radio"
          bind:group={sortSamples}
          name="desc"
          value={"desc"}
        />
        desc
      </label>
      <br />
      <label>
        <input
          type="radio"
          bind:group={sortSamples}
          name="abs_asc"
          value={"abs_asc"}
        />
        abs asc
      </label>
      <label>
        <input
          type="radio"
          bind:group={sortSamples}
          name="abs_desc"
          value={"abs_desc"}
        />
        abs desc
      </label>
    </form>
  </div>
</div>
