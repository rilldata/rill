<script lang="ts">
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
</script>

<div
  class="m-auto mt-20 pb-10 flex-col justify-center items-center gap-y-4 inline-flex"
>
  <div class="flex flex-col justify-center items-center">
    <div class="relative">
      <AlertCircleOutline className="text-gray-300 w-12 h-12" />
    </div>
  </div>
  <div
    class="flex flex-col gap-y-2 justify-center items-center text-center text-sm"
  >
    <div class="text-fg-secondary font-semibold">
      {m.alert_has_not_run_yet()}
    </div>
  </div>
</div>
