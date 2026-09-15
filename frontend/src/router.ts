// A ~50-line hash router: no framework, no history library. Each screen
// module exports a `mount(container, ctx)` that may return a cleanup
// function — the router calls it before swapping screens, which is how the
// Fasting screen's 1s countdown interval gets torn down on navigation
// instead of stacking up.
import { fastingScreen } from "./screens/fasting";
import { onboardingScreen } from "./screens/onboarding";
import { profileScreen } from "./screens/profile";
import { reportsScreen } from "./screens/reports";

export type Route = "onboarding" | "profile" | "fasting" | "reports";

export type UnmountFn = () => void;

export interface ScreenContext {
  navigate(route: Route): void;
}

export interface Screen {
  mount(container: HTMLElement, ctx: ScreenContext): void | UnmountFn | Promise<void | UnmountFn>;
}

const SCREENS: Record<Route, Screen> = {
  onboarding: onboardingScreen,
  profile: profileScreen,
  fasting: fastingScreen,
  reports: reportsScreen,
};

const VALID_ROUTES: Route[] = ["onboarding", "profile", "fasting", "reports"];

let container: HTMLElement | null = null;
let currentUnmount: UnmountFn | void;
let currentRoute: Route = "fasting";
const routeListeners = new Set<() => void>();

function parseHash(): Route {
  const hash = location.hash.replace(/^#\/?/, "");
  return (VALID_ROUTES as string[]).includes(hash) ? (hash as Route) : "fasting";
}

export function getCurrentRoute(): Route {
  return currentRoute;
}

export function onRouteChange(listener: () => void): () => void {
  routeListeners.add(listener);
  return () => routeListeners.delete(listener);
}

export function navigate(route: Route): void {
  if (location.hash === `#/${route}`) {
    render();
  } else {
    location.hash = `#/${route}`;
  }
}

async function render(): Promise<void> {
  if (!container) return;
  currentRoute = parseHash();

  if (currentUnmount) {
    currentUnmount();
    currentUnmount = undefined;
  }
  container.innerHTML = "";

  const result = await SCREENS[currentRoute].mount(container, { navigate });
  currentUnmount = result ?? undefined;

  for (const listener of routeListeners) listener();
}

/** Starts the router, rendering into `root` and reacting to hash changes. */
export function startRouter(root: HTMLElement): void {
  container = root;
  window.addEventListener("hashchange", render);
  render();
}
