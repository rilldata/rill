<script>
  import Switch from "@rilldata/web-local/lib/components/button/Switch.svelte";

  export let leftPane = false;
</script>

<h1>Switch Component</h1>

<div class="flex items-center gap-x-2">
  <Switch bind:checked={leftPane} /> left pane {leftPane}
</div>
