<script lang="ts">
  import {
    createAdminServiceListProjectsForOrganization,
    type V1Quotas,
  } from "@rilldata/web-admin/client";
  import { formatDataSizeQuota } from "@rilldata/web-admin/features/billing/plans/utils";
  import { Progress } from "@rilldata/web-common/components/progress";

  export let organization: string;
  export let quotas: V1Quotas;

  $: projects = createAdminServiceListProjectsForOrganization(organization);

  $: singleProjectLimit = quotas.projects === "1";
  $: dataSize = singleProjectLimit
    ? ""
    : quotas.storageLimitBytesPerDeployment
      ? formatDataSizeQuota(quotas.storageLimitBytesPerDeployment)
      : "";

  $: projectQuota =
    quotas.projects && quotas.projects !== "-1" ? quotas.projects : "Unlimited";
</script>

<div class="quotas">
  <div class="quota-entry">
    <div class="quota-entry-title">Projects</div>
    <div>
      {$projects.data?.projects?.length} of {projectQuota}
    </div>
  </div>

  <div class="quota-entry">
    <div class="quota-entry-title">Data Size {dataSize}</div>
    <div>
      {#if singleProjectLimit}
        <Progress
          value={0}
          max={Number(quotas.storageLimitBytesPerDeployment)}
        />
      {:else}
        <!-- We do not have a way to isolate the data -->
        <!--        <Button-->
        <!--          type="link"-->
        <!--          compact-->
        <!--          href="/#todo"-->
        <!--          forcedStyle="min-height: 18px !important;height: 18px !important;padding:0px !important;"-->
        <!--        >-->
        <!--          See project size breakdown-->
        <!--        </Button>-->
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .quotas {
    @apply flex flex-row items-center gap-x-20 mt-2;
  }

  .quota-entry {
    @apply flex flex-col min-w-24;
  }

  .quota-entry-title {
    @apply font-semibold text-[10px] uppercase;
  }
</style>
