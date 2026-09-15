// Thin typed wrapper over the generated Wails bindings. Centralizes error
// handling so screens don't each need their own try/catch boilerplate, and
// gives call sites plain TS types (with correct null-ability, which the
// generated .d.ts doesn't express for single-object results).
import * as Backend from "../wailsjs/go/main/App";
import type { models } from "../wailsjs/go/models";

export type Profile = models.Profile;
export type Fast = models.Fast;
export type WeightLog = models.WeightLog;
export type Stats = models.Stats;
export type CreateProfileInput = models.CreateProfileInput;
export type UpdateProfileInput = models.UpdateProfileInput;

export class ApiError extends Error {}

async function call<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn();
  } catch (err) {
    const message = typeof err === "string" ? err : err instanceof Error ? err.message : "Something went wrong";
    throw new ApiError(message);
  }
}

export const api = {
  hasProfile: () => call(() => Backend.HasProfile()),
  getProfile: () => call(async () => (await Backend.GetProfile()) ?? null) as Promise<Profile | null>,
  createProfile: (input: CreateProfileInput) => call(() => Backend.CreateProfile(input)),
  updateProfile: (input: UpdateProfileInput) => call(() => Backend.UpdateProfile(input)),
  addWeightLog: (weightKg: number, loggedAt: string) => call(() => Backend.AddWeightLog(weightKg, loggedAt)),
  listWeightLogs: () => call(() => Backend.ListWeightLogs()),

  getActiveFast: () => call(async () => (await Backend.GetActiveFast()) ?? null) as Promise<Fast | null>,
  // startTime: RFC3339 timestamp for when the fast actually began; "" (the
  // default) means "now". A past timestamp back-tracks a fast the user
  // forgot to start in the app at the time.
  startFast: (plannedDurationHours: number, startTime = "") => call(() => Backend.StartFast(plannedDurationHours, startTime)),
  completeFast: (id: string, mood: number, note: string) => call(() => Backend.CompleteFast(id, mood, note)),
  abandonFast: (id: string, mood: number, note: string) => call(() => Backend.AbandonFast(id, mood, note)),
  listFasts: (limit = 50, offset = 0) => call(() => Backend.ListFasts(limit, offset)),

  getStats: () => call(() => Backend.GetStats()),
};
