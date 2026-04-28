<script lang="ts">
  import { page } from "$app/state";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import {
    extractBranchFromPath,
    injectBranchIntoPath,
  } from "@rilldata/web-admin/features/branches/branch-utils.ts";
  import { goto } from "$app/navigation";
  import { projectWelcomeStatus } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";

  const runtimeClient = useRuntimeClient();

  let branch = $derived(extractBranchFromPath(page.url.pathname));
  let { organization, project } = $derived(page.params);

  const filesQuery = createRuntimeServiceListFiles(runtimeClient, {});

  // On cloud, we do not have loader functions for edit sessions so that data is loaded async.
  // So we need this to add a redirect to welcome page.
  $effect(() => {
    if (!$filesQuery.isSuccess) return;
    const hasRillYaml = $filesQuery.data?.files?.some(
      (file) => file.path === "/rill.yaml",
    );
    if (!hasRillYaml) {
      projectWelcomeStatus.setProjectWelcomeStep(project, true);
      void goto(
        injectBranchIntoPath(
          `/${organization}/${project}/-/edit/welcome`,
          branch,
        ),
      );
    }
  });
</script>
