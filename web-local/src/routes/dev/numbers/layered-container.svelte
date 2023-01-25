<script lang="ts">
  export let containerWidth;
</script>

<div class="layers-container" style="width: {containerWidth}px;">
  <div class="background-container"><slot name="background" /></div>
  <div class="foreground-container"><slot name="foreground" /></div>
</div>

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
    left: 0px;
  }

  div.foreground-container {
    display: flex;
    flex-direction: row;
    justify-content: flex-end;
    flex-wrap: nowrap;
    white-space: nowrap;
    overflow: hidden;
    position: relative;
    /* z-index: 10; */
    /* outline: 1px solid black; */
  }
</style>
