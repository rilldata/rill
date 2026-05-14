<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useParserReconcileError } from "../selectors";
  import DeploymentSection from "./DeploymentSection.svelte";
  import ResourcesSection from "./ResourcesSection.svelte";
  import TablesSection from "./TablesSection.svelte";
  import ErrorsSection from "./ErrorsSection.svelte";

  export let organization: string;
  export let project: string;

  const runtimeClient = useRuntimeClient();
  $: parserErrorQuery = useParserReconcileError(runtimeClient);
  $: hasProjectError = !!($parserErrorQuery.data ?? "");

  // Only hide the sections below Deployment when the parser error is actually
  // surfaced to the user (i.e. on GitHub-connected projects). Otherwise a
  // silent parser error on a Rill-managed project leaves the page looking
  // empty with no explanation.
  $: proj = createAdminServiceGetProject(organization, project);
  $: projectData = $proj.data?.project;
  $: githubUrl = projectData?.gitRemote
    ? getGitUrlFromRemote(projectData.gitRemote)
    : "";
  $: isGithubConnected =
    !!projectData?.gitRemote && !projectData?.managedGitId && !!githubUrl;
  $: hideBelow = hasProjectError && isGithubConnected;
</script>

<div class="flex flex-col gap-6">
  <DeploymentSection {organization} {project} />
  {#if !hideBelow}
    <ResourcesSection />
    <TablesSection />
    <ErrorsSection />
  {/if}
</div>
