import { DateTime } from "@easepick/datetime";
import { BasePlugin } from "@easepick/base-plugin";
import type { IEventDetail, IPlugin } from "@easepick/base-plugin";
// import type { IRangeConfig } from "./interface";

type IRangeConfig = any;

export class RangePlugin extends BasePlugin implements IPlugin {
  public tooltipElement: HTMLElement;
  public triggerElement: HTMLElement;

  public binds = {
    setStartDate: this.setStartDate.bind(this),
    setEndDate: this.setEndDate.bind(this),
    setDateRange: this.setDateRange.bind(this),
    getStartDate: this.getStartDate.bind(this),
    getEndDate: this.getEndDate.bind(this),
    onView: this.onView.bind(this),
    onShow: this.onShow.bind(this),
    onMouseEnter: this.onMouseEnter.bind(this),
    onMouseLeave: this.onMouseLeave.bind(this),
    onClickCalendarDay: this.onClickCalendarDay.bind(this),
    onClickApplyButton: this.onClickApplyButton.bind(this),
    parseValues: this.parseValues.bind(this),
    updateValues: this.updateValues.bind(this),
    clear: this.clear.bind(this),
    setEditingDate: this.setEditingDate.bind(this),
    hide: this.hide.bind(this),
  };

  public options: IRangeConfig = {
    elementEnd: undefined,
    startDate: undefined,
    endDate: undefined,
    repick: false,
    strict: true,
    delimiter: " - ",
    tooltip: true,
    tooltipNumber: (num: number) => {
      return num;
    },
    locale: {
      zero: "",
      one: "day",
      two: "",
      few: "",
      many: "",
      other: "days",
    },
    documentClick: this.hidePicker.bind(this),
  };

  /**
   * Returns plugin name
   *
   * @returns String
   */
  public getName(): string {
    return "RangePlugin";
  }

  /**
   * - Called automatically via BasePlugin.attach() -
   * The function execute on initialize the picker
   */
  public onAttach(): void {
    this.picker.editingDate = 0;
    this.binds["_setStartDate"] = this.picker.setStartDate;
    this.binds["_setEndDate"] = this.picker.setEndDate;
    this.binds["_setDateRange"] = this.picker.setDateRange;
    this.binds["_getStartDate"] = this.picker.getStartDate;
    this.binds["_getEndDate"] = this.picker.getEndDate;
    this.binds["_parseValues"] = this.picker.parseValues;
    this.binds["_updateValues"] = this.picker.updateValues;
    this.binds["_clear"] = this.picker.clear;
    this.binds["_onClickCalendarDay"] = this.picker.onClickCalendarDay;
    this.binds["_onClickApplyButton"] = this.picker.onClickApplyButton;
    this.binds["_setEditingDate"] = this.picker.setEditingDate;
    this.binds["_hide"] = this.picker.hide;

    Object.defineProperties(this.picker, {
      setStartDate: {
        configurable: true,
        value: this.binds.setStartDate,
      },
      setEndDate: {
        configurable: true,
        value: this.binds.setEndDate,
      },
      setDateRange: {
        configurable: true,
        value: this.binds.setDateRange,
      },
      getStartDate: {
        configurable: true,
        value: this.binds.getStartDate,
      },
      getEndDate: {
        configurable: true,
        value: this.binds.getEndDate,
      },
      parseValues: {
        configurable: true,
        value: this.binds.parseValues,
      },
      updateValues: {
        configurable: true,
        value: this.binds.updateValues,
      },
      clear: {
        configurable: true,
        value: this.binds.clear,
      },
      onClickCalendarDay: {
        configurable: true,
        value: this.binds.onClickCalendarDay,
      },
      onClickApplyButton: {
        configurable: true,
        value: this.binds.onClickApplyButton,
      },
      setEditingDate: {
        configurable: true,
        value: this.binds.setEditingDate,
      },
      hide: {
        configurable: true,
        value: this.binds.hide,
      },
    });

    if (typeof this.picker.options.documentClick === "function") {
      document.removeEventListener(
        "click",
        this.picker.options.documentClick,
        true
      );

      if (typeof this.options.documentClick === "function") {
        document.addEventListener("click", this.options.documentClick, true);
      }
    }

    // Remove repick?
    this.options.repick =
      this.options.repick && this.options.elementEnd instanceof HTMLElement;

    // TODO What is this for?
    this.picker.options.date = null;

    this.picker.on("view", this.binds.onView);
    this.picker.on("show", this.binds.onShow);
    this.picker.on("mouseenter", this.binds.onMouseEnter, true);
    this.picker.on("mouseleave", this.binds.onMouseLeave, true);

    this.picker.on(
      "mousedown",
      (e) => {
        if (this.picker.isCalendarDay(e.target)) {
          // Don't blur active input if we start clicking on calendar day
          e.preventDefault();
        }
      },
      true
    );

    this.checkIntlPluralLocales();

    // TODO Can we make this UTC?
    this.options.startEl.addEventListener("blur", (e) => {
      const maybeStartDate = new Date(e.target.value + " EDT");
      if (!isNaN(maybeStartDate.valueOf())) {
        this.setStartDate(new DateTime(maybeStartDate));
      }
    });

    this.options.endEl.addEventListener("blur", (e) => {
      const maybeEndDate = new Date(e.target.value + " EDT");
      if (!isNaN(maybeEndDate.valueOf())) {
        this.setEndDate(new DateTime(maybeEndDate));
      }
    });
  }

