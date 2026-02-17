<script lang="ts">
  import { goto } from "$app/navigation";
  import { previewModeStore } from "./preview-mode-store";

  $: isPreview = $previewModeStore;

  function handleClick() {
    const newPreviewMode = !$previewModeStore;
    previewModeStore.set(newPreviewMode);

    if (newPreviewMode) {
      goto("/home");
    } else {
      goto("/");
    }
  }
</script>

<button class="mode-switch" on:click={handleClick}>
  {#if isPreview}
    <span class="label">Preview</span>
  {:else}
    <span class="label">Developer</span>
  {/if}
</button>

<style lang="postcss">
  .mode-switch {
    @apply flex items-center gap-x-1.5 px-2 py-1 rounded-md cursor-pointer;
    @apply text-xs font-medium;
    color: var(--fg-secondary);
    transition: background 0.15s ease, color 0.15s ease;
  }
  .label {
    display: inline;
  }

</style>
