<script lang="ts">
import { tweened } from "svelte/motion";
import { cubicOut as easing } from "svelte/easing";
export let value = 0;
export let color = 'hsla(280, 50%, 90%)';
export let bgColor:string = 'bg-gray-50';
export let title:string;

const valueTween = tweened(0, {duration: 500, easing});
$: valueTween.set(value);

</script>

<div {title} class="
    text-right grid items-center relative w-full"
    style:background-color="hsla(217,5%, 90%, .25)"
>
    <div class='pl-2 pr-2' style="position: relative;"><slot /></div> 
    <div class='number-bar' style="--width: {$valueTween}; --color: {color}" />
</div>

<style>
    .number-bar {
        --width: 0%;
        content: '';
        display: inline-block;
        width: calc(100% * var(--width));
        position: absolute;
        left: 0;
        top: 0;
        height: 1rem;
        background-color: var(--color);
        
        mix-blend-mode: multiply;
        pointer-events: none;
    }
    </style> 