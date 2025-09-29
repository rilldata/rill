// rpc error: code = PermissionDenied desc = does not have permission to create assets
const RPCErrorExtractor = /rpc error: code = (.*) desc = (.*)/;
const ProjectQuotaErrorMatcher =
  /quota exceeded: org .* is limited to (\d+) projects/;
const OrgQuotaErrorMatcher =
  /(quota exceeded: you can only create .* single-user orgs|trial orgs quota exceeded)/;
const TrialEndedMatcher = /trial has ended/;
const SubEndedMatcher = /subscription cancelled/;
export const GithubRepoNoAccessError = "GitNoAccessError";

export enum DeployErrorType {
  Unknown,
  PermissionDenied,
  LargeProject,
  ProjectLimitHit,
  OrgLimitHit,
  TrialEnded,
  SubscriptionEnded,
  GithubNoAccess,
}
const ErrorMessageVariants = {
  [DeployErrorType.OrgLimitHit]: {
    title: "To deploy to more organizations, start a Team plan",
    message: "",
  },
  [DeployErrorType.TrialEnded]: {
    title: "To deploy this project, start a Team plan",
    message:
      "Your trial has ended. In order to deploy this project, start a Team plan.",
  },
  [DeployErrorType.SubscriptionEnded]: {
    title: "To deploy this project, start a Team plan",
    message:
      "Your subscription has ended. In order to deploy this project, renew Team plan.",
  },
};

export type DeployError = {
  type: DeployErrorType;
  title: string;
  message: string;
};

export function getPrettyDeployError(
  error: Error,
  orgOnTrial: boolean,
): DeployError {
  if (!error) {
    return {
      type: DeployErrorType.Unknown,
      title: "",
      message: "",
    };
  }
  let title = "Oops! An error occurred";

  if (error.message === GithubRepoNoAccessError) {
    return {
      type: DeployErrorType.GithubNoAccess,
      title,
      message: "Failed to get access to the repo. Please try again.",
    };
  }

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
  const [, code, desc] = match;
  let message = desc;

  if (code === "PermissionDenied") {
    return {
      type: DeployErrorType.PermissionDenied,
      title,
      message,
    };
  }

  const projectQuotaMatch = ProjectQuotaErrorMatcher.exec(desc);
  if (projectQuotaMatch?.length) {
    const projectQuota = Number(projectQuotaMatch[1]);
    const planLabel = orgOnTrial ? "current plan" : "trial plan";

    return {
      type: DeployErrorType.ProjectLimitHit,
      title: "To deploy this project, start a Team plan",
      message: `Your ${planLabel} is limited to ${projectQuota} project${projectQuota > 1 ? "s" : ""}. To have unlimited projects, upgrade to a Team plan.`,
    };
  }

  let type = DeployErrorType.Unknown;

  switch (true) {
    case OrgQuotaErrorMatcher.test(desc):
      type = DeployErrorType.OrgLimitHit;
      break;

    case TrialEndedMatcher.test(desc):
      type = DeployErrorType.TrialEnded;
      break;

    case SubEndedMatcher.test(desc):
      type = DeployErrorType.SubscriptionEnded;
      break;
  }

  if (type in ErrorMessageVariants) {
    title = ErrorMessageVariants[type].title;
    message = ErrorMessageVariants[type].message;
  }

  return {
    type,
    title,
    message,
  };
}
