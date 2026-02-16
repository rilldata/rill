<script lang="ts">
  import { Dialog } from "@rilldata/web-common/components/dialog-v2";
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { createForm } from "svelte-forms-lib";
  import { page } from "$app/stores";
  import {
    createServiceTokenMutation,
    createProjectListQuery,
  } from "./token-queries";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Copy from "@rilldata/web-common/components/icons/Copy.svelte";

  export let open = false;
  export let organization: string;
  export let onClose: () => void;

  // Steps: "form" or "display"
  let step: "form" | "display" = "form";
  let createdToken = "";
  let copied = false;
  let apiError = "";

  const createMutation = createServiceTokenMutation(organization);
  const projectsQuery = createProjectListQuery(organization);

  // Form fields
  let name = "";
  let description = "";
  let scope: "organization" | "project" = "organization";
  let selectedProjectId = "";
  let permissions: "read" | "editor" | "admin" = "read";

  // Validation
  let nameError = "";
  let nameTouched = false;

  function validateName(value: string): string {
    if (!value || value.trim().length === 0) {
      return "Name is required";
    }
    if (value.length > 100) {
      return "Name must be 100 characters or less";
    }
    return "";
  }

  function handleNameBlur() {
    nameTouched = true;
    nameError = validateName(name);
  }

  function handleNameInput() {
    if (nameTouched) {
      nameError = validateName(name);
    }
  }

  $: isFormValid = name.trim().length > 0 && name.length <= 100 && (scope === "organization" || selectedProjectId !== "");
  $: isSubmitting = $createMutation.isLoading;

  async function handleSubmit() {
    nameTouched = true;
    nameError = validateName(name);
    if (nameError) return;

    if (scope === "project" && !selectedProjectId) {
      return;
    }

    apiError = "";

    try {
      const result = await $createMutation.mutateAsync({
        organization,
        name: name.trim(),
        description: description.trim() || undefined,
        projectId: scope === "project" ? selectedProjectId : undefined,
        permissions,
      });

      // The mutation returns the plaintext token
      createdToken = result.token || "";
      step = "display";
    } catch (err: any) {
      // Extract error message from the API response
      const message =
        err?.response?.data?.message ||
        err?.body?.message ||
        err?.message ||
        "Failed to create service token. Please try again.";
      apiError = message;
    }
  }

  async function handleCopyToken() {
    try {
      await navigator.clipboard.writeText(createdToken);
      copied = true;
      eventBus.emit("notification", {
        message: "Token copied to clipboard",
        type: "success",
      });
      // Reset copied state after 3 seconds
      setTimeout(() => {
        copied = false;
      }, 3000);
    } catch {
      // Fallback: select the text in the input for manual copying
      const input = document.getElementById("created-token-input");
      if (input instanceof HTMLInputElement) {
        input.select();
      }
    }
  }

  function handleClose() {
    resetForm();
    onClose();
  }

  function resetForm() {
    step = "form";
    name = "";
    description = "";
    scope = "organization";
    selectedProjectId = "";
    permissions = "read";
    nameError = "";
    nameTouched = false;
    createdToken = "";
    copied = false;
    apiError = "";
  }

  // Reset form when dialog opens
  $: if (open) {
    resetForm();
  }

  const permissionOptions = [
    {
      value: "read" as const,
      label: "Read",
      description: "Can view dashboards and data sources",
    },
    {
      value: "editor" as const,
      label: "Editor",
      description: "Can edit projects, models, and dashboards",
    },
    {
      value: "admin" as const,
      label: "Admin",
      description: "Full access including member and settings management",
    },
  ];
</script>

