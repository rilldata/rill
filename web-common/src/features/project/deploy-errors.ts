import type { ConnectError } from "@connectrpc/connect";

// rpc error: code = PermissionDenied desc = does not have permission to create assets
const RPCErrorExtractor = /rpc error: code = (.*) desc = (.*)/;

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
  return {
    noAccess: code === "PermissionDenied",
    message: desc,
  };
}
