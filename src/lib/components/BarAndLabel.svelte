<script lang="ts">
import { tweened } from "svelte/motion";
import { cubicOut as easing } from "svelte/easing";
export let value = 0;
export let color;
export let showBackground = true;
export let title:string;

const valueTween = tweened(0, {duration: 500, easing});
$: valueTween.set(value);

</script>

<div {title} class="
    text-right grid items-center relative w-full"
    style:background-color={showBackground ? "hsla(217,5%, 90%, .25)" : 'hsl(217, 0%, 100%, .25)'}
>
    <div class='pl-2 pr-2' style="position: relative;"><slot /></div> 
    <div class='number-bar {color}' style="--width: {$valueTween};" />
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
        
        mix-blend-mode: multiply;
        pointer-events: none;
    }
    </style> 