  /**
   * - Called automatically via BasePlugin.detach() -
   */
  public onDetach(): void {
    Object.defineProperties(this.picker, {
      setStartDate: {
        configurable: true,
        value: this.binds["_setStartDate"],
      },
      setEndDate: {
        configurable: true,
        value: this.binds["_setEndDate"],
      },
      setDateRange: {
        configurable: true,
        value: this.binds["_setDateRange"],
      },
      getStartDate: {
        configurable: true,
        value: this.binds["_getStartDate"],
      },
      getEndDate: {
        configurable: true,
        value: this.binds["_getEndDate"],
      },
      parseValues: {
        configurable: true,
        value: this.binds["_parseValues"],
      },
      updateValues: {
        configurable: true,
        value: this.binds["_updateValues"],
      },
      clear: {
        configurable: true,
        value: this.binds["_clear"],
      },
      onClickCalendarDay: {
        configurable: true,
        value: this.binds["_onClickCalendarDay"],
      },
      onClickApplyButton: {
        configurable: true,
        value: this.binds["_onClickApplyButton"],
      },
      setEditingDate: {
        configurable: true,
        value: this.binds["_setEditingDate"],
      },
    });

    this.picker.off("view", this.binds.onView);
    this.picker.off("show", this.binds.onShow);
    this.picker.off("mouseenter", this.binds.onMouseEnter, true);
    this.picker.off("mouseleave", this.binds.onMouseLeave, true);

    // TODO Other listeners to remove from onAttach?
  }

  /**
   * Parse `startDate`, `endDate` options or value of input elements
   */
  // TODO when is this called, why? can we remove it?
  // Looks like its just on init. maybe we can simplify this then, figure out how to properly pass starting values
  private parseValues() {
    if (this.options.startDate || this.options.endDate) {
      if (this.options.strict) {
        if (this.options.startDate && this.options.endDate) {
          this.setDateRange(this.options.startDate, this.options.endDate);
        } else {
          this.options.startDate = null;
          this.options.endDate = null;
        }
      } else {
        if (this.options.startDate) {
          this.setStartDate(this.options.startDate);
        }

        if (this.options.endDate) {
          this.setEndDate(this.options.endDate);
        }
      }
      return;
    }

    if (this.options.elementEnd) {
      if (this.options.strict) {
        if (
          this.picker.options.element instanceof HTMLInputElement &&
          this.picker.options.element.value.length &&
          this.options.elementEnd instanceof HTMLInputElement &&
          this.options.elementEnd.value.length
        ) {
          this.setDateRange(
            this.picker.options.element.value,
            this.options.elementEnd.value
          );
        }
      } else {
        if (
          this.picker.options.element instanceof HTMLInputElement &&
          this.picker.options.element.value.length
        ) {
          this.setStartDate(this.picker.options.element.value);
        }

        if (
          this.options.elementEnd instanceof HTMLInputElement &&
          this.options.elementEnd.value.length
        ) {
          this.setEndDate(this.options.elementEnd.value);
        }
      }
    } else if (
      this.picker.options.element instanceof HTMLInputElement &&
      this.picker.options.element.value.length
    ) {
      const [_start, _end] = this.picker.options.element.value.split(
        this.options.delimiter
      );

      if (this.options.strict) {
        if (_start && _end) {
          this.setDateRange(_start, _end);
        }
      } else {
        if (_start) this.setStartDate(_start);
        if (_end) this.setEndDate(_end);
      }
    }
  }

