export interface DomainCoordinates<T extends number | Date = number | Date> {
  x?: T;
  xActual?: number;
  y?: number;
  yActual?: number;
}
