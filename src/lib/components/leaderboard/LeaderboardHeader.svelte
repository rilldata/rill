<script>
    export let isActive = false;
</script>

<button 
    style:height="32px"
    style:grid-template-columns="auto max-content"
    class:font-bold={!isActive}

    class="
        grid w-full justify-between items-center pl-2 pr-2
        border
        border-gray-200
        rounded-t
        bg-white
    "
    on:click
> 
    <slot name="title"></slot>
    <slot name="right"></slot>
</button>