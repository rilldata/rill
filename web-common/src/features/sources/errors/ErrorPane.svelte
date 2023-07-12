<script lang="ts">
  import CancelCircle from "../../../components/icons/CancelCircle.svelte";

  export let error: any;
</script>

<div
  class="w-full h-full m-4 bg-white rounded border border-gray-200 flex-col justify-center inline-flex"
>
  <div class="flex-col justify-start items-center gap-1 flex text-red-500">
    <CancelCircle size="24px" />
    <div class="text-center text-sm font-medium">
      {error.message}
    </div>
  </div>
</div>
