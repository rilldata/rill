<script lang="ts">
  import DashboardList from "../home/DashboardList.svelte";

  export let organization: string;
  export let project: string;
</script>

<div class="flex flex-col gap-y-1 max-h-[170px]">
  <span class="uppercase text-gray-500 font-semibold text-[10px] leading-4"
    >Dashboards</span
  >
  <div class="overflow-y-auto">
    <DashboardList {organization} {project} />
  </div>
</div>
