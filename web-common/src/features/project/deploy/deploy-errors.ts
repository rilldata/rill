// rpc error: code = PermissionDenied desc = does not have permission to create assets
const RPCErrorExtractor = /rpc error: code = (.*) desc = (.*)/;
const ProjectQuotaErrorMatcher =
  /quota exceeded: org .* is limited to (\d+) projects/;
const OrgQuotaErrorMatcher =
  /(quota exceeded: you can only create .* single-user orgs|trial orgs quota exceeded)/;
const TrialEndedMatcher = /trial has ended/;
const TrialCreditsDepleted = /trial credits depleted/;
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
    title: "To deploy to another organization, choose a plan",
    message:
      "Your trial plan supports a single organization. Choose a plan to deploy to more.",
  },
  [DeployErrorType.TrialEnded]: {
    title: "To deploy this project, choose a plan",
    message: "Your trial has ended. Choose a plan to deploy this project.",
  },
  [DeployErrorType.SubscriptionEnded]: {
    title: "To deploy this project, renew your plan",
    message:
      "Your subscription has ended. Renew your plan to deploy this project.",
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
  const title = "Oops! An error occurred";

  if (error.message === GithubRepoNoAccessError) {
    return {
      type: DeployErrorType.GithubNoAccess,
      title,
      message: "Failed to get access to the repo. Please try again.",
    };
  }

  const match = RPCErrorExtractor.exec(error.message);
  if (!match) {
    return parseDeployErrorMessage(error.message, orgOnTrial);
  }
  const [, code, desc] = match;

  if (code === "PermissionDenied") {
    return {
      type: DeployErrorType.PermissionDenied,
      title,
      message: desc,
    };
  }

  return parseDeployErrorMessage(desc, orgOnTrial);
}

export function isQuotaDeployError(deployError: DeployError) {
  return (
    deployError.type === DeployErrorType.ProjectLimitHit ||
    deployError.type === DeployErrorType.OrgLimitHit ||
    deployError.type === DeployErrorType.TrialEnded ||
    deployError.type === DeployErrorType.SubscriptionEnded
  );
}

function parseDeployErrorMessage(message: string, orgOnTrial: boolean) {
  let title = "Oops! An error occurred";

  if (message.includes("EntityTooLarge")) {
    return {
      type: DeployErrorType.LargeProject,
      title,
      message:
        "It looks like you have more than 10GB. Contact us to finish deployment.",
    };
  }

  const projectQuotaMatch = ProjectQuotaErrorMatcher.exec(message);
  if (projectQuotaMatch?.length) {
    const projectQuota = Number(projectQuotaMatch[1]);
    const planLabel = orgOnTrial ? "trial plan" : "current plan";

    return {
      type: DeployErrorType.ProjectLimitHit,
      title: "To deploy more projects, upgrade your plan",
      message: `Your ${planLabel} is limited to ${projectQuota} project${projectQuota > 1 ? "s" : ""}. Upgrade your plan to deploy more, or contact us about unlimited projects.`,
    };
  }

  let type = DeployErrorType.Unknown;

  switch (true) {
    case OrgQuotaErrorMatcher.test(message):
      type = DeployErrorType.OrgLimitHit;
      break;

    case TrialEndedMatcher.test(message):
    case TrialCreditsDepleted.test(message):
      type = DeployErrorType.TrialEnded;
      break;

    case SubEndedMatcher.test(message):
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
