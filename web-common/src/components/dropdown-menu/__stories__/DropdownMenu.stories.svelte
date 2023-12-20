<script lang="ts">
  import { Meta, Story } from "@storybook/addon-svelte-csf";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSub,
    DropdownMenuSubContent,
    DropdownMenuSubTrigger,
    DropdownMenuTrigger,
  } from "..";
  import { Button } from "../../button";
  import CopyIcon from "../../icons/CopyIcon.svelte";
  import DashboardIcon from "../../icons/DashboardIcon.svelte";
</script>

<Meta title="Components/Dropdown Menu" />

<Story name="Simple">
  <DropdownMenu>
    <DropdownMenuTrigger>
      <Button type="secondary">Open menu</Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start">
      <DropdownMenuItem>Option 1</DropdownMenuItem>
      <DropdownMenuItem>Option 2</DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</Story>

<Story name="With icons">
  <DropdownMenu>
    <DropdownMenuTrigger>
      <Button type="secondary">Open menu</Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start">
      <DropdownMenuItem>
        <CopyIcon className="w-4 h-4 mr-2" />
        Option 1
      </DropdownMenuItem>
      <DropdownMenuItem>
        <DashboardIcon className="w-4 h-4 mr-2" />
        Option 2
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</Story>

<Story name="With submenu">
  <DropdownMenu>
    <DropdownMenuTrigger>
      <Button type="secondary">Open menu</Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start">
      <DropdownMenuSub>
        <DropdownMenuSubTrigger>Option 1</DropdownMenuSubTrigger>
        <DropdownMenuSubContent class="whitespace-nowrap">
          <DropdownMenuItem>Submenu option 1</DropdownMenuItem>
          <DropdownMenuItem>Submenu option 2</DropdownMenuItem>
        </DropdownMenuSubContent>
      </DropdownMenuSub>
      <DropdownMenuItem>Option 2</DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</Story>