  /**
   * Update value of input element
   */
  private updateValues() {
    const el = this.options.startEl;
    const elEnd = this.options.endEl;
    const start = this.picker.getStartDate();
    const end = this.picker.getEndDate();
    const startString =
      start instanceof Date
        ? start.format(this.picker.options.format, this.picker.options.lang)
        : "";
    const endString =
      end instanceof Date
        ? end.format(this.picker.options.format, this.picker.options.lang)
        : "";

    if (el instanceof HTMLInputElement) {
      el.value = startString;
    } else if (el instanceof HTMLElement) {
      el.innerText = startString;
    }

    if (elEnd instanceof HTMLInputElement) {
      elEnd.value = endString;
    } else if (elEnd instanceof HTMLElement) {
      elEnd.innerText = endString;
    }
  }

  /**
   * Clear selection
   */
  private clear() {
    this.options.startDate = null;
    this.options.endDate = null;
    // TODO: do we need this?
    this.picker.datePicked.length = 0;
    // alt: this.picker.datePicked = new Array(2);
    this.updateValues();
    this.picker.renderAll();
    this.picker.trigger("clear");
  }

  /**
   * Function `show` event
   *
   * @param event
   */
  private onShow(event) {
    const { target }: IEventDetail = event.detail;
    this.triggerElement = target;

    // TODO: scroll to date config? defaults to true in core plugin
    if (
      this.picker.options.scrollToDate &&
      this.getStartDate() instanceof Date
    ) {
      this.picker.gotoDate(this.getStartDate());
    }

    // TODO: do we need this?
    this.initializeRepick();
  }

  /**
   * Function `view` event
   * Adds HTML layout of current plugin to the picker layout
   *
   * @param event
   */
  private onView(event: CustomEvent) {
    const { view, target }: IEventDetail = event.detail;

    if (view === "Main") {
      this.tooltipElement = document.createElement("span");
      this.tooltipElement.className = "range-plugin-tooltip";
      target.appendChild(this.tooltipElement);
    }

    // TODO: other logic for modifying the view? Like highlighting potential new range?
    if (view === "CalendarDay") {
      const date = new DateTime(target.dataset.time);
      const datePicked = this.picker.datePicked;
      const start = datePicked.length
        ? this.picker.datePicked[0]
        : this.getStartDate();
      const end = datePicked.length
        ? this.picker.datePicked[1]
        : this.getEndDate();

      if (start && start.isSame(date, "day")) {
        target.classList.add("start");
      }

      if (start && end) {
        if (end.isSame(date, "day")) {
          target.classList.add("end");
        }

        if (date.isBetween(start, end)) {
          target.classList.add("in-range");
        }
      }
    }

    // TODO review footer logic for Apply button
    if (view === "Footer") {
      const allowApplyBtn =
        (this.picker.datePicked.length === 1 && !this.options.strict) ||
        this.picker.datePicked.length === 2;
      const applyButton = target.querySelector(
        ".apply-button"
      ) as HTMLButtonElement;
      applyButton.disabled = !allowApplyBtn;
    }
  }

  /**
   * Function for documentClick option
   * Allows the picker to close when the user clicks outside
   *
   * @param e
   */
  private hidePicker(e) {
    let target = e.target;
    let host = null;

    if (target.shadowRoot) {
      target = e.composedPath()[0];
      host = target.getRootNode().host;
    }

    if (
      this.picker.isShown() &&
      host !== this.picker.ui.wrapper &&
      target !== this.picker.options.element &&
      target !== this.options.startEl &&
      target !== this.options.endEl
    ) {
      this.picker.hide();
    }
  }

  /**
   * Set startDate programmatically
   *
   * @param date
   */
  private setStartDate(date: Date | string | number) {
    const d = new DateTime(date, this.picker.options.format);
    this.options.startDate = d ? d.clone() : null;

    this.picker.datePicked[0] = date;

    // TODO maybe rename updateValues to updateDisplayValues
    this.updateValues();

    this.picker.renderAll();
  }

