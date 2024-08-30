<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import {
    ArrowUpToLineIcon,
    ArrowDownToLineIcon,
    Pen,
    CopyIcon,
    Trash2Icon,
  } from "lucide-svelte";

  export let onEdit: () => void;
  export let onDelete: () => void;
  export let onDuplicate: () => void;
  export let onMoveToTop: () => void;
  export let onMoveToBottom: () => void;
  export let first = false;
  export let last = false;
  export let right: number;
</script>

<div style:right="{right}px">
  <Button type="ghost" noStroke square gray title="Edit" on:click={onEdit}>
    <Pen size="14px" />
  </Button>
  <Button type="ghost" noStroke square gray title="Delete" on:click={onDelete}>
    <Trash2Icon size="14px" />
  </Button>

  <Button
    type="ghost"
    noStroke
    square
    gray
    title="Duplicate"
    on:click={onDuplicate}
  >
    <CopyIcon size="14px" />
  </Button>
  <Button
    type="ghost"
    noStroke
    square
    gray
    disabled={first}
    title="Move to top"
    on:click={onMoveToTop}
  >
    <ArrowUpToLineIcon size="14px" />
  </Button>
  <Button
    type="ghost"
    noStroke
    square
    gray
    disabled={last}
    title="Move to bottom"
    on:click={onMoveToBottom}
  >
    <ArrowDownToLineIcon size="14px" />
  </Button>
</div>

<style lang="postcss">
  div {
    height: 39px;
    width: 192px;
    @apply bg-gray-50;
    @apply gap-x-2.5 px-4 py-2 flex items-center justify-center absolute top-0  z-50;
  }
</style>
