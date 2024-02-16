<script lang="ts">
  export let onCancel: () => void;

  function cancelGeneration() {
    console.log("Cancel AI generation");
    onCancel();
  }
</script>

<div class="flex flex-col gap-y-1">
  <span> This could take up to 20 seconds. </span>
  <span>
    Prefer not to wait?
    <button on:click={cancelGeneration} class="font-semibold text-primary-400"
      >Start simple</button
    >
  </span>
</div>
