<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { ChevronDown, Plus } from "lucide-svelte";
  import { getNameFromFile } from "../../entity-management/entity-mappers";
  import { createResourceFile } from "../../file-explorer/new-files";

  export let addComponent: (componentName: string) => void;

  let open = false;

  async function handleAddComponent() {
    const newFilePath = await createResourceFile(ResourceKind.Component);

    if (!newFilePath) return;

    const componentName = getNameFromFile(newFilePath);

    if (componentName) {
      addComponent(componentName);
    }
  }
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
      <DropdownMenu.Item on:click={handleAddComponent}>
        Create new component
      </DropdownMenu.Item>
    </DropdownMenu.Group>
  </DropdownMenu.Content>
</DropdownMenu.Root>
