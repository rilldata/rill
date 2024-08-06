<script lang="ts">
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
</script>

<div
  class="m-auto flex-col justify-center items-center gap-y-4 inline-flex w-full"
>
  <div class="flex flex-col justify-center items-center">
    <div class="relative">
      <!-- TODO: replace this with a magic auth token icon if any -->
      <ReportIcon className="text-slate-300 w-12 h-12" />
    </div>
  </div>
  <div
    class="flex flex-col gap-y-2 justify-center items-center text-center text-sm"
  >
    <div class="text-gray-600 font-semibold">
      You don't have any magic auth tokens yet
    </div>
    <div>
      <span style="text-gray-500 font-normal">
        To create a magic auth token, click the Share button.
      </span>
    </div>
  </div>
</div>
