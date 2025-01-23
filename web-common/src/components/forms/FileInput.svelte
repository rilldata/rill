<script lang="ts">
  import Viz from "@rilldata/web-common/components/icons/Viz.svelte";

  export let value: string | string[] | undefined = undefined;
  export let multiple: boolean = false;
  export let onInput: (files: File[]) => void;

  $: values = value ? (multiple ? value : [value]) : [];
  let fileInput: HTMLInputElement;

  function handleInput() {
    onInput(Array.from(fileInput.files));
  }

  $: console.log(values);
</script>

<button class="upload-button" on:click={() => fileInput.click()}>
  <Viz size="28px" class="text-gray-400" />
  <div class="flex flex-col">
    <span class="upload-title">Upload an image</span>
    {#if multiple}
      <span class="upload-description">
        Support for a single or bulk upload.
      </span>
    {/if}
  </div>
</button>
{#if values}
  {#each values as val (val)}
    <div>{val}</div>
  {/each}
{/if}
<input
  type="file"
  accept="image/*"
  hidden
  multiple
  bind:this={fileInput}
  on:input={handleInput}
/>

<style lang="postcss">
  .upload-button {
    @apply flex flex-row gap-x-2.5 items-center justify-center py-5 min-h-10 w-80;
    @apply border border-neutral-400 fill-neutral-100;
  }

  .upload-title {
    @apply text-sm font-medium text-gray-500;
  }

  .upload-description {
    @apply text-xs font-normal text-gray-400;
  }
</style>
