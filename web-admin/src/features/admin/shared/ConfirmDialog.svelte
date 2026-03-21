<!-- web-admin/src/features/admin/shared/ConfirmDialog.svelte -->
<script lang="ts">
  export let open = false;
  export let title: string;
  export let description: string = "";
  export let confirmLabel: string = "Confirm";
  export let cancelLabel: string = "Cancel";
  export let destructive: boolean = false;

  let loading = false;

  async function handleConfirm() {
    loading = true;
    try {
      await onConfirm();
      open = false;
    } catch {
      // Error handling is the caller's responsibility (via notifyError in onConfirm).
      // Keep dialog open so the user can retry or cancel.
    } finally {
      loading = false;
    }
  }

  function handleCancel() {
    open = false;
  }

  export let onConfirm: () => void | Promise<void>;
</script>

{#if open}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
    on:click={handleCancel}
  >
    <div
      class="bg-white dark:bg-slate-800 rounded-lg p-6 max-w-md w-full mx-4 shadow-xl"
      on:click|stopPropagation
    >
      <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">
        {title}
      </h2>
      {#if description}
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-2">
          {description}
        </p>
      {/if}
      <div class="flex justify-end gap-3 mt-6">
        <button
          class="px-4 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 disabled:opacity-50 disabled:cursor-not-allowed"
          on:click={handleCancel}
          disabled={loading}
        >
          {cancelLabel}
        </button>
        <button
          class="px-4 py-2 text-sm rounded-md text-white disabled:opacity-50 disabled:cursor-not-allowed {destructive
            ? 'bg-red-600 hover:bg-red-700'
            : 'bg-blue-600 hover:bg-blue-700'}"
          on:click={handleConfirm}
          disabled={loading}
        >
          {#if loading}Working...{:else}{confirmLabel}{/if}
        </button>
      </div>
    </div>
  </div>
{/if}