  /**
   * Set endDate programmatically
   *
   * @param date
   */
  private setEndDate(date: Date | string | number) {
    const d = new DateTime(date, this.picker.options.format);
    this.options.endDate = d ? d.clone() : null;

    this.picker.datePicked[1] = date;

    this.updateValues();

    this.picker.renderAll();
  }

  private setEditingDate(v) {
    this.picker.editingDate = v;
    // Why does this need a timeout to work?
    setTimeout(() => {
      if (v === 0) this.options.startEl.focus();
      else this.options.endEl.focus();
    });

    this.picker.trigger("editingDate", v);
  }

  /**
   * Set date range programmatically
   *
   * @param start
   * @param end
   */
  private setDateRange(
    start: Date | string | number,
    end: Date | string | number
  ) {
    const startDate = new DateTime(start, this.picker.options.format);
    const endDate = new DateTime(end, this.picker.options.format);

    this.options.startDate = startDate ? startDate.clone() : null;
    this.options.endDate = endDate ? endDate.clone() : null;

    this.updateValues();

    this.picker.renderAll();
  }

  /**
   *
   * @returns DateTime
   */
  private getStartDate(): DateTime {
    return this.options.startDate instanceof Date
      ? this.options.startDate.clone()
      : null;
  }

  /**
   *
   * @returns
   */
  private getEndDate(): DateTime {
    return this.options.endDate instanceof Date
      ? this.options.endDate.clone()
      : null;
  }

  /**
   * Handle `mouseenter` event
   *
   * @param event
   */
  // TODO: what is this for? tweak for our purposes
  private onMouseEnter(event) {
    const target = event.target;

    if (target instanceof HTMLElement) {
      if (this.isContainer(target)) {
        this.initializeRepick();
      }

      const element = target.closest(".unit");

      if (!(element instanceof HTMLElement)) return;

      if (this.picker.isCalendarDay(element)) {
        if (this.picker.datePicked.length !== 1) return;

        let date1 = this.picker.datePicked[0].clone();
        let date2 = new DateTime(element.dataset.time);
        let isFlipped = false;

        if (date1.isAfter(date2, "day")) {
          const tempDate = date1.clone();
          date1 = date2.clone();
          date2 = tempDate.clone();
          isFlipped = true;
        }

        const days = [...this.picker.ui.container.querySelectorAll(".day")];

        days.forEach((d: HTMLElement) => {
          const date = new DateTime(d.dataset.time);
          const dayView = this.picker.Calendar.getCalendarDayView(date);

          if (date.isBetween(date1, date2)) {
            dayView.classList.add("in-range");
          }

          if (date.isSame(this.picker.datePicked[0], "day")) {
            dayView.classList.add("start");
            dayView.classList.toggle("flipped", isFlipped);
          }

          if (d === element) {
            dayView.classList.add("end");
            dayView.classList.toggle("flipped", isFlipped);
          }

          d.className = dayView.className;
        });

        if (this.options.tooltip) {
          const diff = this.options.tooltipNumber(date2.diff(date1, "day") + 1);

          if (diff > 0) {
            const pluralKey = new Intl.PluralRules(
              this.picker.options.lang
            ).select(diff);
            const text = `${diff} ${this.options.locale[pluralKey]}`;

            this.showTooltip(element, text);
          } else {
            this.hideTooltip();
          }
        }
      }
    }
  }

  /**
   * Handle `mouseleave` event
   *
   * @param event
   */
  private onMouseLeave(event) {
    if (this.isContainer(event.target) && this.options.repick) {
      const start = this.getStartDate();
      const end = this.getEndDate();

      if (start && end) {
        this.picker.datePicked.length = 0;

        this.picker.renderAll();
      }
    }
  }

