<script>
    export let isActive = false;
</script>

<button 
    style:height="32px"
    style:grid-template-columns="auto max-content"
    class:font-bold={!isActive}
    bg-gray-100={!isActive}
    bg-white={isActive}
    class="
        grid w-full justify-between items-center pl-2 pr-2
        border
        border-gray-300
        hover:bg-gray-200
        rounded-t
    "
    on:click
> 
    <slot name="title"></slot>
    <slot name="right"></slot>
</button>