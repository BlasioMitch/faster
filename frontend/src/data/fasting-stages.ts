// Static, general-education content on fasting physiology. Identical for
// every user and never queried against user data, so this lives as a plain
// TS module rather than a DB table (see the project plan for the reasoning).
export interface FastingStage {
  minHour: number;
  maxHour: number | null; // null = open-ended (last stage)
  title: string;
  body: string;
}

export const FASTING_STAGES: FastingStage[] = [
  {
    minHour: 0,
    maxHour: 4,
    title: "Fed State",
    body: "Digesting and absorbing your last meal. Blood sugar and insulin are elevated, and energy comes from what you just ate.",
  },
  {
    minHour: 4,
    maxHour: 16,
    title: "Early Fasting / Post-Absorptive",
    body: "Blood sugar and insulin begin to fall. The body shifts toward using stored liver glycogen for energy.",
  },
  {
    minHour: 16,
    maxHour: 18,
    title: "Glycogen Depletion Begins",
    body: "Liver glycogen stores start running low. Fat metabolism increases and ketone production begins to rise.",
  },
  {
    minHour: 18,
    maxHour: 24,
    title: "Rising Ketosis",
    body: "Ketone levels climb as the body increasingly relies on fat for fuel. Shifts in hunger and mental clarity are common in this window.",
  },
  {
    minHour: 24,
    maxHour: 48,
    title: "Ketosis Established",
    body: "Predominantly fat- and ketone-fueled. Growth hormone secretion increases, which may help preserve lean muscle.",
  },
  {
    minHour: 48,
    maxHour: 72,
    title: "Deepening Autophagy",
    body: "Autophagy (cellular clean-up) increases significantly. Growth hormone may rise well above baseline, and insulin sensitivity continues improving.",
  },
  {
    minHour: 72,
    maxHour: null,
    title: "Extended Fasting / Immune Renewal",
    body: "Some research points to stem-cell-mediated immune cell regeneration in this range. Extra caution and professional guidance are recommended for fasts this long.",
  },
];

export const FASTING_DISCLAIMER =
  "Educational information only, based on general fasting research — not medical advice. Consult a healthcare provider before extended fasting, especially beyond 48 hours.";

/** Returns the stage whose [minHour, maxHour) range contains elapsedHours. */
export function currentStage(elapsedHours: number): FastingStage {
  for (const stage of FASTING_STAGES) {
    if (elapsedHours >= stage.minHour && (stage.maxHour === null || elapsedHours < stage.maxHour)) {
      return stage;
    }
  }
  return FASTING_STAGES[FASTING_STAGES.length - 1];
}