  private onClickCalendarDay(element: HTMLElement) {
    if (this.picker.isCalendarDay(element)) {
      const date = new DateTime(element.dataset.time);
      this.picker.datePicked.length = 2;
      this.picker.datePicked[this.picker.editingDate] = date;
      this.setEditingDate(this.picker.editingDate === 0 ? 1 : 0);

      if (!this.picker.datePicked[0]) {
        this.picker.datePicked[0] = this.picker.datePicked[1].clone();
      } else if (!this.picker.datePicked[1]) {
        this.picker.datePicked[1] = this.picker.datePicked[0].clone();
      }

      // This flips the dates if they are in the wrong order,
      // But should it reset instead?
      if (
        this.picker.datePicked[0] &&
        this.picker.datePicked[0].isAfter(this.picker.datePicked[1])
      ) {
        const tempDate = this.picker.datePicked[1].clone();
        this.picker.datePicked[1] = this.picker.datePicked[0].clone();
        this.picker.datePicked[0] = tempDate.clone();
      }

      this.setDateRange(this.picker.datePicked[0], this.picker.datePicked[1]);

      if (this.picker.options.autoApply) {
        this.setDateRange(this.picker.datePicked[0], this.picker.datePicked[1]);

        this.picker.trigger("select", {
          start: this.picker.getStartDate(),
          end: this.picker.getEndDate(),
        });

        this.picker.hide();
      } else {
        this.hideTooltip();

        this.picker.renderAll();
      }
    }
  }

  private hide(): void {
    this.picker.ui.container.classList.remove("show");

    this.picker.renderAll();

    this.picker.trigger("hide");
  }

  // TODO: do we even need this? may just use apply button in existing TimeRangeInput?
  private onClickApplyButton(element: HTMLElement) {
    if (this.picker.isApplyButton(element)) {
      //   if (this.picker.datePicked.length === 1 && !this.options.strict) {
      //     if (this.picker.options.element === this.triggerElement) {
      //       this.options.endDate = null;
      //       this.setStartDate(this.picker.datePicked[0]);
      //     }

      //     if (this.options.elementEnd === this.triggerElement) {
      //       this.options.startDate = null;
      //       this.setEndDate(this.picker.datePicked[0]);
      //     }
      //   }

      //   if (this.picker.datePicked.length === 2) {
      //     this.setDateRange(this.picker.datePicked[0], this.picker.datePicked[1]);
      //   }

      this.picker.trigger("select", {
        start: this.picker.getStartDate(),
        end: this.picker.getEndDate(),
      });

      this.picker.hide();
    }
  }

  /**
   * Displays tooltip of selected days
   *
   * @param element
   * @param text
   */
  private showTooltip(element: HTMLElement, text: string) {
    this.tooltipElement.style.visibility = "visible";
    this.tooltipElement.innerHTML = text;

    const container = this.picker.ui.container.getBoundingClientRect();
    const tooltip = this.tooltipElement.getBoundingClientRect();
    const day = element.getBoundingClientRect();
    let top = day.top;
    let left = day.left;

    top -= container.top;
    left -= container.left;

    top -= tooltip.height;
    left -= tooltip.width / 2;
    left += day.width / 2;

    this.tooltipElement.style.top = `${top}px`;
    this.tooltipElement.style.left = `${left}px`;
  }

  /**
   * Hide tooltip
   */
  private hideTooltip() {
    this.tooltipElement.style.visibility = "hidden";
  }

  /**
   * Determines if the locale option contains all required plurals
   */
  private checkIntlPluralLocales() {
    if (!this.options.tooltip) return;

    const rules = [
      ...new Set([
        new Intl.PluralRules(this.picker.options.lang).select(0),
        new Intl.PluralRules(this.picker.options.lang).select(1),
        new Intl.PluralRules(this.picker.options.lang).select(2),
        new Intl.PluralRules(this.picker.options.lang).select(6),
        new Intl.PluralRules(this.picker.options.lang).select(18),
      ]),
    ];

    const locales = Object.keys(this.options.locale);

    if (!rules.every((x) => locales.includes(x))) {
      console.warn(
        `${this.getName()}: provide locales (${rules.join(
          ", "
        )}) for correct tooltip text.`
      );
    }
  }

  /**
   * Handle `repick` option
   */
  // TODO do we need this or can we drop the repick option?
  private initializeRepick() {
    if (!this.options.repick) return;

    const start = this.getStartDate();
    const end = this.getEndDate();

    if (end && this.triggerElement === this.picker.options.element) {
      this.picker.datePicked[0] = end;
    }

    if (start && this.triggerElement === this.options.elementEnd) {
      this.picker.datePicked[0] = start;
    }
  }

  /**
   * Determines if the element is the picker container
   *
   * @param element
   * @returns Boolean
   */
  private isContainer(element: HTMLElement): boolean {
    return element === this.picker.ui.container;
  }
}
