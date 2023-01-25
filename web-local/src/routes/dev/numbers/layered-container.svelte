<script lang="ts">
  export let containerWidth;

  export let barPosition: "left" | "behind" | "right";
  export let barOffset: number;
</script>

{#if barPosition === "behind"}
  <div class="layers-container" style="width: {containerWidth}px;">
    <div class="background-container"><slot name="background" /></div>
    <div class="foreground-container"><slot name="foreground" /></div>
  </div>
{:else if barPosition === "left"}
  <div class="side-by-side-container">
    <div class="side-by-side-child"><slot name="background" /></div>
    <div class="spacer" style="width: {barOffset}px;" />

    <div class="side-by-side-child"><slot name="foreground" /></div>
  </div>
{:else if barPosition === "right"}
  <div class="side-by-side-container">
    <div class="side-by-side-child">
      <slot name="foreground" style="width: fit-content;" />
    </div>
    <div class="spacer" style="width: {barOffset}px;" />

    <div class="side-by-side-child"><slot name="background" /></div>
  </div>
{/if}

<style>
  div.layers-container {
    position: relative;
  }
  div.background-container {
    display: block;
    position: absolute;
    width: 100%;
    height: 100%;
    /* background-color: rgba(255, 0, 225, 0.29); */
    top: 0px;
    right: 0px;
  }

  div.foreground-container {
    /* display: flex;
    flex-direction: row;
    justify-content: flex-end;
    flex-wrap: nowrap;
    white-space: nowrap;
    overflow: hidden; */
    position: relative;
    /* z-index: 10; */
    /* outline: 1px solid black; */
  }

  div.side-by-side-container {
    position: relative;
    display: flex;
    flex-direction: row;
    width: fit-content;
  }
  div.side-by-side-child {
    position: relative;
    width: fit-content;
  }
</style>
