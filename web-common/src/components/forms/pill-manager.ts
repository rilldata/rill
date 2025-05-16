export type PillItem = {
  id: string;
  value: string;
};

export class PillManager {
  private pills: PillItem[] = [];
  private onChange: (values: string[]) => void;

  constructor(initialValues: string[], onChange: (values: string[]) => void) {
    this.onChange = onChange;
    this.setValues(initialValues);
  }

  private generateId(): string {
    return Math.random().toString(36).substring(2, 10);
  }

  getPills(): PillItem[] {
    return this.pills;
  }

  getValues(): string[] {
    return this.pills.map((p) => p.value);
  }

  setValues(values: string[]) {
    // Always ensure we have an empty pill at the end
    const nonEmptyValues = values.filter((v) => v.trim());
    this.pills = [
      ...nonEmptyValues.map((val) => ({
        id: this.generateId(),
        value: val,
      })),
      { id: this.generateId(), value: "" },
    ];
    this.notifyChange();
  }

  updatePillValue(pillId: string, newValue: string) {
    const index = this.pills.findIndex((p) => p.id === pillId);
    if (index !== -1) {
      this.pills[index] = { ...this.pills[index], value: newValue };
      this.notifyChange();
    }
  }

  removePill(pillId: string) {
    this.pills = this.pills.filter((p) => p.id !== pillId);
    // Ensure we always have an empty pill at the end
    if (!this.pills.some((p) => !p.value.trim())) {
      this.pills.push({ id: this.generateId(), value: "" });
    }
    this.notifyChange();
  }

  private notifyChange() {
    this.onChange(this.getValues());
  }
}
