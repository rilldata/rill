import type { ConnectError } from "@connectrpc/connect";

// rpc error: code = PermissionDenied desc = does not have permission to create assets
const RPCErrorExtractor = /rpc error: code = (.*) desc = (.*)/;
const QuotaExceededExtractor = /quota exceeded: .* to (\d+) projects?/;

export function extractDeployError(error: ConnectError) {
  if (!error) {
    return {
      message: "",
    };
  }
  const match = RPCErrorExtractor.exec(error.message);
  if (!match) {
    if (error.message.includes("EntityTooLarge")) {
      return {
        message:
          "It looks like you have more than 10GB. Contact us to finish deployment.",
      };
    }
    return { message: error.message };
  }
  const [, code, desc] = match;

  let message = desc;

  let quotaError = false;
  const quotaErrorMatch = QuotaExceededExtractor.exec(desc);
  if (quotaErrorMatch?.length) {
    const projectQuota = Number(quotaErrorMatch[1]);
    message = `Trial plans are limited to just ${projectQuota} project${projectQuota > 1 ? "s" : ""}. Upgrade to have unlimited projects`;
    quotaError = true;
  }

  return {
    noAccess: code === "PermissionDenied",
    quotaError,
    message,
  };
}