<Dialog {open} on:close={handleClose}>
  <svelte:fragment slot="title">
    {#if step === "form"}
      Create Service Token
    {:else}
      Service Token Created
    {/if}
  </svelte:fragment>

  <svelte:fragment slot="body">
    {#if step === "form"}
      <form
        on:submit|preventDefault={handleSubmit}
        class="flex flex-col gap-4"
        id="create-service-token-form"
      >
        <!-- Name -->
        <div class="flex flex-col gap-1.5">
          <label for="token-name" class="text-sm font-medium text-gray-700">
            Name <span class="text-red-500">*</span>
          </label>
          <input
            id="token-name"
            type="text"
            bind:value={name}
            on:blur={handleNameBlur}
            on:input={handleNameInput}
            placeholder="e.g., CI/CD Pipeline Token"
            maxlength={100}
            class="w-full rounded-sm border px-3 py-2 text-sm outline-none transition-colors
              {nameError && nameTouched
                ? 'border-red-400 focus:border-red-500'
                : 'border-gray-300 focus:border-primary-500'}"
          />
          {#if nameError && nameTouched}
            <p class="text-xs text-red-500">{nameError}</p>
          {:else}
            <p class="text-xs text-gray-400">{name.length}/100 characters</p>
          {/if}
        </div>

        <!-- Description -->
        <div class="flex flex-col gap-1.5">
          <label
            for="token-description"
            class="text-sm font-medium text-gray-700"
          >
            Description
            <span class="text-xs font-normal text-gray-400">(optional)</span>
          </label>
          <textarea
            id="token-description"
            bind:value={description}
            placeholder="What is this token used for?"
            rows={2}
            class="w-full rounded-sm border border-gray-300 px-3 py-2 text-sm outline-none transition-colors focus:border-primary-500 resize-none"
          ></textarea>
        </div>

        <!-- Scope -->
        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-gray-700">Scope</label>
          <div class="flex flex-col gap-2">
            <label
              class="flex items-center gap-2 rounded-sm border px-3 py-2.5 cursor-pointer transition-colors
                {scope === 'organization'
                  ? 'border-primary-500 bg-primary-50'
                  : 'border-gray-200 hover:border-gray-300'}"
            >
              <input
                type="radio"
                bind:group={scope}
                value="organization"
                class="accent-primary-500"
              />
              <div>
                <span class="text-sm font-medium">Organization</span>
                <p class="text-xs text-gray-500">
                  Access to all projects in the organization
                </p>
              </div>
            </label>

            <label
              class="flex items-center gap-2 rounded-sm border px-3 py-2.5 cursor-pointer transition-colors
                {scope === 'project'
                  ? 'border-primary-500 bg-primary-50'
                  : 'border-gray-200 hover:border-gray-300'}"
            >
              <input
                type="radio"
                bind:group={scope}
                value="project"
                class="accent-primary-500"
              />
              <div>
                <span class="text-sm font-medium">Project</span>
                <p class="text-xs text-gray-500">
                  Access to a specific project only
                </p>
              </div>
            </label>

            {#if scope === "project"}
              <div class="ml-6">
                <select
                  bind:value={selectedProjectId}
                  class="w-full rounded-sm border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500"
                >
                  <option value="" disabled>Select a project...</option>
                  {#if $projectsQuery.isSuccess && $projectsQuery.data}
                    {#each $projectsQuery.data as project}
                      <option value={project.id}>{project.name}</option>
                    {/each}
                  {/if}
                </select>
                {#if $projectsQuery.isLoading}
                  <p class="mt-1 text-xs text-gray-400">
                    Loading projects...
                  </p>
                {/if}
                {#if scope === "project" && !selectedProjectId}
                  <p class="mt-1 text-xs text-gray-400">
                    Please select a project
                  </p>
                {/if}
              </div>
            {/if}
          </div>
        </div>

        <!-- Permissions -->
        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-gray-700">Permissions</label>
          <div class="flex flex-col gap-2">
            {#each permissionOptions as option}
              <label
                class="flex items-start gap-2 rounded-sm border px-3 py-2.5 cursor-pointer transition-colors
                  {permissions === option.value
                    ? 'border-primary-500 bg-primary-50'
                    : 'border-gray-200 hover:border-gray-300'}"
              >
                <input
                  type="radio"
                  bind:group={permissions}
                  value={option.value}
                  class="accent-primary-500 mt-0.5"
                />
                <div>
                  <span class="text-sm font-medium">{option.label}</span>
                  <p class="text-xs text-gray-500">{option.description}</p>
                </div>
              </label>
            {/each}
          </div>
        </div>

        <!-- API Error -->
        {#if apiError}
          <div
            class="rounded-sm border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700"
          >
            {apiError}
          </div>
        {/if}
      </form>
    {:else}
      <!-- Step 2: Token Display -->
      <div class="flex flex-col items-center gap-4">
        <div
          class="flex h-12 w-12 items-center justify-center rounded-full bg-green-100"
        >
          <svg
            class="h-6 w-6 text-green-600"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M5 13l4 4L19 7"
            />
          </svg>
        </div>

        <div class="text-center">
          <h3 class="text-base font-semibold text-gray-900">
            Token Created Successfully
          </h3>
          <p class="mt-1 text-sm text-gray-500">
            Your service token <strong>{name}</strong> has been created.
          </p>
        </div>

        <!-- Token Value -->
        <div class="w-full">
          <label
            for="created-token-input"
            class="mb-1.5 block text-sm font-medium text-gray-700"
          >
            Your Token
          </label>
          <div class="flex gap-2">
            <input
              id="created-token-input"
              type="text"
              readonly
              value={createdToken}
              class="flex-1 rounded-sm border border-gray-300 bg-gray-50 px-3 py-2 font-mono text-sm select-all"
            />
            <button
              type="button"
              on:click={handleCopyToken}
              class="inline-flex items-center gap-1.5 rounded-sm border px-3 py-2 text-sm font-medium transition-colors
                {copied
                  ? 'border-green-300 bg-green-50 text-green-700'
                  : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'}"
            >
              {#if copied}
                <svg
                  class="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                Copied
              {:else}
                <svg
                  class="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                  />
                </svg>
                Copy
              {/if}
            </button>
          </div>
        </div>

        <!-- Warning Banner -->
        <div
          class="flex w-full items-start gap-2 rounded-sm border border-amber-200 bg-amber-50 px-3 py-2.5"
        >
          <svg
            class="mt-0.5 h-4 w-4 flex-shrink-0 text-amber-600"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"
            />
          </svg>
          <p class="text-sm text-amber-800">
            <strong>Make sure to copy your token now.</strong> You won't be able
            to see it again!
          </p>
        </div>
      </div>
    {/if}
  </svelte:fragment>

  <svelte:fragment slot="footer">
    {#if step === "form"}
      <div class="flex justify-end gap-2">
        <Button type="secondary" on:click={handleClose}>Cancel</Button>
        <Button
          type="primary"
          disabled={!isFormValid || isSubmitting}
          on:click={handleSubmit}
        >
          {#if isSubmitting}
            <svg
              class="mr-1.5 h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            Creating...
          {:else}
            Create Token
          {/if}
        </Button>
      </div>
    {:else}
      <div class="flex justify-end">
        <Button type="primary" on:click={handleClose}>Done</Button>
      </div>
    {/if}
  </svelte:fragment>
</Dialog>