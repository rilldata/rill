<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ChevronDown } from "lucide-svelte";
  import type { FormTemplate } from "./schemas/types";

  export let templates: FormTemplate[];
  export let onSelectTemplate: (template: FormTemplate) => void;

  let open = false;
</script>

{#if templates.length > 0}
  <div class="mb-4">
    <DropdownMenu.Root bind:open>
      <DropdownMenu.Trigger asChild let:builder>
        <button
          {...builder}
          use:builder.action
          class="inline-flex items-center gap-1.5 text-sm font-medium text-primary-500 hover:text-primary-600"
        >
          Use a template
          <ChevronDown class="h-4 w-4" />
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-64">
        {#each templates as template (template.id)}
          <DropdownMenu.Item
            on:click={() => {
              onSelectTemplate(template);
              open = false;
            }}
          >
            <div class="flex flex-col gap-0.5">
              <span class="font-medium">{template.label}</span>
              {#if template.description}
                <span class="text-xs text-gray-500">{template.description}</span>
              {/if}
            </div>
          </DropdownMenu.Item>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </div>
{/if}
