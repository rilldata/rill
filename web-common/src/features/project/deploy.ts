import { page } from "$app/stores";
import {
  localServiceDeploy,
  localServiceDeployValidation,
} from "@rilldata/web-common/runtime-client/local-service";
import { get } from "svelte/store";

export async function deploy() {
  const deployValidation = await localServiceDeployValidation();
  console.log(deployValidation);
  if (!deployValidation.isAuthenticated) {
    const url = new URL(get(page).url);
    url.searchParams.set("deploying", "true");
    window.open(
      `${deployValidation.loginUrl}/?redirect=${url.toString()}`,
      "_self",
    );
  }

  if (!deployValidation.isGithubConnected) {
    window.open(`${deployValidation.githubGrantAccessUrl}`, "__target");
    return;
  }

  // await localServiceDeploy({
  //   projectName: deployValidation.localProjectName,
  //   org: deployValidation.rillUserOrgs[0],
  // });
}
