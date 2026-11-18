<!-- @component
A measure chip of the expression editor (see MeasureExpressionInput), mounted
as the view of a measure node: the measure's display name, and a button that
removes the chip.
-->
<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { X } from "lucide-svelte";

  let {
    name,
    label,
    onRemove,
  }: { name: string; label: string; onRemove: () => void } = $props();
</script>

<span class="measure-chip" title={name}>
  {label}
  <!-- preventDefault on mousedown keeps the editor's focus and selection. -->
  <button
    type="button"
    tabindex="-1"
    class="remove"
    aria-label={m.dashboard_pivot_ephemeral_remove_measure({ name: label })}
    onmousedown={(e) => e.preventDefault()}
    onclick={onRemove}
  >
    <X size="10px" strokeWidth={2.5} />
  </button>
</span>

<style lang="postcss">
  .measure-chip {
    @apply inline-flex items-center gap-x-0.5 pl-1 pr-0.5 rounded-sm border;
    @apply bg-secondary-50 border-secondary-200 text-secondary-800;
  }

  .remove {
    @apply flex items-center justify-center size-3.5 rounded-sm;
    @apply text-secondary-600 cursor-pointer;
  }

  .remove:hover {
    @apply bg-secondary-200 text-secondary-900;
  }
</style>
