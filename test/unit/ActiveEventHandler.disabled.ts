/**
 * @jest-environment jsdom
 */

import { TestBase } from "@adityahegde/typescript-test-utils";
import { JestTestLibrary } from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import type { SinonStub } from "sinon";
import { assert, useFakeTimers } from "sinon";
import { RillIntakeClient } from "$common/metrics-service/RillIntakeClient";
import { RootConfig } from "$common/config/RootConfig";
import { ActiveEventHandler } from "$lib/metrics/ActiveEventHandler";
import {
  dataModelerStateServiceFactory,
  metricsServiceFactory,
} from "$server/serverFactory";
import { asyncWait } from "$common/utils/waitUtils";
import type { ActiveEvent } from "$common/metrics-service/MetricsTypes";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";

const CommonUserMetricsData = {
  browser: "chrome",
  device_model: "mac",
  os: "macos",
  locale: "en-US",
};

// Disabling this since testing this will need a complete server
// We need a better framework for it.
// TODO
// @TestBase.Suite
@TestBase.TestLibrary(JestTestLibrary)
export class ActiveEventHandlerDisabled extends TestBase {
  private windowListenerStub: SinonStub;
  private rillIntakeStub: SinonStub;
  private config: RootConfig;

  @TestBase.BeforeSuite()
  public setupSuite() {
    this.config = new RootConfig({});
    this.windowListenerStub = this.sandbox.stub(window, "addEventListener");
    this.rillIntakeStub = this.sandbox.stub(
      RillIntakeClient.prototype,
      "fireEvent"
    );
    this.sandbox.useFakeServer();
  }

  @TestBase.Test()
  public async shouldDelayTimer() {
    const { dataModelerStateService, focusCallback } = this.initTest(30 * 1000);

    this.sandbox.clock.tick(30 * 1000);
    focusCallback();
    this.sandbox.clock.tick(35 * 1000);
    assert.notCalled(this.rillIntakeStub);

    this.sandbox.clock.tick(30 * 1000);
    this.sandbox.clock.restore();
    await asyncWait(100);
    assert.calledOnce(this.rillIntakeStub);

    expect(this.rillIntakeStub.firstCall.args[0]).toEqual(
      this.getExpectedActiveEvent(dataModelerStateService, 120000, "", 60, 1)
    );
  }

  @TestBase.Test()
  public async shouldHandleBlurAndFocusMultipleTimes() {
    const { dataModelerStateService, blurCallback, focusCallback } =
      this.initTest();

    this.sandbox.clock.tick(15 * 1000);
    focusCallback();
    this.sandbox.clock.tick(15 * 1000);
    blurCallback();
    this.sandbox.clock.tick(15 * 1000);
    focusCallback();
    this.sandbox.clock.tick(20 * 1000);
    this.sandbox.clock.restore();
    await asyncWait(100);

    expect(this.rillIntakeStub.firstCall.args[0]).toEqual(
      this.getExpectedActiveEvent(dataModelerStateService, 120999, "", 31, 2)
    );
  }

  @TestBase.Test()
  public async shouldNotFireIfNeverFocused() {
    const { blurCallback } = this.initTest();

    this.sandbox.clock.tick(15 * 1000);
    blurCallback();
    this.sandbox.clock.tick(50 * 1000);
    this.sandbox.clock.restore();
    await asyncWait(100);

    assert.notCalled(this.rillIntakeStub);
  }

  @TestBase.Test()
  public async shouldFireMultipleTimes() {
    const { blurCallback, focusCallback, dataModelerStateService } =
      this.initTest();

    this.sandbox.clock.tick(30 * 1000);
    focusCallback();
    this.sandbox.clock.tick(60 * 1000);
    blurCallback();
    this.sandbox.clock.tick(60 * 1000);
    focusCallback();
    this.sandbox.clock.tick(60 * 1000);
    this.sandbox.clock.restore();
    await asyncWait(100);

    assert.calledTwice(this.rillIntakeStub);
    expect(this.rillIntakeStub.firstCall.args[0]).toEqual(
      this.getExpectedActiveEvent(dataModelerStateService, 120999, "", 31, 1)
    );
    expect(this.rillIntakeStub.secondCall.args[0]).toEqual(
      this.getExpectedActiveEvent(dataModelerStateService, 240999, "", 31, 1)
    );
  }

  private initTest(initialTime = 59999) {
    this.sandbox.clock = useFakeTimers();
    this.sandbox.clock.tick(initialTime);

    const { dataModelerStateService, activeEventHandler } =
      this.createActiveEventHandler();
    const blurCallback = this.windowListenerStub.args[0][1];
    const focusCallback = this.windowListenerStub.args[1][1];

    return {
      dataModelerStateService,
      activeEventHandler,
      blurCallback,
      focusCallback,
    };
  }

  private createActiveEventHandler() {
    const dataModelerStateService = dataModelerStateServiceFactory(this.config);
    const activeEventHandler = new ActiveEventHandler(
      this.config,
      metricsServiceFactory(this.config, dataModelerStateService),
      CommonUserMetricsData
    );
    return { dataModelerStateService, activeEventHandler };
  }

  private getExpectedActiveEvent(
    dataModelerStateService: DataModelerStateService,
    event_datetime: number,
    model_id: string,
    duration_sec: number,
    total_in_focus: number
  ): ActiveEvent {
    const applicationState = dataModelerStateService.getApplicationState();
    return {
      ...CommonUserMetricsData,
      app_name: this.config.metrics.appName,
      install_id: undefined,
      build_id: "",
      is_dev: false,
      version: "",
      project_id: applicationState.projectId,
      event_type: "active",
      duration_sec,
      total_in_focus,
      event_datetime,
    };
  }
}
