import { page } from "$app/stores";
import type { PartialMessage } from "@bufbuild/protobuf";
import type { ConnectError } from "@connectrpc/connect";
import { createMutation, CreateMutationOptions } from "@rilldata/svelte-query";
import { PushToGithubRequest } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import {
  localServiceDeploy,
  localServiceDeployValidation,
} from "@rilldata/web-common/runtime-client/local-service";
import { get } from "svelte/store";

export function createDeployer(options?: {
  mutation?: CreateMutationOptions<
    Awaited<ReturnType<typeof deploy>>,
    ConnectError,
    PartialMessage<PushToGithubRequest>,
    unknown
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof deploy>>,
    unknown,
    PartialMessage<PushToGithubRequest>,
    unknown
  >(deploy, mutationOptions);
}

async function deploy() {
  const deployValidation = await localServiceDeployValidation();
  if (!deployValidation.isAuthenticated) {
    const url = new URL(get(page).url);
    url.searchParams.set("deploying", "true");
    window.open(
      `${deployValidation.loginUrl}/?redirect=${url.toString()}`,
      "_self",
    );
  }

  if (deployValidation.isGithubRepo && !deployValidation.isGithubConnected) {
    window.open(`${deployValidation.githubGrantAccessUrl}`, "__target");
    return;
  }

  const resp = await localServiceDeploy({
    projectName: deployValidation.localProjectName,
    org: deployValidation.rillUserOrgs[0],
    upload: !deployValidation.isGithubRepo,
  });
  if (resp.frontendUrl) {
    window.open(resp.frontendUrl, "_self");
  }
}
