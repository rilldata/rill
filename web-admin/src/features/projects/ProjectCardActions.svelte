<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button/index.ts";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";

  let {
    organization,
    project,
    open = $bindable(false),
  }: {
    organization: string;
    project: string;
    open?: boolean;
  } = $props();
</script>

<div class="flex flex-row gap-2">
  <Button type="secondary" href={`/${organization}/${project}/#TODO`}>
    Edit
  </Button>
  <Dropdown.Root bind:open>
    <Dropdown.Trigger>
      <ThreeDot size="16px" />
    </Dropdown.Trigger>
    <Dropdown.Content class="w-48" align="start" side="right">
      <Dropdown.Item>Rename</Dropdown.Item>
      <Dropdown.Item href="/{organization}/{project}/-/dashboards?share=true">
        Share
      </Dropdown.Item>
      <Dropdown.Item>Delete</Dropdown.Item>
    </Dropdown.Content>
  </Dropdown.Root>
</div>
