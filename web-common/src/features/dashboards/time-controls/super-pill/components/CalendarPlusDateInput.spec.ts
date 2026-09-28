import CalendarPlusDateInputTest from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/CalendarPlusDateInputTest.svelte";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { act, fireEvent, render, screen } from "@testing-library/svelte";
import { DateTime, Interval, type DateTimeUnit } from "luxon";
import { describe, expect, it, vi } from "vitest";

const UTC = "UTC";
const NEW_YORK = "America/New_York";

function dateTime(iso: string, zone: string) {
  return DateTime.fromISO(iso, { zone }) as DateTime<true>;
}

function intervalOf(startIso: string, endIso: string, zone: string) {
  return Interval.fromDateTimes(
    dateTime(startIso, zone),
    dateTime(endIso, zone),
  ) as Interval<true>;
}

function renderCalendar({
  interval,
  maxDate,
  minTimeGrain,
  zone,
  minDate,
}: {
  interval: Interval<true>;
  maxDate: DateTime<true>;
  minTimeGrain: DateTimeUnit;
  zone: string;
  minDate?: DateTime<true>;
}) {
  const updateRange = vi.fn();
  const onApply = vi.fn();
  const closeMenu = vi.fn();

  render(CalendarPlusDateInputTest, {
    props: {
      interval,
      minDate: minDate ?? dateTime("2024-01-01T00:00:00", zone),
      maxDate,
      minTimeGrain,
      zone,
      updateRange,
      onApply,
      closeMenu,
    },
  });

  return { updateRange, onApply, closeMenu };
}

/**
 * The out-of-range indicator is a tooltip trigger rendered next to the input.
 * It uses a yellow icon for out-of-range dates and a red one for invalid dates.
 */
function getOutOfRangeIndicator(boundary: "start" | "end") {
  const indicator = screen
    .getByLabelText(`${boundary} date`)
    .parentElement!.querySelector("button");
  const isOutOfRange = indicator
    ?.querySelector("svg")
    ?.classList.contains("text-yellow-500");
  return isOutOfRange ? indicator : null;
}

async function enterDate(boundary: "start" | "end", value: string) {
  const input = screen.getByLabelText(`${boundary} date`);
  await act(() => fireEvent.focus(input));
  await act(() => fireEvent.input(input, { target: { value } }));
  await act(() => fireEvent.blur(input));
}

describe("CalendarPlusDateInput", () => {
  mockAnimationsForComponentTesting();

  it("allows the day containing the max date when the min grain is smaller than a day", () => {
    renderCalendar({
      // The selected range ends at the end of the day that contains the max date.
      interval: intervalOf("2024-03-15T00:00:00", "2024-03-16T00:00:00", UTC),
      maxDate: dateTime("2024-03-15T10:30:00", UTC),
      minTimeGrain: "hour",
      zone: UTC,
    });

    expect(getOutOfRangeIndicator("start")).toBeNull();
    expect(getOutOfRangeIndicator("end")).toBeNull();
  });

  it("flags dates past the day containing the max date when the min grain is smaller than a day", async () => {
    renderCalendar({
      interval: intervalOf("2024-03-15T00:00:00", "2024-03-16T00:00:00", UTC),
      maxDate: dateTime("2024-03-15T10:30:00", UTC),
      minTimeGrain: "hour",
      zone: UTC,
    });

    await enterDate("end", "Mar 20, 2024");

    expect(getOutOfRangeIndicator("end")).not.toBeNull();
  });

  it("resets an out-of-range end date to the day after the max date", async () => {
    const { updateRange } = renderCalendar({
      interval: intervalOf("2024-03-15T00:00:00", "2024-03-16T00:00:00", UTC),
      maxDate: dateTime("2024-03-15T10:30:00", UTC),
      minTimeGrain: "hour",
      zone: UTC,
    });

    await enterDate("end", "Mar 20, 2024");
    await act(() => fireEvent.click(getOutOfRangeIndicator("end")!));

    expect(updateRange).toHaveBeenLastCalledWith("2024-03-15 to 2024-03-16");
    expect(getOutOfRangeIndicator("end")).toBeNull();
  });

  it("snaps the max date to the min grain when the grain is larger than a day", async () => {
    renderCalendar({
      // Later in the same month as the max date, which is allowed for a month grain.
      interval: intervalOf("2024-03-20T00:00:00", "2024-03-21T00:00:00", UTC),
      maxDate: dateTime("2024-03-15T10:30:00", UTC),
      minTimeGrain: "month",
      zone: UTC,
    });

    expect(getOutOfRangeIndicator("end")).toBeNull();

    // The next month is past the snapped max date.
    await enterDate("end", "Apr 5, 2024");

    expect(getOutOfRangeIndicator("end")).not.toBeNull();
  });

  it("snaps the max date in the dashboard time zone", async () => {
    renderCalendar({
      // In New York the max date is Mar 14 at 22:00, so Mar 14 is the last selectable day.
      interval: intervalOf(
        "2024-03-14T00:00:00",
        "2024-03-15T00:00:00",
        NEW_YORK,
      ),
      maxDate: dateTime("2024-03-15T02:00:00", UTC),
      minTimeGrain: "hour",
      zone: NEW_YORK,
    });

    expect(getOutOfRangeIndicator("end")).toBeNull();

    await enterDate("end", "Mar 15, 2024");

    expect(getOutOfRangeIndicator("end")).not.toBeNull();
  });

  it("keeps the min date snapped to the start of the day for larger grains", () => {
    renderCalendar({
      interval: intervalOf("2024-02-01T00:00:00", "2024-02-02T00:00:00", UTC),
      minDate: dateTime("2024-02-01T14:00:00", UTC),
      maxDate: dateTime("2024-03-15T10:30:00", UTC),
      minTimeGrain: "month",
      zone: UTC,
    });

    expect(getOutOfRangeIndicator("start")).toBeNull();
  });
});
