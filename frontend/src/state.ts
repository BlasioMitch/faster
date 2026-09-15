// Minimal observable store — no state-management library. Holds just the
// two pieces of data shared across screens/nav: the cached profile (so nav
// knows onboarding is done) and whether a fast is currently active (so the
// Fasting tab/nav can reflect it without every screen re-fetching).
import type { Fast, Profile } from "./api";

type Listener = () => void;

interface AppState {
  profile: Profile | null;
  activeFast: Fast | null;
}

const state: AppState = { profile: null, activeFast: null };
const listeners = new Set<Listener>();

export function getProfile(): Profile | null {
  return state.profile;
}

export function getActiveFast(): Fast | null {
  return state.activeFast;
}

export function setProfile(profile: Profile | null): void {
  state.profile = profile;
  notify();
}

export function setActiveFast(fast: Fast | null): void {
  state.activeFast = fast;
  notify();
}

export function subscribe(listener: Listener): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

function notify(): void {
  for (const listener of listeners) listener();
}
