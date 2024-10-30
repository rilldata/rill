<script lang="ts">
  import {
    createAdminServiceListProjectsForOrganization,
    type V1Quotas,
  } from "@rilldata/web-admin/client";

  export let organization: string;
  export let quotas: V1Quotas;

  $: projects = createAdminServiceListProjectsForOrganization(organization);

  $: projectQuota =
    quotas.projects && quotas.projects !== "-1" ? quotas.projects : "Unlimited";
</script>

<div class="quotas">
  <div class="quota-entry">
    <div class="quota-entry-title">Projects</div>
    <div class="quota-entry-body">
      {$projects.data?.projects?.length} of {projectQuota}
    </div>
  </div>

  <!-- TODO: we need backend support for these -->
  <!--{#if singleProjectLimit}-->
  <!--  <div class="quota-entry">-->
  <!--    <div class="quota-entry-title">Data Size {dataSize}</div>-->
  <!--    <div>-->
  <!--      <Progress-->
  <!--        value={0}-->
  <!--        max={Number(quotas.storageLimitBytesPerDeployment)}-->
  <!--      />-->
  <!--    </div>-->
  <!--  </div>-->
  <!--{:else}-->
  <!--   <Button-->
  <!--     type="link"-->
  <!--     compact-->
  <!--     href="/#todo"-->
  <!--     forcedStyle="min-height: 18px !important;height: 18px !important;padding:0px !important;"-->
  <!--   >-->
  <!--     See project size breakdown-->
  <!--   </Button>-->
  <!--{/if}-->
</div>

<style lang="postcss">
  .quotas {
    @apply flex flex-row items-center gap-x-20 mt-2;
  }

  .quota-entry {
    @apply flex flex-col min-w-24;
  }

  .quota-entry-title {
    @apply font-semibold text-[10px] uppercase text-gray-500;
  }

  .quota-entry-body {
    @apply text-gray-800 text-xs;
  }
</style>
