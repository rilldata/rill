import UAParser from "ua-parser-js";
import type { CommonUserFields } from "$common/metrics-service/MetricsTypes";

export async function collectCommonUserFields(): Promise<CommonUserFields> {
    const parser = new UAParser();
    const result = parser.getResult();
    let ipInfo;
    try {
        ipInfo = await (await fetch("https://ipapi.co/json/")).json();
    } catch (err) {
        ipInfo = {
            city: "",
            country_code: "",
            languages: ",",
        }
    }
    return {
        country_code: ipInfo.country_code,
        city: ipInfo.city,
        locale: ipInfo.languages.split(",")[0],
        browser: result.browser.name,
        os: result.os.name,
        device_model: result.os.model,
    };
}
