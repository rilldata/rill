<script lang="ts">
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ephemeralMeasureDialog } from "./dialog-store";

  // Optional hook so menus can close themselves before the dialog opens.
  export let onOpen: (() => void) | undefined = undefined;
  // Canvas keeps its definitions on the component spec, so it opens its own
  // editor instead of the dashboard-wide dialog.
  export let onCreate: () => void = () => ephemeralMeasureDialog.set({});
</script>

<button
  class="create-ephemeral-measure"
  type="button"
  on:click={() => {
    onOpen?.();
    onCreate();
  }}
>
  <Add size="14px" />
  {m.dashboard_pivot_ephemeral_create()}
</button>

<style lang="postcss">
  .create-ephemeral-measure {
    @apply flex w-full items-center gap-x-1.5;
    @apply h-9 px-2 text-xs text-fg-primary;
  }

  .create-ephemeral-measure:hover {
    @apply bg-surface-hover;
  }
</style>
