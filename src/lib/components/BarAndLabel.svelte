<script lang="ts">
export let value = 0;
export let color = 'hsl(280, 50%, 90%)';
</script>

<div class="text-right" style="position:relative;">
        <slot />
    <div class='number-bar' style="--width: {value}; --color: {color}" />
</div>

<style>
    .number-bar {
        --width: 0%;
        content: '';
        display: inline-block;
        width: calc(100% * var(--width));
        position: absolute;
        left: 0;
        z-index: -1;
        height: 1rem;
        background-color: var(--color);
    }
    </style>