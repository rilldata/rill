<script>
  import EmptyTableIcon from "../pivot/EmptyTableIcon.svelte";
</script>

<div class="flex flex-col items-center w-full h-full justify-center gap-y-3">
  <EmptyTableIcon />
  <div class="flex flex-col items-center gap-y-2">
    <div class="font-semibold text-gray-600 mt-1 text-lg">
      Your canvas dashboard looks lonely
    </div>
  </div>
  <div class="text-gray-600">
    <button class="text-primary-600" on:click={() => console.log("sfd")}
      >Add a component</button
    > to get started
  </div>
</div>
