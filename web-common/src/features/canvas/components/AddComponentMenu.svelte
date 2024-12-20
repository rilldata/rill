<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { CanvasComponentType } from "@rilldata/web-common/features/canvas/components/types";
  import { ChevronDown, Plus } from "lucide-svelte";

  export let addComponent: (componentName: CanvasComponentType) => void;

  let open = false;

  const menuItems: {
    id: CanvasComponentType;
    label: string;
  }[] = [
    { id: "markdown", label: "Text" },
    { id: "kpi", label: "KPI" },
    { id: "image", label: "Image" },
    { id: "bar_chart", label: "Chart" },
    { id: "table", label: "Table" },
  ];
</script>

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <Button builders={[builder]} type="secondary">
      <Plus class="flex items-center justify-center" size="16px" />
      <div class="flex gap-x-1 items-center">
        Add component
        <ChevronDown size="14px" />
      </div>
    </Button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content class="flex flex-col gap-y-1 ">
    <DropdownMenu.Group>
      {#each menuItems as item}
        <DropdownMenu.Item on:click={() => addComponent(item.id)}>
          {item.label}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Group>
  </DropdownMenu.Content>
</DropdownMenu.Root>
