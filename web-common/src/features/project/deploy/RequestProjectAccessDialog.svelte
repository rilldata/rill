<script lang="ts">
  import * as Alert from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button";
  import { getRequestProjectAccessUrl } from "@rilldata/web-common/features/project/selectors.ts";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb.ts";

  export let open: boolean;
  export let project: Project;

  $: requestProjectAccessUrl = getRequestProjectAccessUrl(project);
</script>

<Alert.Root bind:open>
  <Alert.Trigger asChild>
    <div class="hidden"></div>
  </Alert.Trigger>
  <Alert.Content class="min-w-[600px]">
    <Alert.Header>
      <Alert.Title>
        Request admin access for project: {project.name}
      </Alert.Title>
      <Alert.Description>
        You don’t have permissions to update this project in Rill Cloud. To gain
        access, please request to be added as an Admin.
      </Alert.Description>
    </Alert.Header>
    <Alert.Footer class="mt-5">
      <Button onClick={() => (open = false)} type="secondary">Cancel</Button>
      <Button
        onClick={() => (open = false)}
        type="primary"
        href={requestProjectAccessUrl}
        target="_blank"
      >
        Request admin access
      </Button>
    </Alert.Footer>
  </Alert.Content>
</Alert.Root>
