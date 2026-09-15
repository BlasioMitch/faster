// Unit conversion helpers. Weight/height are always stored canonically in
// kg/cm in the database; these only translate for display and form input.
const KG_PER_LB = 0.45359237;

export function kgToLb(kg: number): number {
  return kg / KG_PER_LB;
}

export function lbToKg(lb: number): number {
  return lb * KG_PER_LB;
}

export function cmToInches(cm: number): number {
  return cm / 2.54;
}

export function inchesToCm(inches: number): number {
  return inches * 2.54;
}

export function cmToFeetInches(cm: number): { feet: number; inches: number } {
  const totalInches = cmToInches(cm);
  const feet = Math.floor(totalInches / 12);
  const inches = totalInches - feet * 12;
  return { feet, inches };
}

export function feetInchesToCm(feet: number, inches: number): number {
  return inchesToCm(feet * 12 + inches);
}

export function formatWeight(weightKg: number, unit: "kg" | "lb"): string {
  return unit === "lb" ? `${kgToLb(weightKg).toFixed(1)} lb` : `${weightKg.toFixed(1)} kg`;
}

export function formatHeight(heightCm: number, unit: "cm" | "ft_in"): string {
  if (unit === "ft_in") {
    const { feet, inches } = cmToFeetInches(heightCm);
    return `${feet}' ${inches.toFixed(0)}"`;
  }
  return `${heightCm.toFixed(0)} cm`;
}
