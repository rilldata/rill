<script lang="ts">
  import * as Alert from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button";
  import { getRequestProjectAccessUrl } from "@rilldata/web-common/features/project/selectors.ts";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb.ts";

  export let project: Project;
  export let disabled = false;

  let open = false;

  $: requestProjectAccessUrl = getRequestProjectAccessUrl(project);
</script>

<Alert.Root bind:open>
  <Alert.Trigger asChild let:builder>
    <Button type="primary" builders={[builder]} {disabled}>Update</Button>
  </Alert.Trigger>
  <Alert.Content class="min-w-[600px]">
    <Alert.Header>
      <Alert.Title>
        Request admin access for project: {project.name}
      </Alert.Title>
      <Alert.Description>
        You don’t have permissions to update this project in Rill Cloud. To gain
        access, please request to be added as an Admin.
        <br /> <br />
        Would like to deploy a new Rill Cloud project instead? If so, remove the
        .git subdirectory or
        <a
          href="https://docs.github.com/en/get-started/getting-started-with-git/managing-remote-repositories"
          target="_blank"
          class="text-primary-600"
        >
          change the Git remote URL
        </a>. Rill Cloud only allows one cloud project to be deployed from a
        given GitHub repo.
      </Alert.Description>
    </Alert.Header>
    <Alert.Footer class="mt-5">
      <Button onClick={() => (open = false)} type="secondary">Cancel</Button>
      <Button
        onClick={() => (open = false)}
        type="primary"
        href={$requestProjectAccessUrl}
        target="_blank"
      >
        Request admin access
      </Button>
    </Alert.Footer>
  </Alert.Content>
</Alert.Root>
