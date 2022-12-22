<script>
  import Switch from "@rilldata/web-common/components/button/Switch.svelte";

  export let leftPane = false;
</script>

<h1>Switch Component</h1>

<div class="flex items-center gap-x-2">
  <Switch on:click={() => (leftPane = !leftPane)} checked={leftPane}>
    Left Pane
  </Switch>
  {leftPane}
</div>
