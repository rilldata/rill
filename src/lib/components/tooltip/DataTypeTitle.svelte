<script lang="ts">
export let name:string;
export let type:string;
</script>

<div class="grid gap-x-2 items-center pt-1 pb-1" style="grid-template-columns: auto max-content" style:min-width="200px">
    <div class="font-bold">
        {name}
    </div>
    <div class="text-gray-300 justify-self-end pl-3">
        {type}
    </div>
</div>