<script lang="ts">
export let value = 0;
export let color = 'hsl(280, 50%, 90%)';
export let bgColor:string = 'bg-gray-50';
export let title:string;
</script>

<div {title} class="text-right {bgColor}" style="position:relative;">
    <div class='pl-2 pr-2' style="position: relative; z-index: 3;"><slot /></div> 
    <div class='number-bar' style="--width: {value}; --color: {color}" />
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