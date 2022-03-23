<script>
export let hovered = false;
</script>
<div

    class="
        flex flex-row
        hover:bg-gray-100
        {hovered && 'bg-gray-100'}
    "    
>
    <slot />
</div>