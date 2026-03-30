<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    PlusIcon,
    PlayIcon,
    ChevronDownIcon,
    ChevronRightIcon,
    CopyIcon,
  } from "lucide-svelte";
  import { tick } from "svelte";
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
  let previewHeight = 500;
  let argsOpen = false;

  // Clear response and args when switching to a different API
  $: apiName, resetState();
  function resetState() {
    apiResponse = null;
    responseError = null;
    isLoading = false;
    args = [];
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
        // Focus the new row's key input after Svelte updates the DOM.
        // Each arg row contributes 2 <input> elements (key + value),
        // so length - 2 targets the new row's key input.
        tick().then(() => {
          const updatedInputs = Array.from(
            container.querySelectorAll<HTMLInputElement>("input"),
          );
          updatedInputs[updatedInputs.length - 2]?.focus();
        });
      }
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

<div
  class="preview-panel"
  style:height="{previewHeight}px"
  style:min-height="100px"
  style:max-height="80%"
>
  <Resizer max={500} direction="NS" side="top" bind:dimension={previewHeight} />

  <div class="flex items-center gap-x-2 px-3 py-2 border-b">
    <span class="text-xs font-medium text-fg-primary uppercase tracking-wide"
    >URL Preview: </span>
    <span class="text-[11px] font-mono text-fg-muted truncate flex-1 min-w-0"
      >{fullUrl}</span
    >
    <div class="flex items-center gap-x-2 shrink-0">
      <Tooltip distance={8}>
        <Button type="text" compact small onClick={handleCopyUrl}>
          <CopyIcon size="10px" />
        </Button>
        <TooltipContent slot="tooltip-content">Copy URL</TooltipContent>
      </Tooltip>
      <Button type="text" small compact onClick={() => (argsOpen = !argsOpen)}>
        {#if argsOpen}
          <ChevronDownIcon size="12px" />
        {:else}
          <ChevronRightIcon size="12px" />
        {/if}
        Args
        {#if args.length > 0}
          <span
            class="inline-flex items-center justify-center w-4 h-4 text-[10px] font-medium bg-primary-100 text-primary-600 rounded-full"
          >
            {args.length}
          </span>
        {/if}
      </Button>
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

  {#if argsOpen}
    <div class="border-b px-3 py-2">
      <div class="flex items-center justify-between mb-2">
        <span
          class="text-[10px] font-semibold text-fg-primary uppercase tracking-wide"
          >Query Parameters</span
        >
        <Button type="text" compact small onClick={addArg}>
          <PlusIcon size="12px" />
          Add
        </Button>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="args-container flex flex-col gap-y-2"
        onkeydown={handleArgsKeydown}
      >
        {#if args.length === 0}
          <p class="text-xs text-fg-muted py-1">
            No query parameters. Click "Add" to create one.
          </p>
        {:else}
          <div class="grid grid-cols-[1fr_1fr_28px] gap-x-2 gap-y-2">
            <span
              class="text-[10px] font-medium text-fg-muted uppercase tracking-wide"
              >Key</span
            >
            <span
              class="text-[10px] font-medium text-fg-muted uppercase tracking-wide"
              >Value</span
            >
            <span></span>
            {#each args as arg (arg.id)}
              <Input bind:value={arg.key} placeholder="key" size="sm" />
              <Input bind:value={arg.value} placeholder="value" size="sm" />
              <Button
                type="ghost"
                square
                small
                compact
                onClick={() => removeArg(arg.id)}
              >
                <Trash size="12px" />
              </Button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}

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
