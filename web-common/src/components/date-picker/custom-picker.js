// @ts-nocheck
import Litepicker from "litepicker";
import { DateTime } from "./datetime";
import { findNestedMonthItem, parseLocaleStringDate } from "./util";

const style = {
  dayItem: "day-item",
  litepicker: "litepicker",
  isLocked: "is-locked",
  buttonPreviousMonth: "button-previous-month",
  buttonNextMonth: "button-next-month",
  monthItem: "month-item",
  buttonCancel: "button-cancel",
  buttonApply: "button-apply",
};

export default class Custompicker extends Litepicker {
  constructor(options) {
    super(options);
    this.editingDate = 0;

    this.datePicked = [
      new DateTime(options.startDate),
      new DateTime(options.endDate),
    ];

    this.updateValues();

    this.options.startEl.addEventListener("blur", (e) => {
      const maybeStartDate = parseLocaleStringDate(e.target.value); // new Date(e.target.value);
      if (!isNaN(maybeStartDate.valueOf())) {
        this.datePicked[0] = new DateTime(maybeStartDate);
        this.emitChanged();
        // Only render if the blur event was not caused by clicking on a day-item
        if (
          !e.relatedTarget ||
          !e.relatedTarget.classList.contains("day-item")
        ) {
          this.scrollToSpecificDate(this.datePicked[0]);
          this.render();
        }
      }
    });

    this.options.endEl.addEventListener("blur", (e) => {
      const maybeEndDate = parseLocaleStringDate(e.target.value); // new Date(e.target.value);
      if (!isNaN(maybeEndDate.valueOf())) {
        this.datePicked[1] = new DateTime(maybeEndDate);

        this.emitChanged();
        // Only render if the blur event was not caused by clicking on a day-item
        if (
          !e.relatedTarget ||
          !e.relatedTarget.classList.contains("day-item")
        ) {
          this.scrollToSpecificDate(this.datePicked[1]);
          this.render();
        }
      }
    });
  }

  scrollToSpecificDate(date) {
    const clonedDate = date.clone();
    clonedDate.setDate(1);
    this.calendars[0] = clonedDate.clone();
  }

  setEditingDate(idx) {
    this.editingDate = idx;
    this.emit("editingDate", idx);
  }

  updateValues() {
    const { startEl, endEl } = this.options;
    const [start, end] = this.datePicked;
    const startString = start
      .toJSDate()
      .toLocaleDateString(window.navigator.language);
    const endString = end
      .toJSDate()
      .toLocaleDateString(window.navigator.language);

    if (startEl instanceof HTMLInputElement) {
      startEl.value = startString;
    } else if (startEl instanceof HTMLElement) {
      startEl.innerText = startString;
    }

    if (endEl instanceof HTMLInputElement) {
      endEl.value = endString;
    } else if (endEl instanceof HTMLElement) {
      endEl.innerText = endString;
    }
  }

  emitChanged() {
    this.emit("change", {
      start: this.datePicked[0].toJSDate(),
      end: this.datePicked[1].toJSDate(),
    });
  }

  // Override Litepicker method
  shouldResetDatePicked() {
    return false;
  }

  // Override Litepicker method
  onClick(e) {
    let target = e.target;

    if (e.target.shadowRoot) {
      target = e.composedPath()[0];
    }

    if (!target || !this.ui) {
      return;
    }

    // Click on element
    if (this.shouldShown(target)) {
      this.show(target);
      return;
    }

    // Click outside picker
    if (
      !target.closest(`.${style.litepicker}`) &&
      this.isShowning() &&
      target !== this.options.startEl &&
      target !== this.options.endEl
    ) {
      this.hide();
      return;
    }

    if (!this.isSamePicker(target)) {
      return;
    }

    this.emit("before:click", target);

    if (this.preventClick) {
      this.preventClick = false;
      return;
    }

    // Click on date
    if (target.classList.contains(style.dayItem)) {
      e.preventDefault();

      if (target.classList.contains(style.isLocked)) {
        return;
      }

      this.datePicked = this.getNextProposedRange(
        new DateTime(target.dataset.time)
      );

      const nextIndex = this.editingDate === 0 ? 1 : 0;

      this.setEditingDate(nextIndex);
      this.render();
      this.emitChanged();
      this.updateValues();

      return;
    }

    // Click on button previous month
    if (target.classList.contains(style.buttonPreviousMonth)) {
      e.preventDefault();

      let idx = 0;
      let numberOfMonths =
        this.options.switchingMonths || this.options.numberOfMonths;

      if (this.options.splitView) {
        const monthItem = target.closest(`.${style.monthItem}`);
        idx = findNestedMonthItem(monthItem);
        numberOfMonths = 1;
      }

      this.calendars[idx].setMonth(
        this.calendars[idx].getMonth() - numberOfMonths
      );
      this.gotoDate(this.calendars[idx], idx);

      this.emit("change:month", this.calendars[idx], idx);
      return;
    }

    // Click on button next month
    if (target.classList.contains(style.buttonNextMonth)) {
      e.preventDefault();

      let idx = 0;
      let numberOfMonths =
        this.options.switchingMonths || this.options.numberOfMonths;

      if (this.options.splitView) {
        const monthItem = target.closest(`.${style.monthItem}`);
        idx = findNestedMonthItem(monthItem);
        numberOfMonths = 1;
      }

      this.calendars[idx].setMonth(
        this.calendars[idx].getMonth() + numberOfMonths
      );
      this.gotoDate(this.calendars[idx], idx);

      this.emit("change:month", this.calendars[idx], idx);
      return;
    }
  }

