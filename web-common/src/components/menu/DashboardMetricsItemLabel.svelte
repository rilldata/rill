<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { PencilIcon } from "lucide-svelte";

  type Props = {
    displayName: string;
    // Marks the row as an ephemeral measure with an fx marker
    // and, when `onEdit` is set, an edit button.
    ephemeral?: boolean;
    onEdit?: (e: Event) => void;
    // Extra classes for the label span, e.g. a text color.
    class?: string;
    // Classes for the edit button, shared with the row's other buttons.
    buttonClass?: string;
  };

  let {
    displayName,
    ephemeral = false,
    onEdit = undefined,
    class: className = "",
    buttonClass = "",
  }: Props = $props();
</script>

<span class="truncate min-w-0 flex-1 text-left pointer-events-none {className}">
  {displayName}
  {#if ephemeral}
    <span class="text-[10px] font-semibold italic">ƒx</span>
  {/if}
</span>
{#if ephemeral && onEdit}
  <button
    class="{buttonClass} ml-auto"
    onclick={onEdit}
    onmousedown={(e) => e.stopPropagation()}
    aria-label={m.dashboard_pivot_ephemeral_edit_title()}
    type="button"
  >
    <PencilIcon size="14px" />
  </button>
{/if}
