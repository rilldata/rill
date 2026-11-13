<!-- @component
The "New measure" action of the menus and lists that offer creating an
ephemeral measure, placed next to their search input. At the limit of
definitions it is disabled and says why: the "button" variant renders that note
as a full-width sibling, so place it in a wrapping flex row (`flex-wrap`); the
"icon" variant puts it in the tooltip.
-->
<script lang="ts">
  import AlertTriangle from "@rilldata/web-common/components/icons/AlertTriangle.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ephemeralMeasureDialog } from "./dialog-store";
  import { MAX_EPHEMERAL_MEASURES } from "./validation";

  // The number of definitions that already exist where this creates one.
  export let count: number;
  // "icon" is a bare "+" for tight headers such as the pivot sidebar's.
  export let variant: "button" | "icon" = "button";
  // Optional hook so menus can close themselves before the dialog opens.
  export let onOpen: (() => void) | undefined = undefined;
  // Canvas keeps its definitions on the component spec, so it opens its own
  // editor instead of the dashboard-wide dialog.
  export let onCreate: () => void = () => ephemeralMeasureDialog.set({});

  $: atLimit = count >= MAX_EPHEMERAL_MEASURES;
  $: limitMessage = m.dashboard_pivot_ephemeral_limit_reached({
    count: String(MAX_EPHEMERAL_MEASURES),
  });

  function create() {
    if (atLimit) return;
    onOpen?.();
    onCreate();
  }
</script>

{#if variant === "icon"}
  <!-- aria-disabled instead of disabled keeps the tooltip on hover. -->
  <button
    class="icon-button"
    class:at-limit={atLimit}
    type="button"
    aria-label={m.dashboard_pivot_ephemeral_create()}
    aria-disabled={atLimit}
    title={atLimit ? limitMessage : m.dashboard_pivot_ephemeral_create()}
    on:click={create}
  >
    <Add size="14px" />
  </button>
{:else}
  <!-- A compact text action, like the menus' "Hide all" and "Show all", so
       the search input stays the prominent control of the row. -->
  <button
    class="text-button"
    type="button"
    disabled={atLimit}
    data-testid="create-ephemeral-measure"
    on:click={create}
  >
    <Add size="14px" />
    {m.dashboard_pivot_ephemeral_create()}
  </button>
  {#if atLimit}
    <p class="limit-note" role="status">
      <span class="flex-none mt-px"><AlertTriangle size="14px" /></span>
      {limitMessage}
    </p>
  {/if}
{/if}

<style lang="postcss">
  .icon-button {
    @apply text-fg-secondary;
  }

  .icon-button:hover {
    @apply text-fg-primary;
  }

  .icon-button.at-limit {
    @apply text-fg-disabled cursor-not-allowed;
  }

  .text-button {
    @apply flex flex-none items-center gap-x-1 h-7 px-1.5 rounded-sm;
    @apply text-xs font-medium whitespace-nowrap text-theme-500;
  }

  .text-button:hover:not(:disabled) {
    @apply bg-surface-hover text-theme-600;
  }

  .text-button:disabled {
    @apply text-fg-disabled cursor-not-allowed;
  }

  .limit-note {
    @apply basis-full flex items-start gap-x-1.5 px-2 py-1.5 rounded-sm;
    @apply text-xs bg-yellow-50 text-yellow-800;
    @apply dark:bg-yellow-900/20 dark:text-yellow-300;
  }
</style>
