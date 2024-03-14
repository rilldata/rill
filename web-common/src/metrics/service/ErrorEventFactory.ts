import { MetricsEventFactory } from "./MetricsEventFactory";
import type {
  CommonFields,
  CommonUserFields,
  MetricsEvent,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "./MetricsTypes";
import type {
  SourceConnectionType,
  SourceErrorCodes,
  SourceFileType,
} from "./SourceEventTypes";

export enum ErrorEventAction {
  SourceError = "source-error",
  ErrorBoundary = "error-boundary",
}

export interface SourceErrorEvent extends MetricsEvent {
  action: ErrorEventAction;
  error_code: SourceErrorCodes;
  space: MetricsEventSpace;
  screen_name: MetricsEventScreenName;
  file_type: SourceFileType;
  connection_type: SourceConnectionType;
  glob: boolean;
}

export interface HTTPErrorEvent extends MetricsEvent {
  action: ErrorEventAction;
  screen_name: MetricsEventScreenName;
  api: string;
  status: string;
  message: string;
}

export interface JavascriptErrorEvent extends MetricsEvent {
  action: ErrorEventAction;
  screen_name: MetricsEventScreenName;
  stack: string;
  message: string;
}

export class ErrorEventFactory extends MetricsEventFactory {
  public sourceErrorEvent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    space: MetricsEventSpace,
    screen_name: MetricsEventScreenName,
    error_code: SourceErrorCodes,
    connection_type: SourceConnectionType,
    file_type: SourceFileType,
    glob: boolean,
  ): SourceErrorEvent {
    const event = this.getBaseMetricsEvent(
      "error",
      ErrorEventAction.SourceError,
      commonFields,
      commonUserFields,
    ) as SourceErrorEvent;
    event.action = ErrorEventAction.SourceError;
    event.space = space;
    event.screen_name = screen_name;
    event.error_code = error_code;
    event.connection_type = connection_type;
    event.file_type = file_type;
    event.glob = glob;
    return event;
  }

  public httpErrorEvent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    screen_name: MetricsEventScreenName,
    api: string,
    status: string,
    message: string,
  ) {
    const event = this.getBaseMetricsEvent(
      "error",
      ErrorEventAction.ErrorBoundary,
      commonFields,
      commonUserFields,
    ) as HTTPErrorEvent;
    event.action = ErrorEventAction.ErrorBoundary;
    event.screen_name = screen_name;
    event.api = api;
    event.status = status;
    event.message = message;
    return event;
  }

  public javascriptErrorEvent(
    commonFields: CommonFields,
    commonUserFields: CommonUserFields,
    screen_name: MetricsEventScreenName,
    stack: string,
    message: string,
  ) {
    const event = this.getBaseMetricsEvent(
      "error",
      ErrorEventAction.ErrorBoundary,
      commonFields,
      commonUserFields,
    ) as JavascriptErrorEvent;
    event.action = ErrorEventAction.ErrorBoundary;
    event.screen_name = screen_name;
    event.stack = stack;
    event.message = message;
    return event;
  }
}
