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
}

export interface ErrorEvent extends MetricsEvent {
  action: ErrorEventAction;
  error_code: SourceErrorCodes;
  space: MetricsEventSpace;
  screen_name: MetricsEventScreenName;
  file_type: SourceFileType;
  connection_type: SourceConnectionType;
  glob: boolean;
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
    glob: boolean
  ): ErrorEvent {
    const event = this.getBaseMetricsEvent(
      "error",
      commonFields,
      commonUserFields
    ) as ErrorEvent;
    event.action = ErrorEventAction.SourceError;
    event.space = space;
    event.screen_name = screen_name;
    event.error_code = error_code;
    event.connection_type = connection_type;
    event.file_type = file_type;
    event.glob = glob;
    return event;
  }
}
