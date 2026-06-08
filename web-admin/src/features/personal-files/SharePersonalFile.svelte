<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import Button from "web-common/src/components/button/Button.svelte";
  import { createAdminServiceGetPersonalFile } from "@rilldata/web-admin/client";
  import { parseDocument, YAMLMap } from "yaml";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";

  let {
    organization,
    project,
    name,
  }: {
    organization: string;
    project: string;
    name: string;
  } = $props();

  let open = $state(false);
  let sharing = $state(false);

  let personalFileQuery = $derived(
    createAdminServiceGetPersonalFile(organization, project, name),
  );
  let { data, isPending } = $derived($personalFileQuery);

  let parsedDocument = $derived(parseDocument(data?.yaml ?? ""));
  let shared = $derived(
    (parsedDocument.get("annotations") as YAMLMap | null)?.get(
      "admin_shared",
    ) === "true",
  );
  $effect(() => console.log("shared", shared));

  let loading = $derived(isPending || sharing);

  async function handleShareToggle(share: boolean) {
    if (!data) return;

    sharing = true;
    try {
      (parsedDocument.get("annotations") as YAMLMap | null)?.set(
        "admin_shared",
        share ? "true" : "false",
      );
      const yaml = parsedDocument.toString();

      const fileArtifact = fileArtifacts.getFileArtifact(
        removeLeadingSlash(data.path ?? ""),
      );
      fileArtifact.updateEditorContent(yaml);
      await fileArtifact.saveLocalContent();

      eventBus.emit("notification", {
        type: "success",
        message: `Dashboard ${share ? "shared" : "hidden"} successfully`,
      });
    } catch (e) {
      console.error("Error sharing dashboard:", e);
    }
    sharing = false;
    open = false;
  }
</script>

<Popover.Root bind:open>
  <Popover.Trigger>
    {#snippet child({ props })}
      <Tooltip.Root disabled={open}>
        <Tooltip.Trigger>
          <Button
            {...props}
            type="secondary"
            selected={open}
            loading={isPending}
          >
            Share
          </Button>
        </Tooltip.Trigger>
        <Tooltip.Content>Share dashboard</Tooltip.Content>
      </Tooltip.Root>
    {/snippet}
  </Popover.Trigger>
  <Popover.Content align="end">
    {#if shared}
      This dashboard is shared with all users in the project.
    {:else}
      Share this dashboard with other users in the project?
    {/if}

    <div class="flex pt-2">
      <div class="grow"></div>
      <Button
        type="primary"
        onClick={() => handleShareToggle(!shared)}
        {loading}
        disabled={loading}
      >
        {shared ? "Hide" : "Share"}
      </Button>
    </div>
  </Popover.Content>
</Popover.Root>
