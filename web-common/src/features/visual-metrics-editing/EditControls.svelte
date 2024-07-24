<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import {
    ArrowUpToLineIcon,
    ArrowDownToLineIcon,
    Pen,
    CopyIcon,
    Trash2Icon,
  } from "lucide-svelte";
</script>

<div>
  <Button type="ghost" noStroke square gray title="Edit">
    <Pen size="14px" />
  </Button>
  <Button type="ghost" noStroke square gray title="Delete">
    <Trash2Icon size="14px" />
  </Button>

  <Button type="ghost" noStroke square gray title="Duplicate">
    <CopyIcon size="14px" />
  </Button>
  <Button type="ghost" noStroke square gray title="Move to top">
    <ArrowUpToLineIcon size="14px" />
  </Button>
  <Button type="ghost" noStroke square gray title="Move to bottom">
    <ArrowDownToLineIcon size="14px" />
  </Button>
</div>

<style lang="postcss">
  div {
    width: 192px;
    @apply bg-gray-50;
    @apply gap-x-2.5 px-4 py-2 flex items-center justify-center absolute h-10 top-0 right-0 z-50;
  }
</style>
