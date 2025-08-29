<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import {
    getDashboardToRedirect,
    useRefetchingDashboards,
  } from "@rilldata/web-admin/features/dashboards/listing/refetching-dashboards.ts";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { deploying, deployingName } = data;

  $: ({
    params: { organization, project },
  } = $page);
  $: ({ instanceId } = $runtime);

  $: query = useRefetchingDashboards(instanceId, deployingName);
  $: ({ data: queryData } = $query);
  $: ({ dashboards, dashboard, dashboardsReconciling, dashboardsErrored } =
    queryData ?? {
      dashboards: [],
    });

  $: dashboardToRedirect =
    deploying && dashboard
      ? getDashboardToRedirect(organization, project, dashboard)
      : null;
  $: if (dashboardToRedirect) {
    void goto(dashboardToRedirect);
  }
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>

{#if dashboardsErrored}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <div class="h-36">
        <CancelCircleInverse size="7em" className="text-gray-200" />
      </div>
      <CtaHeader variant="bold">
        Sorry, your dashboard isn't working right now!
      </CtaHeader>
      <p class="text-gray-500 text-base">
        <ProjectAccessControls {organization} {project}>
          <svelte:fragment slot="manage-project">
            View project status for errors that may help you find a fix.
          </svelte:fragment>
          <svelte:fragment slot="read-project">
            Contact your project's admin for help.
          </svelte:fragment>
        </ProjectAccessControls>
      </p>
      <ProjectAccessControls {organization} {project}>
        <svelte:fragment slot="manage-project">
          <CtaButton
            variant="secondary"
            href={`/${organization}/${project}/-/status`}
          >
            View project status
          </CtaButton>
        </svelte:fragment>
        <svelte:fragment slot="read-project">
          <CtaButton variant="secondary" href={`/${organization}/${project}`}>
            View project
          </CtaButton>
        </svelte:fragment>
      </ProjectAccessControls>
      <CtaNeedHelp />
    </CtaContentContainer>
  </CtaLayoutContainer>
{:else if dashboardsReconciling || deploying}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <div class="h-36">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
      <CtaHeader variant="bold">
        Hang tight! We're building your dashboards...
      </CtaHeader>
      <CtaNeedHelp />
    </CtaContentContainer>
  </CtaLayoutContainer>
{:else}
  <ContentContainer
    maxWidth={800}
    title="Project dashboards"
    showTitle={dashboards?.length > 0}
  >
    <div class="flex flex-col items-center gap-y-4">
      <DashboardsTable />
    </div>
  </ContentContainer>
{/if}
