<script>
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
export let rotated = false;
export let suppressTooltip = false;

</script>

<Tooltip location="top" alignment="start" distance={8} pad={8} suppress={suppressTooltip}>

<button
on:click
style:width="16px"
style:height="16px"
style:grid-column="left-control"
class="
    hover:bg-gray-300
    transition-tranform 
    text-gray-500
    duration-100
    grid
    items-center
    justify-center
    rounded
    {!rotated ? '-rotate-90' : ''}"
>
    <slot />
</button>
<TooltipContent slot="tooltip-content">
    {!rotated ? "show columns" : "hide columne"}
</TooltipContent>
</Tooltip>