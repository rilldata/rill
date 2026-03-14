<script lang="ts">
  import InputLabel from "./InputLabel.svelte";

  export let value: string = "";
  export let id: string = "";
  export let label: string = "";
  export let placeholder: string = "";
  export let optional: boolean = false;
  export let hint: string | undefined = undefined;
  export let rows: number = 3;
  export let disabled: boolean = false;
  export let errors: string | string[] | null | undefined = null;

  export { className as class };
  let className: string = "";
</script>

<div class="textarea-wrapper">
  {#if label}
    <InputLabel {id} {label} {optional} {hint} />
  {/if}

  <textarea
    {id}
    name={id}
    class="textarea {className}"
    class:error={!!errors?.length}
    bind:value
    {placeholder}
    {rows}
    {disabled}
  />

  {#if errors}
    {#if typeof errors === "string"}
      <div class="error-text">{errors}</div>
    {:else}
      {#each errors as error (error)}
        <div class="error-text">{error}</div>
      {/each}
    {/if}
  {/if}
</div>

<style lang="postcss">
  .textarea-wrapper {
    @apply flex flex-col gap-y-1;
  }

  .textarea {
    @apply w-full px-3 py-2;
    @apply border rounded-[2px];
    @apply text-sm;
    @apply resize-none;
    @apply bg-surface;
  }

  .textarea:focus {
    @apply outline-none;
    @apply border-primary-500;
    @apply ring-2 ring-primary-100;
  }

  .textarea:disabled {
    @apply bg-surface-muted text-fg-disabled cursor-not-allowed;
  }

  .textarea::placeholder {
    @apply text-fg-muted;
  }

  .textarea.error {
    @apply border-destructive;
  }

  .error-text {
    @apply text-destructive text-xs;
  }
</style>
