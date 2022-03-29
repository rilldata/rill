<script lang="ts">
export let sticky:string = 'top';
</script>
<th class="
    p-0
"
    class:sticky={!!sticky}
    class:top-0={sticky === 'top'}
    class:left-0={sticky === 'left'}
    class:z-10={!!sticky}

>
    <div 
        class="
            p-3
            pl-4
            pr-4
            border
            border-gray-200
            border-t-0
            border-l-0
            text-left
            bg-gray-100
        "
    >
        <slot />
    </div>
</th>