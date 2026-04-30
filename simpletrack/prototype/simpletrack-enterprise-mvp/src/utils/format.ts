export const integer = new Intl.NumberFormat("en-US");

export function compactPercent(value: number) {
  return `${value.toFixed(1)}%`;
}

export function propertyRows(distribution: Record<string, number>) {
  return Object.entries(distribution).map(([name, value]) => ({ key: name, name, value }));
}
