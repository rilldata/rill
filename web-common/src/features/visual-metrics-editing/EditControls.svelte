<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
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
  export let onMoveTo: (top: boolean) => void;
  export let first = false;
  export let last = false;
  export let selected = false;

  let type: "subtle" | "ghost";
  $: type = selected ? "subtle" : "ghost";
</script>

<Tooltip distance={8} activeDelay={500}>
  <Button {type} noStroke gray={!selected} square on:click={onEdit}>
    <Pen size="14px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    <span>Edit</span>
  </TooltipContent>
</Tooltip>

<Tooltip distance={8} activeDelay={500}>
  <Button {type} noStroke square gray={!selected} on:click={onDelete}>
    <Trash2Icon size="14px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    <span>Delete</span>
  </TooltipContent>
</Tooltip>

<Tooltip distance={8} activeDelay={500}>
  <Button {type} noStroke square gray={!selected} on:click={onDuplicate}>
    <CopyIcon size="14px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    <span>Duplicate</span>
  </TooltipContent>
</Tooltip>

<Tooltip distance={8} activeDelay={500}>
  <Button
    {type}
    noStroke
    square
    gray={!selected}
    disabled={first}
    on:click={() => onMoveTo(true)}
  >
    <ArrowUpToLineIcon size="14px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    <span>Move to top</span>
  </TooltipContent>
</Tooltip>

<Tooltip distance={8} activeDelay={500}>
  <Button
    {type}
    noStroke
    square
    gray={!selected}
    disabled={last}
    on:click={() => onMoveTo(false)}
  >
    <ArrowDownToLineIcon size="14px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    <span>Move to bottom</span>
  </TooltipContent>
</Tooltip>
