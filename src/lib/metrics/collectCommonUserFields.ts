import UAParser from "ua-parser-js";
import type { CommonUserFields } from "$common/metrics/MetricsTypes";

export async function collectCommonUserFields(): Promise<CommonUserFields> {
    const parser = new UAParser();
    const result = parser.getResult();
    return {
        country_code: "",
        city: "",
        locale: "",
        browser: result.browser.name,
        os: result.os.name,
        device_model: result.os.model,
    };
}
