<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { Code2Icon, Database, Folder, PlusCircleIcon } from "lucide-svelte";
  import ApiIcon from "../../components/icons/APIIcon.svelte";
  import AlertCircleOutline from "../../components/icons/AlertCircleOutline.svelte";
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import Chart from "../../components/icons/Chart.svelte";
  import DashboardIcon from "../../components/icons/DashboardIcon.svelte";
  import File from "../../components/icons/File.svelte";
  import ReportIcon from "../../components/icons/ReportIcon.svelte";
  import ThemeIcon from "../../components/icons/ThemeIcon.svelte";

  function handleAddSource() {
    console.log("Add source");
  }

  function handleAddModel() {
    console.log("Add model");
  }

  function handleAddDashboard() {
    console.log("Add dashboard");
  }

  function handleAddFolder() {
    console.log("Add folder");
  }

  function handleAddBlankFile() {
    console.log("Add blank file");
  }

  function handleAddAPI() {
    console.log("Add API");
  }

  function handleAddChart() {
    console.log("Add chart");
  }

  function handleAddTheme() {
    console.log("Add theme");
  }

  function handleAddReport() {
    console.log("Add report");
  }

  function handleAddAlert() {
    console.log("Add alert");
  }
</script>

<div class="p-2">
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <button
        {...builder}
        use:builder.action
        class:open
        class="p-2 bg-primary-50 hover:bg-primary-100 text-primary-700 hover:text-primary-800 w-full flex gap-x-2 items-center font-medium h-7 rounded-sm justify-center"
      >
        <PlusCircleIcon size="14px" />
        <div class="flex gap-x-1 items-center">
          Add
          <CaretDownIcon size="10px" />
        </div>
      </button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="w-[240px]">
      <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddSource}>
        <Database size="16px" className="text-gray-900" />
        Source
      </DropdownMenu.Item>
      <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddModel}>
        <Code2Icon size="16px" className="text-gray-900" />
        Model
      </DropdownMenu.Item>
      <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddDashboard}>
        <DashboardIcon size="16px" className="text-gray-900" />
        Dashboard
      </DropdownMenu.Item>
      <DropdownMenu.Sub>
        <DropdownMenu.SubTrigger>More</DropdownMenu.SubTrigger>
        <DropdownMenu.SubContent class="w-[240px]">
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddFolder}>
            <Folder size="16px" /> Folder
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddBlankFile}>
            <File size="16px" /> Blank file
          </DropdownMenu.Item>
          <DropdownMenu.Separator />
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddAPI}>
            <ApiIcon size="16px" /> API
          </DropdownMenu.Item>
          <DropdownMenu.Separator />
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddChart}>
            <Chart size="16px" /> Chart
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddTheme}>
            <ThemeIcon size="16px" /> Theme
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddReport}>
            <ReportIcon size="16px" /> Report
          </DropdownMenu.Item>
          <DropdownMenu.Item class="flex gap-x-2" on:click={handleAddAlert}>
            <AlertCircleOutline size="16px" /> Alert
          </DropdownMenu.Item>
        </DropdownMenu.SubContent>
      </DropdownMenu.Sub>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>
