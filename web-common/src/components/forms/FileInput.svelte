<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size.ts";

  export let files: FileList | undefined;
  export let error: string | Record<string | number, string[]> | undefined =
    undefined;
  export let multiple: boolean = false;
  export let accept: string | undefined = ".json";
  export let id: string | undefined = undefined;
  export let label: string | undefined = undefined;
  export let hint: string | undefined = undefined;
  export let optional: boolean = false;
  export let hidden: boolean = false;

  let fileInput: HTMLInputElement;

  $: errors = error ? (multiple ? error : { 0: error }) : [];
  $: errorMessages = Object.values({
    ...(errors as Record<string, any>),
  })
    .map((e, i) => (files?.[i] && e ? `${files[i].name}:${e}` : ""))
    .filter(Boolean);

  $: selectedFile = files?.[0];
</script>

<div class="container">
  {#if !hidden}
    {#if label && id}
      <InputLabel {id} {label} {hint} {optional} />
    {/if}

    <div class="file-input-wrapper">
      <button
        type="button"
        class="file-input-button"
        onclick={() => fileInput.click()}
        aria-label="Choose file"
      >
        <span class="choose-file-text">Choose file</span>
        <span class="file-status-text">
          {#if selectedFile}
            {@const formattedSize = formatMemorySize(
              selectedFile.size ? Number(selectedFile.size) : 0,
            )}
            {selectedFile.name} ({formattedSize})
          {:else}
            No file chosen
          {/if}
        </span>
      </button>
    </div>
    {#if errorMessages.length > 0}
      <div class="error">
        {#each errorMessages as errorMessage, i (i)}
          <div>{errorMessage}</div>
        {/each}
      </div>
    {/if}
    <input
      type="file"
      bind:files
      {accept}
      hidden
      {multiple}
      bind:this={fileInput}
    />
  {/if}
</div>

<style lang="postcss">
  .container {
    @apply flex flex-col gap-y-2;
  }

  .file-input-wrapper {
    @apply w-full relative bg-input border rounded-sm;
  }

  .file-input-button {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 6px;
    padding: 0.5rem 0.75rem;
    border-radius: 0.125rem;
    text-align: left;
    cursor: pointer;
    transition: border-color 0.2s;
  }

  .file-input-button:focus {
    outline: none;
    border-color: #3b82f6;
  }

  .choose-file-text {
    @apply font-medium text-fg-primary text-sm;
  }

  .file-status-text {
    @apply text-fg-secondary text-sm;
  }

  .error {
    @apply text-red-600 text-xs py-px mt-0.5;
  }
</style>
