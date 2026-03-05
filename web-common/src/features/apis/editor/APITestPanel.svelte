<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { PlusIcon, PlayIcon, ChevronDownIcon, CopyIcon } from "lucide-svelte";
  import APIResponsePreview from "./APIResponsePreview.svelte";
  import type { Arg } from "./types";

  export let apiName: string;
  export let hasErrors: boolean;
  export let isReconciling: boolean;
  export let host: string;
  export let instanceId: string;
  export let args: Arg[];

  let apiResponse: unknown[] | null = null;
  let responseError: string | null = null;
  let isLoading = false;
  let previewHeight = 200;

  // Clear response when switching to a different API
  $: apiName, resetResponse();
  function resetResponse() {
    apiResponse = null;
    responseError = null;
    isLoading = false;
  }

  $: baseUrl = `${host}/v1/instances/${instanceId}/api/${apiName}`;
  $: fullUrl = buildFullUrl(baseUrl, args);
  $: isDisabled = hasErrors || isReconciling;

  function buildFullUrl(base: string, params: Arg[]): string {
    try {
      const url = new URL(base);
      params.forEach((arg) => {
        if (arg.key.trim()) {
          url.searchParams.set(arg.key, arg.value);
        }
      });
      return url.toString();
    } catch {
      return base;
    }
  }

  function addArg() {
    args = [...args, { id: crypto.randomUUID(), key: "", value: "" }];
  }

  function removeArg(id: string) {
    args = args.filter((arg) => arg.id !== id);
  }

  function handleCopyUrl() {
    copyToClipboard(fullUrl, "Copied endpoint URL to clipboard");
  }

  function handleArgsKeydown(e: KeyboardEvent) {
    if (e.key !== "Tab") return;

    const container = (e.target as HTMLElement)?.closest(".args-container");
    if (!container) return;

    const inputs = Array.from(
      container.querySelectorAll<HTMLInputElement>("input"),
    );
    const currentIndex = inputs.indexOf(e.target as HTMLInputElement);
    if (currentIndex === -1) return;

    if (e.shiftKey) {
      // Shift+Tab: move to previous input, or let dropdown handle if at first
      if (currentIndex > 0) {
        e.preventDefault();
        inputs[currentIndex - 1].focus();
      }
    } else {
      // Tab: move to next input, or add a new arg row if at the last one
      if (currentIndex < inputs.length - 1) {
        e.preventDefault();
        inputs[currentIndex + 1].focus();
      } else {
        e.preventDefault();
        addArg();
        // Focus the new row's key input after Svelte updates the DOM
        requestAnimationFrame(() => {
          const updatedInputs = Array.from(
            container.querySelectorAll<HTMLInputElement>("input"),
          );
          updatedInputs[updatedInputs.length - 2]?.focus();
        });
      }
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter" && !isDisabled) {
      // Don't fire when focus is in a dialog, modal, or navigation sidebar
      const target = e.target as HTMLElement;
      if (target?.closest?.('[role="dialog"], nav, [role="alertdialog"]'))
        return;
      e.preventDefault();
      testAPI();
    }
  }

  async function testAPI() {
    isLoading = true;
    responseError = null;
    apiResponse = null;

    try {
      const response = await fetch(fullUrl);

      if (!response.ok) {
        const errorText = await response.text();
        try {
          const errorJson = JSON.parse(errorText);
          responseError = errorJson.message || errorJson.error || errorText;
        } catch {
          responseError = errorText;
        }
        return;
      }

      const data = await response.json();
      apiResponse = Array.isArray(data) ? data : [data];
    } catch (e) {
      responseError = e instanceof Error ? e.message : "Unknown error occurred";
    } finally {
      isLoading = false;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div
  class="preview-panel"
  style:height="{previewHeight}px"
  style:min-height="100px"
  style:max-height="60%"
>
  <Resizer max={500} direction="NS" side="top" bind:dimension={previewHeight} />

  <div class="flex items-center gap-x-3 px-3 py-2 border-b">
    <div class="flex items-center gap-x-2 flex-1 min-w-0">
      <span class="text-xs font-medium text-fg-secondary shrink-0">GET</span>
      <span class="text-xs font-mono text-fg-muted truncate">{fullUrl}</span>
    </div>

    <div class="flex items-center gap-x-2 shrink-0">
      <Button type="text" compact small onClick={handleCopyUrl}>
        <CopyIcon size="10px" />
      </Button>

      <DropdownMenu.Root closeOnItemClick={false}>
        <DropdownMenu.Trigger asChild let:builder>
          <Button type="text" compact small builders={[builder]}>
            Args
            {#if args.length > 0}
              <span
                class="inline-flex items-center justify-center w-4 h-4 text-[10px] font-medium bg-surface-active text-fg-accent rounded-full"
              >
                {args.length}
              </span>
            {/if}
            <ChevronDownIcon size="10px" />
          </Button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="end" class="w-72 p-2">
          <!-- svelte-ignore a11y-no-static-element-interactions -->
          <div
            class="args-container flex flex-col gap-y-2"
            on:keydown={handleArgsKeydown}
          >
            {#if args.length === 0}
              <p class="text-xs text-fg-muted px-1 py-2">
                No arguments. Click "Add" below.
              </p>
            {:else}
              {#each args as arg (arg.id)}
                <div class="flex items-center gap-x-1">
                  <Input
                    bind:value={arg.key}
                    placeholder="key"
                    size="sm"
                    width="100px"
                  />
                  <Input
                    bind:value={arg.value}
                    placeholder="value"
                    size="sm"
                    full
                  />
                  <Button
                    type="ghost"
                    square
                    small
                    compact
                    onClick={() => removeArg(arg.id)}
                  >
                    <Trash size="12px" />
                  </Button>
                </div>
              {/each}
            {/if}
            <Button type="text" compact small onClick={addArg}>
              <PlusIcon size="12px" />
              Add
            </Button>
          </div>
        </DropdownMenu.Content>
      </DropdownMenu.Root>

      <Button
        type="primary"
        small
        onClick={testAPI}
        disabled={isDisabled}
        loading={isLoading}
        loadingCopy="Testing"
      >
        {#if isReconciling}
          Reconciling...
        {:else}
          <PlayIcon size="12px" />
          Test API
        {/if}
      </Button>
    </div>
  </div>

  <div class="flex-1 overflow-auto">
    <APIResponsePreview
      response={apiResponse}
      error={responseError}
      {isLoading}
      {apiName}
    />
  </div>
</div>

<style lang="postcss">
  .preview-panel {
    @apply relative flex flex-col bg-surface-background border rounded-[2px] overflow-hidden;
  }
</style>
