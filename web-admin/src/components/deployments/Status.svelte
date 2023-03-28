<script lang="ts">
  import type { V1DeploymentStatus } from "../../client";
  import { getDeploymentStatusText } from "./status-text";

  export let status: V1DeploymentStatus;
</script>

<h1 class="text-2xl font-medium text-gray-800 mb-4">
  Deployment status: {getDeploymentStatusText(status)}
</h1>
