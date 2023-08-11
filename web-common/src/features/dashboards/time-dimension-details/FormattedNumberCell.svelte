<script lang="ts">
  export let cell: any;
</script>

<div class="text-right w-full">
  {typeof cell.d === "number" ? cell.d?.toFixed(2) : ""}
</div>
