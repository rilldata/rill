<script lang="ts">
  export let excludeMode;

  $: console.log("TestPropsReciever -- update excludeMode to:", excludeMode);
</script>

<div class="flex gap-x-2">
  excludeMode {excludeMode}
</div>
