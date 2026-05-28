<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    V1PersonalVirtualFileType,
    V1PersonalVirtualFileSourceKind,
    adminServiceCreatePersonalVirtualFile,
    adminServiceCopyPersonalVirtualFile,
  } from "@rilldata/web-admin/client";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { createEventDispatcher } from "svelte";

  export let org: string;
  export let project: string;
  export let copyableCanvases: { name: string; displayName: string }[] = [];
  export let open = false;

  const dispatch = createEventDispatcher<{ created: { name: string } }>();

  let displayName = "";
  let mode: "blank" | "copy" = "blank";
  let copySourceName = "";
  let submitting = false;
  let error: string | null = null;

  async function submit() {
    error = null;
    if (!displayName.trim()) {
      error = "Display name is required";
      return;
    }
    submitting = true;
    try {
      let name: string | undefined;
      if (mode === "copy") {
        if (!copySourceName) {
          error = "Pick a canvas to copy from";
          submitting = false;
          return;
        }
        const res = await adminServiceCopyPersonalVirtualFile(org, project, {
          type: V1PersonalVirtualFileType.PERSONAL_VIRTUAL_FILE_TYPE_CANVAS,
          sourceKind:
            V1PersonalVirtualFileSourceKind.PERSONAL_VIRTUAL_FILE_SOURCE_KIND_SHARED,
          sourceName: copySourceName,
          displayName: displayName.trim(),
        });
        name = res.name;
      } else {
        const res = await adminServiceCreatePersonalVirtualFile(org, project, {
          type: V1PersonalVirtualFileType.PERSONAL_VIRTUAL_FILE_TYPE_CANVAS,
          displayName: displayName.trim(),
        });
        name = res.name;
      }
      if (name) {
        dispatch("created", { name });
        open = false;
        await goto(
          `/${org}/${project}/-/my-canvases/${encodeURIComponent(name)}?mode=edit`,
        );
      }
    } catch (e) {
      error = (e as Error)?.message ?? "Failed to create canvas";
    } finally {
      submitting = false;
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Create personal canvas</Dialog.Title>
      <Dialog.Description>
        Personal canvases are only visible to you. They live alongside the
        project but never sync to git.
      </Dialog.Description>
    </Dialog.Header>

    <div class="flex flex-col gap-4">
      <label class="flex flex-col gap-1 text-sm">
        <span>Display name</span>
        <input
          class="border rounded px-2 py-1"
          bind:value={displayName}
          placeholder="e.g. My revenue dashboard"
          disabled={submitting}
        />
      </label>

      <fieldset class="flex flex-col gap-2 text-sm">
        <legend class="font-medium">Start from</legend>
        <label class="flex items-center gap-2">
          <input type="radio" bind:group={mode} value="blank" />
          Blank canvas
        </label>
        <label class="flex items-center gap-2">
          <input
            type="radio"
            bind:group={mode}
            value="copy"
            disabled={copyableCanvases.length === 0}
          />
          Copy from an existing canvas
        </label>
        {#if mode === "copy"}
          <select
            class="border rounded px-2 py-1"
            bind:value={copySourceName}
            disabled={submitting}
          >
            <option value="">Select a canvas...</option>
            {#each copyableCanvases as c (c.name)}
              <option value={c.name}>{c.displayName || c.name}</option>
            {/each}
          </select>
        {/if}
      </fieldset>

      {#if error}
        <p class="text-red-600 text-sm">{error}</p>
      {/if}
    </div>

    <Dialog.Footer>
      <Button type="secondary" onClick={() => (open = false)}>Cancel</Button>
      <Button type="primary" onClick={submit} disabled={submitting}>
        {submitting ? "Creating..." : "Create"}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
