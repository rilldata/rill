<script lang="ts">
  import RedeployProjectCta from "@rilldata/web-admin/features/projects/RedeployProjectCTA.svelte";
  import { createAdminServiceListDeployments } from "@rilldata/web-admin/client";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import { injectBranchIntoPath } from "@rilldata/web-admin/features/branches/branch-utils.ts";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";

  let { organization, project }: { organization: string; project: string } =
    $props();

  let deploymentsQuery = $derived(
    createAdminServiceListDeployments(organization, project, {}),
  );
  let prodDeployment = $derived(
    $deploymentsQuery.data?.deployments?.find(
      (d) => !d.editable && d.environment === "prod",
    ),
  );
  let editableDeployment = $derived(
    $deploymentsQuery.data?.deployments?.find((d) => d.editable),
  );
</script>

{#if $deploymentsQuery.isPending}
  <CtaLayoutContainer>
    <Spinner status={EntityStatus.Running} size="3rem" duration={725} />
  </CtaLayoutContainer>
{:else if !prodDeployment && editableDeployment}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <CtaHeader variant="bold">
        This project hasn't been published yet. What would you like to do next?
      </CtaHeader>
      <div class="flex flex-row gap-2 justify-start">
        <Button
          type="secondary"
          href={injectBranchIntoPath(
            `/${organization}/${project}/-/edit`,
            editableDeployment.branch,
          )}
        >
          Continue editing
        </Button>
        <Tooltip distance={8}>
          <Button type="primary" disabled>Publish</Button>
          <TooltipContent slot="tooltip-content" maxWidth="200px">
            <span class="text-xs">Coming soon</span>
          </TooltipContent>
        </Tooltip>
      </div>
    </CtaContentContainer>
  </CtaLayoutContainer>
{:else}
  <RedeployProjectCta {organization} {project} />
{/if}
