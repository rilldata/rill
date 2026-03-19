<!-- web-admin/src/features/admin/shared/ConfirmDialog.svelte -->
<script lang="ts">
  export let open = false;
  export let title: string;
  export let description: string = "";
  export let confirmLabel: string = "Confirm";
  export let cancelLabel: string = "Cancel";
  export let destructive: boolean = false;
  export let onConfirm: () => void | Promise<void>;

  let loading = false;

  async function handleConfirm() {
    loading = true;
    try {
      await onConfirm();
      open = false;
    } finally {
      loading = false;
    }
  }

  function handleCancel() {
    open = false;
  }
</script>

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="overlay" on:click={handleCancel}>
    <div class="dialog" on:click|stopPropagation>
      <h2 class="dialog-title">{title}</h2>
      {#if description}
        <p class="dialog-description">{description}</p>
      {/if}
      <div class="dialog-actions">
        <button class="btn-cancel" on:click={handleCancel} disabled={loading}>
          {cancelLabel}
        </button>
        <button
          class="btn-confirm"
          class:destructive
          on:click={handleConfirm}
          disabled={loading}
        >
          {#if loading}Working...{:else}{confirmLabel}{/if}
        </button>
      </div>
    </div>
  </div>
{/if}

<style lang="postcss">
  .overlay {
    @apply fixed inset-0 bg-black/50 flex items-center justify-center z-50;
  }

  .dialog {
    @apply bg-white dark:bg-slate-800 rounded-lg p-6 max-w-md w-full mx-4 shadow-xl;
  }

  .dialog-title {
    @apply text-lg font-semibold text-slate-900 dark:text-slate-100;
  }

  .dialog-description {
    @apply text-sm text-slate-500 dark:text-slate-400 mt-2;
  }

  .dialog-actions {
    @apply flex justify-end gap-3 mt-6;
  }

  .btn-cancel {
    @apply px-4 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600
      text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700;
  }

  .btn-confirm {
    @apply px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700;
  }

  .btn-confirm.destructive {
    @apply bg-red-600 hover:bg-red-700;
  }

  button:disabled {
    @apply opacity-50 cursor-not-allowed;
  }
</style>
