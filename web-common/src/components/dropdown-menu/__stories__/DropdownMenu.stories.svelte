<script lang="ts">
  import { Meta, Story } from "@storybook/addon-svelte-csf";
  import * as DropdownMenu from "..";

  import CopyIcon from "../../icons/CopyIcon.svelte";
  import DashboardIcon from "../../icons/DashboardIcon.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  let subMenuOpen = false;
</script>

<Meta title="Components/Dropdown Menu" />

<Story name="Simple">
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <Button builders={[builder]} type="secondary">Open menu</Button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.Item>Option 1</DropdownMenu.Item>
      <DropdownMenu.Item>Option 2</DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</Story>

<Story name="With icons">
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <Button builders={[builder]} type="secondary">Open menu</Button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.Item>
        <CopyIcon className="w-4 h-4 mr-2" />
        Option 1
      </DropdownMenu.Item>
      <DropdownMenu.Item>
        <DashboardIcon className="w-4 h-4 mr-2" />
        Option 2
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</Story>

<Story name="With submenu">
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <Button builders={[builder]} type="secondary">Open menu</Button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.Sub bind:open={subMenuOpen}>
        <DropdownMenu.SubTrigger
          disabled={true}
          on:click={() => (subMenuOpen = !subMenuOpen)}
        >
          Option 1
        </DropdownMenu.SubTrigger>
        <DropdownMenu.SubContent class="whitespace-nowrap">
          <DropdownMenu.Item>Submenu option 1</DropdownMenu.Item>
          <DropdownMenu.Item>Submenu option 2</DropdownMenu.Item>
        </DropdownMenu.SubContent>
      </DropdownMenu.Sub>
      <DropdownMenu.Item>Option 2</DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</Story>
