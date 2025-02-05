<script lang="ts">
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import { BellIcon } from "lucide-svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";

  export let kind: "report" | "dashboard" | "alert";

  const icons = {
    report: ReportIcon,
    alert: BellIcon,
    dashboard: ExploreIcon,
  };
</script>

<div
  class="m-auto mt-20 pb-10 flex-col justify-center items-center gap-y-4 inline-flex"
>
  <div class="flex flex-col justify-center items-center">
    <div class="relative">
      <svelte:component
        this={icons[kind]}
        className="text-slate-300 w-12 h-12"
      />
    </div>
  </div>
  <div
    class="flex flex-col gap-y-2 justify-center items-center text-center text-sm"
  >
    <div class="text-gray-600 font-semibold">
      You don't have any {kind}s yet
    </div>
    <div>
      <span style="text-gray-500 font-normal">
        <slot />
      </span>
    </div>
  </div>
</div>
