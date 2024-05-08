<script lang="ts">
  import { BellIcon } from "lucide-svelte";
</script>

<div
  class="m-auto mt-20 pb-10 flex-col justify-center items-center gap-y-4 inline-flex"
>
  <div class="flex flex-col justify-center items-center">
    <div class="relative">
      <BellIcon className="text-slate-300 w-12 h-12" />
    </div>
  </div>
  <div
    class="flex flex-col gap-y-2 justify-center items-center text-center text-sm"
  >
    <div class="text-gray-600 font-semibold">You don't have any alerts yet</div>
    <div>
      <span style="text-gray-500 font-normal">
        To create an alert, click the "Create alert" button in a dashboard.
      </span>
    </div>
  </div>
</div>
