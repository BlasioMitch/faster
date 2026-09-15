import { getCurrentRoute, navigate, onRouteChange, type Route } from "../router";

interface NavItem {
  route: Extract<Route, "profile" | "fasting" | "reports">;
  label: string;
  icon: string;
}

const NAV_ITEMS: NavItem[] = [
  { route: "profile", label: "Profile", icon: "👤" },
  { route: "fasting", label: "Fasting", icon: "⏱" },
  { route: "reports", label: "Reports", icon: "📊" },
];

/** Mounts the 3-button bottom nav and keeps its active state in sync with the router. */
export function mountNav(container: HTMLElement): () => void {
  container.className = "app-nav";
  render();
  return onRouteChange(render);

  function render(): void {
    const active = getCurrentRoute();
    // Hidden during onboarding — there's nothing to navigate to yet, and
    // every other screen bounces back to onboarding until a profile exists.
    container.hidden = active === "onboarding";
    if (container.hidden) return;

    container.innerHTML = "";
    for (const item of NAV_ITEMS) {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "app-nav__btn" + (active === item.route ? " active" : "");
      btn.innerHTML = `<span class="app-nav__icon">${item.icon}</span><span>${item.label}</span>`;
      btn.addEventListener("click", () => navigate(item.route));
      container.appendChild(btn);
    }
  }
}
