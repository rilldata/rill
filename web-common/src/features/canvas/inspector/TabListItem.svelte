<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { ArrowDown, ArrowUp, Copy, Trash2 } from "lucide-svelte";

  export let displayName: string;
  export let name: string;
  export let active = false;
  export let canMoveUp: boolean;
  export let canMoveDown: boolean;
  export let onCommitLabel: (value: string) => void;
  // Live (per-keystroke) display-name update, so the tab strip reflects typing immediately.
  export let onInputLabel: (value: string) => void;
  export let onCommitName: (value: string) => void;
  export let onMoveUp: () => void;
  export let onMoveDown: () => void;
  export let onDuplicate: () => void;
  export let onDelete: () => void;

  // Local drafts synced from props only when the committed value changes (not while typing),
  // so a reconcile elsewhere can't clobber an in-progress edit in this row.
  let label = displayName;
  let lastDisplayName = displayName;
  $: if (displayName !== lastDisplayName) {
    label = displayName;
    lastDisplayName = displayName;
  }

  let nameValue = name;
  let lastName = name;
  $: if (name !== lastName) {
    nameValue = name;
    lastName = name;
  }
</script>

<li class="tab-row" class:active>
  <div class="fields">
    <Input
      capitalizeLabel={false}
      textClass="text-sm"
      size="sm"
      labelGap={2}
      label="Name"
      hint="Stable identifier used as the tab's deep-link URL key"
      placeholder="Defaults to a slug of the display name"
      bind:value={nameValue}
      onBlur={() => onCommitName(nameValue)}
      onEnter={() => onCommitName(nameValue)}
    />
    <Input
      capitalizeLabel={false}
      textClass="text-sm"
      size="sm"
      labelGap={2}
      label="Display name"
      placeholder="Shown on the tab"
      bind:value={label}
      onInput={(value) => onInputLabel(value)}
      onBlur={() => onCommitLabel(label)}
      onEnter={() => onCommitLabel(label)}
    />
  </div>

  <div class="actions">
    <button title="Move up" disabled={!canMoveUp} on:click={onMoveUp}>
      <ArrowUp size="14px" />
    </button>
    <button title="Move down" disabled={!canMoveDown} on:click={onMoveDown}>
      <ArrowDown size="14px" />
    </button>
    <button title="Duplicate tab" on:click={onDuplicate}>
      <Copy size="14px" />
    </button>
    <button title="Delete tab" on:click={onDelete}>
      <Trash2 size="14px" />
    </button>
  </div>
</li>

<style lang="postcss">
  .tab-row {
    @apply flex items-start gap-x-2 rounded-md border border-gray-200 p-2;
  }

  .tab-row.active {
    @apply border-primary-300 bg-primary-50/40;
  }

  .fields {
    @apply flex min-w-0 flex-1 flex-col gap-y-2;
  }

  .actions {
    @apply flex flex-none items-center gap-x-0.5 pt-5;
  }

  .actions button {
    @apply grid size-6 place-content-center rounded text-fg-secondary;
    @apply hover:bg-surface-subtle hover:text-fg-primary;
    @apply disabled:pointer-events-none disabled:opacity-30;
  }
</style>
