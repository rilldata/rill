<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import {
    createLocalServiceListBranches,
    createLocalServiceCreatePreviewDeployment,
    createLocalServiceListPreviewDeployments,
    createLocalServiceDeletePreviewDeployment,
    getLocalServiceListPreviewDeploymentsQueryKey,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Rocket, ExternalLink, Trash2 } from "lucide-svelte";

  export let open = false;
  export let org = "";
  export let project = "";

  const queryClient = useQueryClient();
  const branchesQuery = createLocalServiceListBranches();

  $: currentBranch = $branchesQuery.data?.currentBranch ?? "";

  // Query for existing preview deployments
  $: previewDeploymentsQuery = createLocalServiceListPreviewDeployments(
    { org, project },
    { query: { enabled: !!org && !!project } },
  );

  // Mutation for creating preview deployment
  const createPreviewMutation = createLocalServiceCreatePreviewDeployment({
    mutation: {
      onSuccess: () => {
        // Invalidate the preview deployments query
        queryClient.invalidateQueries({
          queryKey: getLocalServiceListPreviewDeploymentsQueryKey(org, project),
        });
      },
    },
  });

  // Mutation for deleting preview deployment
  const deletePreviewMutation = createLocalServiceDeletePreviewDeployment({
    mutation: {
      onSuccess: () => {
        // Invalidate the preview deployments query
        queryClient.invalidateQueries({
          queryKey: getLocalServiceListPreviewDeploymentsQueryKey(org, project),
        });
      },
    },
  });

  $: isCreating = $createPreviewMutation.isPending;
  $: previewUrl = $createPreviewMutation.data?.frontendUrl ?? "";

  async function handleCreatePreview() {
    try {
      const result = await $createPreviewMutation.mutateAsync({ org, project });
      if (result.frontendUrl) {
        window.open(result.frontendUrl, "_blank");
      }
    } catch (error) {
      console.error("Failed to create preview deployment:", error);
    }
  }

  async function handleDeletePreview(deploymentId: string) {
    try {
      await $deletePreviewMutation.mutateAsync({ org, project, deploymentId });
    } catch (error) {
      console.error("Failed to delete preview deployment:", error);
    }
  }

  $: existingDeployments = $previewDeploymentsQuery.data?.deployments ?? [];
  $: currentBranchDeployment = existingDeployments.find(
    (d) => d.branch === currentBranch,
  );
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-md">
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-x-2">
        <Rocket size={18} />
        Preview Deployment
      </Dialog.Title>
      <Dialog.Description>
        Create a preview deployment for the current branch
      </Dialog.Description>
    </Dialog.Header>

    <div class="py-4 space-y-4">
      <div class="p-4 bg-slate-50 rounded">
        <div class="text-sm text-slate-600">
          <span class="font-medium">Branch:</span>
          <span class="font-mono ml-2">{currentBranch}</span>
        </div>
      </div>

      {#if currentBranchDeployment}
        <div class="p-4 bg-green-50 border border-green-200 rounded">
          <div class="text-sm text-green-700 font-medium mb-2">
            Preview deployment exists for this branch
          </div>
          <div class="flex items-center justify-between">
            <a
              href={currentBranchDeployment.frontendUrl}
              target="_blank"
              rel="noopener noreferrer"
              class="text-sm text-green-600 hover:underline flex items-center gap-x-1"
            >
              <ExternalLink size={12} />
              Open Preview
            </a>
            <button
              class="text-sm text-red-600 hover:text-red-700 flex items-center gap-x-1"
              on:click={() => handleDeletePreview(currentBranchDeployment.id)}
              disabled={$deletePreviewMutation.isPending}
            >
              <Trash2 size={12} />
              {$deletePreviewMutation.isPending ? "Deleting..." : "Delete"}
            </button>
          </div>
        </div>
      {:else}
        <div class="text-sm text-slate-600">
          Creating a preview deployment will:
          <ul class="list-disc list-inside mt-2 space-y-1">
            <li>Push your branch to the remote repository</li>
            <li>Create a new deployment environment</li>
            <li>Build and deploy your project</li>
            <li>Provide a unique preview URL</li>
          </ul>
        </div>

        <div
          class="p-3 bg-amber-50 border border-amber-200 rounded text-sm text-amber-700"
        >
          <strong>Note:</strong> Preview deployments are separate from production
          and can be used to test changes before merging.
        </div>
      {/if}

      {#if existingDeployments.length > 0 && !currentBranchDeployment}
        <div class="border-t pt-4">
          <div class="text-sm font-medium text-slate-700 mb-2">
            Other Preview Deployments
          </div>
          <div class="space-y-2 max-h-32 overflow-y-auto">
            {#each existingDeployments as deployment}
              <div
                class="flex items-center justify-between text-sm p-2 bg-slate-50 rounded"
              >
                <span class="font-mono text-slate-600">{deployment.branch}</span
                >
                <div class="flex items-center gap-x-2">
                  <a
                    href={deployment.frontendUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-blue-600 hover:underline"
                  >
                    <ExternalLink size={12} />
                  </a>
                  <button
                    class="text-red-600 hover:text-red-700"
                    on:click={() => handleDeletePreview(deployment.id)}
                  >
                    <Trash2 size={12} />
                  </button>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      {#if $createPreviewMutation.isError}
        <div
          class="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700"
        >
          <strong>Error:</strong>
          {$createPreviewMutation.error?.message ??
            "Failed to create preview deployment"}
        </div>
      {/if}
    </div>

    <Dialog.Footer>
      <Button type="secondary" onClick={() => (open = false)}>Cancel</Button>
      {#if !currentBranchDeployment}
        <Button
          type="primary"
          disabled={isCreating || !org || !project}
          onClick={handleCreatePreview}
        >
          {#if isCreating}
            Creating...
          {:else}
            <ExternalLink size={14} />
            Create Preview
          {/if}
        </Button>
      {/if}
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
