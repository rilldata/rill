<script lang="ts">
import { tweened } from "svelte/motion";
import { cubicOut as easing } from "svelte/easing";
export let value = 0;
export let color = 'hsl(280, 50%, 90%)';
export let bgColor:string = 'bg-gray-50';
export let title:string;

const valueTween = tweened(0, {duration: 500, easing});
$: valueTween.set(value);

</script>

<div {title} class="text-right {bgColor} grid items-center" style="position: relative; width: 100%;">
    <div class='pl-2 pr-2' style="position: relative; z-index: 3;"><slot /></div> 
    <div class='number-bar' style="--width: {$valueTween}; --color: {color}" />
</div>

<style>
    .number-bar {
        --width: 0%;
        content: '';
        display: inline-block;
        width: calc(100% * var(--width));
        position: absolute;
        z-index: 2;
        left: 0;
        top: 0;
        height: 1rem;
        background-color: var(--color);
    }
    </style> 