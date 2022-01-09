<script lang="ts">
export let selected = false;
export let title:string;
</script>

<button
    {title}
    class="
    text-black 
    w-7 h-7
    grid
    place-content-center place-items-center
    hover:border
    hover:border-gray-200 
    hover:border-solid
    " class:bg-gray-200={selected} on:click>
    <slot />
</button>