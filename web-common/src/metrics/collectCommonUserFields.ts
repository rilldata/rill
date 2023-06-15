import type { CommonUserFields } from "@rilldata/web-common/metrics/service/MetricsTypes";
import UAParser from "ua-parser-js";

export async function collectCommonUserFields(): Promise<CommonUserFields> {
  const parser = new UAParser();
  const result = parser.getResult();
  return {
    locale: navigator.language,
    browser: result.browser.name,
    os: result.os.name,
    device_model: result.os.model,
  };
}
