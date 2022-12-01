import UAParser from "ua-parser-js";
import type { CommonUserFields } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";

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
