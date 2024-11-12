import type { ConnectError } from "@connectrpc/connect";

// rpc error: code = PermissionDenied desc = does not have permission to create assets
const RPCErrorExtractor = /rpc error: code = (.*) desc = (.*)/;
const ProjectQuotaErrorMatcher =
  /quota exceeded: org .* is limited to (\d+) projects/;
const OrgQuotaErrorMatcher =
  /(quota exceeded: you can only create .* single-user orgs|trial orgs quota exceeded)/;
const TrialEndedMatcher = /trial has ended/;
const SubEndedMatcher = /subscription cancelled/;

export enum DeployErrorType {
  Unknown,
  PermissionDenied,
  LargeProject,
  ProjectLimitHit,
  OrgLimitHit,
  SubscriptionEnded,
}

export type DeployError = {
  type: DeployErrorType;
  title: string;
  message: string;
};

export function extractDeployError(error: ConnectError): DeployError {
  if (!error) {
    return {
      type: DeployErrorType.Unknown,
      title: "",
      message: "",
    };
  }
  const title = "Oops! An error occurred";

  const match = RPCErrorExtractor.exec(error.message);
  if (!match) {
    if (error.message.includes("EntityTooLarge")) {
      return {
        type: DeployErrorType.LargeProject,
        title,
        message:
          "It looks like you have more than 10GB. Contact us to finish deployment.",
      };
    }
    return {
      type: DeployErrorType.Unknown,
      title,
      message: error.message,
    };
  }
  const [, code, message] = match;

  if (code === "PermissionDenied") {
    return {
      type: DeployErrorType.PermissionDenied,
      title,
      message,
    };
  }

  const projectQuotaMatch = ProjectQuotaErrorMatcher.exec(message);
  if (projectQuotaMatch?.length) {
    const projectQuota = Number(projectQuotaMatch[1]);
    return {
      type: DeployErrorType.ProjectLimitHit,
      title: "To deploy this project, start a Team plan",
      message: `Your trial plan is limited to ${projectQuota} project${projectQuota > 1 ? "s" : ""}. To have unlimited projects, upgrade to a Team plan.`,
    };
  }

  if (OrgQuotaErrorMatcher.test(message)) {
    return {
      type: DeployErrorType.OrgLimitHit,
      title: "To deploy to more organizations, start a Team plan",
      message: "",
    };
  }

  if (SubEndedMatcher.test(message) || TrialEndedMatcher.test(message)) {
    return {
      type: DeployErrorType.SubscriptionEnded,
      title: "To deploy this project, start a Team plan",
      message: "",
    };
  }

  return {
    type: DeployErrorType.Unknown,
    title,
    message,
  };
}
