<script lang="ts">
  export let path: string;
  export let addEllipses = true;
</script>

<div
  class="w-full text-ellipsis overflow-hidden whitespace-nowrap"
  style="direction: rtl;
text-align: left;"
>
  {path}{#if addEllipses}<span class="text-gray-400">/..</span>{/if}
</div>
