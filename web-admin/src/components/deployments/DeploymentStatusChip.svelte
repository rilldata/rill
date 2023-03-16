<script lang="ts">
  import { page } from "$app/stores";
  import { useAdminServiceGetProject, V1DeploymentStatus } from "../../client";
  import { getDeploymentStatusText } from "./status-text";

  const organizationName = $page.params.organization;
  const projectName = $page.params.project;

  const project = useAdminServiceGetProject(organizationName, projectName);
  $: deploymentStatus = $project.data?.productionDeployment?.status;

  function getDeploymentStatusClasses(status: V1DeploymentStatus) {
    switch (status) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_OK:
        return "text-white bg-green-500 hover:bg-green-600";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING:
        return "text-white bg-blue-400 hover:bg-blue-500";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR:
        return "text-white bg-red-400 hover:bg-red-500";
      default:
        return "text-white bg-gray-400 hover:bg-gray-500";
    }
  }
</script>

<a
  href={`/-/${organizationName}/${projectName}/deployment`}
  class={`inline-block px-1 py-0 rounded cursor-pointer ${getDeploymentStatusClasses(
    deploymentStatus
  )}`}
>
  {getDeploymentStatusText(deploymentStatus)}
</a>
