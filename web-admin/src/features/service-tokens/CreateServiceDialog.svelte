<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateService,
    createAdminServiceIssueServiceAuthToken,
    createAdminServiceSetProjectMemberServiceRole,
    createAdminServiceListProjectsForOrganization,
    getAdminServiceListServicesQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { CopyIcon } from "lucide-svelte";
  import { validateServiceName } from "./utils";
  import ServiceForm from "./ServiceForm.svelte";

  let { open = $bindable(false) }: { open: boolean } = $props();

  let name = $state("");
  let orgRole = $state("");
  let projectAssignments: { project: string; role: string }[] = $state([]);
  let attributes: { key: string; value: string }[] = $state([]);
  let issuedToken = $state("");
  let tokenCopied = $state(false);
  let step: "form" | "token" = $state("form");

  let organization = $derived($page.params.organization);
  let projectsQuery = $derived(
    createAdminServiceListProjectsForOrganization(organization),
  );
  let allProjects = $derived($projectsQuery.data?.projects ?? []);

  let nameError = $derived(name ? validateServiceName(name) : "");
  let hasAtLeastOneAssignment = $derived(
    orgRole !== "" || projectAssignments.length > 0,
  );
  let isValid = $derived(
    name.trim() !== "" && !nameError && hasAtLeastOneAssignment,
  );

  const queryClient = useQueryClient();
  const createService = createAdminServiceCreateService();
  const issueToken = createAdminServiceIssueServiceAuthToken();
  const setProjectRole = createAdminServiceSetProjectMemberServiceRole();

  function handleReset() {
    name = "";
    orgRole = "";
    projectAssignments = [];
    attributes = [];
    issuedToken = "";
    tokenCopied = false;
    step = "form";
  }

  async function handleSubmit() {
    try {
      const firstProject = projectAssignments[0];
      const attrObj = Object.fromEntries(
        attributes
          .filter((a) => a.key.trim())
          .map((a) => [a.key.trim(), a.value]),
      );

      await $createService.mutateAsync({
        org: organization,
        data: {
          name: name.trim(),
          ...(orgRole ? { orgRoleName: orgRole } : {}),
          ...(firstProject
            ? {
                project: firstProject.project,
                projectRoleName: firstProject.role,
              }
            : {}),
          ...(Object.keys(attrObj).length > 0 ? { attributes: attrObj } : {}),
        },
      });

      for (let i = 1; i < projectAssignments.length; i++) {
        const pa = projectAssignments[i];
        await $setProjectRole.mutateAsync({
          org: organization,
          project: pa.project,
          name: name.trim(),
          data: { role: pa.role },
        });
      }

      const result = await $issueToken.mutateAsync({
        org: organization,
        serviceName: name.trim(),
        data: {},
      });

      issuedToken = result.token ?? "";

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListServicesQueryKey(organization),
      });

      step = "token";
    } catch (e: any) {
      console.error("Error creating service", e);
      eventBus.emit("notification", {
        message: e?.response?.data?.message ?? "Error creating service",
        type: "error",
      });
    }
  }

  function handleClose() {
    open = false;
    handleReset();
  }
</script>

<Dialog
  bind:open
  onOpenChange={(isOpen) => {
    if (!isOpen) handleReset();
  }}
>
  <DialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>
        {step === "form" ? "Create service" : "Service created"}
      </DialogTitle>
    </DialogHeader>

    {#if step === "form"}
      <DialogDescription>
        Create a service account to access Rill programmatically.
      </DialogDescription>
      <ServiceForm
        bind:name
        bind:orgRole
        bind:projectAssignments
        bind:attributes
        {nameError}
        {allProjects}
        formId="create-service-form"
        showOptionalLabels
        onSubmit={handleSubmit}
      />
      <DialogFooter>
        <Button type="tertiary" onClick={handleClose}>Cancel</Button>
        <Button
          type="primary"
          form="create-service-form"
          disabled={!isValid || $createService.isPending}
          submitForm
        >
          Create
        </Button>
      </DialogFooter>
    {:else}
      <!-- Token display step -->
      <div class="flex flex-col gap-y-4">
        <p class="text-sm text-fg-tertiary">
          Service <span class="font-medium text-fg-primary">{name}</span> has been
          created. Copy the token below — it will not be shown again.
        </p>
        <div class="flex items-center gap-x-2">
          <code
            class="text-xs bg-surface-subtle border rounded px-2 py-2 flex-1 break-all select-all"
          >
            {issuedToken}
          </code>
          <IconButton
            onclick={() => {
              copyToClipboard(issuedToken);
              tokenCopied = true;
            }}
          >
            <CopyIcon size="14px" />
          </IconButton>
        </div>
        <p class="text-xs text-fg-secondary">
          This token will only be shown once. Make sure to copy it now.
        </p>
      </div>
      <DialogFooter>
        {#if tokenCopied}
          <Button type="primary" onClick={handleClose}>Done</Button>
        {:else}
          <Button
            type="primary"
            onClick={() => {
              copyToClipboard(issuedToken);
              tokenCopied = true;
            }}
          >
            <CopyIcon size="14px" />
            Copy token
          </Button>
        {/if}
      </DialogFooter>
    {/if}
  </DialogContent>
</Dialog>
