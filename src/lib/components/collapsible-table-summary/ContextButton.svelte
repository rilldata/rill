<script>
    export let rotated = false;
    </script>
    
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
        border
        border-transparent
        hover:border-gray-400
        {rotated ? '-rotate-90' : ''}"
    >
        <slot />
    </button>