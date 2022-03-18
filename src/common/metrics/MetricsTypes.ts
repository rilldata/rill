export interface CommonMetricsFields {
    country_code: string;
    city: string;
    locale: string;
    browser: string;
    os: string;
    device_model: string;
}
export interface MetricsEvent extends CommonMetricsFields {
    install_id: string;
    event_datetime: number;
    event_type: string;
    app_name: string;
    build_id: string;
    version: string;

    project_id: string;
    model_id: string;
}

export interface ActiveEvent extends MetricsEvent {
    event_type: "active";
    duration_sec: number;
    total_in_focus: number;
}
