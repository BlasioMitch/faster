// Fasting duration presets shown as tiles on the Fasting screen. This is
// plain static data, not a DB table — fasts.plannedDurationHours accepts
// any number, so a custom duration is just a bounded numeric input, not a
// separate concept.
export interface FastPreset {
  hours: number;
  label: string;
}

export const FAST_PRESETS: FastPreset[] = [
  { hours: 24, label: "24 hours" },
  { hours: 36, label: "36 hours" },
  { hours: 48, label: "48 hours" },
  { hours: 72, label: "72 hours" },
  { hours: 120, label: "120 hours" },
];

export const CUSTOM_DURATION_MIN_HOURS = 1;
export const CUSTOM_DURATION_MAX_HOURS = 240;

// How far in the past a fast's start time may be back-tracked (e.g. "I
// forgot to open the app when I actually started"). Must match
// internal/db.MaxBackdateHours in fasts.go — the backend re-validates this
// bound independently, this constant only drives the UI's min/max hints.
export const MAX_BACKDATE_HOURS = 24 * 7;
