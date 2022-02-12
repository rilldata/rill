<script lang="ts">
import { createEventDispatcher } from "svelte";
const dispatch = createEventDispatcher();
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
            border-gray-300
            border-t-0
            border-l-0
            text-left
            bg-gray-200
        "
    >
        <slot />
    </div>
</th>