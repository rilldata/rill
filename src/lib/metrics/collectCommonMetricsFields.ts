import type { CommonMetricsFields } from "$common/metrics/MetricsTypes";
import UAParser from "ua-parser-js";

export async function collectCommonMetricsFields(): Promise<CommonMetricsFields> {
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