  getNextProposedRange(newDay) {
    const nextProposedRange = [];
    if (this.editingDate === 0) {
      nextProposedRange.push(newDay);
      if (newDay.getTime() > this.datePicked[1].getTime()) {
        nextProposedRange.push(newDay);
      } else nextProposedRange.push(this.datePicked[1]);
    } else if (this.editingDate === 1) {
      nextProposedRange.push(this.datePicked[0]);
      nextProposedRange.push(newDay);
      if (newDay.getTime() < this.datePicked[0].getTime()) {
        nextProposedRange[0] = newDay;
      }
    }
    return nextProposedRange;
  }

  // Override Litepicker method
  onMouseEnter(event) {
    const target = event.target;
    if (this.editingDate === 0) {
      this.ui.classList.add("editing-start");
    } else this.ui.classList.remove("editing-start");
    if (this.editingDate === 1) {
      this.ui.classList.add("editing-end");
    } else this.ui.classList.remove("editing-end");

    if (!this.isDayItem(target)) {
      return;
    }

    const currentDay = new DateTime(target.dataset.time);

    const nextProposedRange = this.getNextProposedRange(currentDay);

    if (this.shouldAllowMouseEnter(target)) {
      let [date1, date2] = this.datePicked;

      const allDayItems = Array.prototype.slice.call(
        this.ui.querySelectorAll(`.${style.dayItem}`)
      );

      allDayItems.forEach((d) => {
        const date = new DateTime(d.dataset.time);
        const day = this.renderDay(date);

        if (date.isBetween(date1, date2)) {
          day.classList.add(style.isInRange);
        }
        if (date.isBetween(nextProposedRange[0], nextProposedRange[1])) {
          day.classList.add("is-in-proposed-range");
        } else day.classList.remove("is-in-proposed-range");

        if (date.isSame(nextProposedRange[0]) && this.editingDate === 0) {
          day.classList.add("is-proposed-start");
        }

        if (date.isSame(nextProposedRange[1]) && this.editingDate === 1) {
          day.classList.add("is-proposed-end");
        }

        d.className = day.className;
      });

      if (this.options.showTooltip) {
        let days = nextProposedRange[1].diff(nextProposedRange[0], "day") + 1;

        if (typeof this.options.tooltipNumber === "function") {
          days = this.options.tooltipNumber.call(this, days);
        }

        if (days > 0) {
          const pluralName = this.pluralSelector(days);
          const pluralText = this.options.tooltipText[pluralName]
            ? this.options.tooltipText[pluralName]
            : `[${pluralName}]`;
          const text = `${days} ${pluralText}`;

          this.showTooltip(target, text);

          // fix bug iOS 10-12 - https://github.com/wakirin/Litepicker/issues/124
          const ua = window.navigator.userAgent;
          const iDevice = /(iphone|ipad)/i.test(ua);
          const iOS11or12 = /OS 1([0-2])/i.test(ua);
          if (iDevice && iOS11or12) {
            target.dispatchEvent(new Event("click"));
          }
        } else {
          this.hideTooltip();
        }
      }
    }
  }

  onMouseLeave() {
    this.render();
  }

  hide() {
    if (!this.isShowning()) {
      return;
    }
    this.updateInput();

    if (this.options.inlineMode) {
      this.render();
      return;
    }

    this.ui.style.display = "none";

    this.emit("hide");
  }

  // Override Litepicker method to support removing event listeners
  bindEvents() {
    this.teardownFns = this.teardownFns || [];
    const clickHandler = this.onClick.bind(this);
    document.addEventListener("click", clickHandler, true);
    this.teardownFns.push(() =>
      document.removeEventListener("click", clickHandler, true)
    );

    this.ui = document.createElement("div");
    this.ui.className = style.litepicker;
    this.ui.style.display = "none";
    const handleMouseEnter = this.onMouseEnter.bind(this);
    const handleMouseLeave = this.onMouseLeave.bind(this);
    this.ui.addEventListener("mouseenter", handleMouseEnter, true);
    this.ui.addEventListener("mouseleave", handleMouseLeave, false);
    this.teardownFns.push(() =>
      this.ui.removeEventListener("mouseenter", handleMouseEnter, true)
    );
    this.teardownFns.push(() =>
      this.ui.removeEventListener("mouseleave", handleMouseLeave, true)
    );

    if (this.options.autoRefresh) {
      if (this.options.element instanceof HTMLElement) {
        this.options.element.addEventListener(
          "keyup",
          this.onInput.bind(this),
          true
        );
      }
      if (this.options.elementEnd instanceof HTMLElement) {
        this.options.elementEnd.addEventListener(
          "keyup",
          this.onInput.bind(this),
          true
        );
      }
    } else {
      if (this.options.element instanceof HTMLElement) {
        this.options.element.addEventListener(
          "change",
          this.onInput.bind(this),
          true
        );
      }
      if (this.options.elementEnd instanceof HTMLElement) {
        this.options.elementEnd.addEventListener(
          "change",
          this.onInput.bind(this),
          true
        );
      }
    }

    if (this.options.parentEl) {
      if (this.options.parentEl instanceof HTMLElement) {
        this.options.parentEl.appendChild(this.ui);
      } else {
        document.querySelector(this.options.parentEl).appendChild(this.ui);
      }
    } else {
      if (this.options.inlineMode) {
        if (this.options.element instanceof HTMLInputElement) {
          this.options.element.parentNode.appendChild(this.ui);
        } else {
          this.options.element.appendChild(this.ui);
        }
      } else {
        document.body.appendChild(this.ui);
      }
    }

    this.updateInput();

    this.init();

    if (typeof this.options.setup === "function") {
      this.options.setup.call(this, this);
    }

    this.render();

    if (this.options.inlineMode) {
      this.show();
    }
  }

  // Teardown picker when done with it
  destroy() {
    this.teardownFns.forEach((fn) => fn());
    this.ui.remove();
  }
}
