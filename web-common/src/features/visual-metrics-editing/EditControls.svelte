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
  export let itemType: "measures" | "dimensions";
  export let name: string;

  let type: "subtle" | "ghost";
  $: type = selected ? "subtle" : "ghost";

  $: singularType = itemType.slice(0, -1);
</script>

<Tooltip distance={8} activeDelay={500}>
  <Button
    {type}
    noStroke
    gray={!selected}
    square
    on:click={onEdit}
    label="Edit {singularType} {name}"
  >
    <Pen size="14px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    <span>Edit</span>
  </TooltipContent>
</Tooltip>

<Tooltip distance={8} activeDelay={500}>
  <Button
    {type}
    noStroke
    square
    gray={!selected}
    on:click={onDelete}
    label="Delete {singularType} {name}"
  >
    <Trash2Icon size="14px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    <span>Delete</span>
  </TooltipContent>
</Tooltip>

<Tooltip distance={8} activeDelay={500}>
  <Button
    {type}
    noStroke
    square
    gray={!selected}
    on:click={onDuplicate}
    label="Duplicate {singularType} {name}"
  >
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
    label="Move {singularType} {name} to top"
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
    label="Move {singularType} {name} to bottom"
    on:click={() => onMoveTo(false)}
  >
    <ArrowDownToLineIcon size="14px" />
  </Button>
  <TooltipContent slot="tooltip-content">
    <span>Move to bottom</span>
  </TooltipContent>
</Tooltip>
