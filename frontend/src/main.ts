import "./style.css";
import { mountNav } from "./components/nav";
import { startRouter } from "./router";

// App shell: a scrollable main area plus the 3-button bottom nav. Each
// screen is responsible for checking whether onboarding is done (via
// api.hasProfile()) and redirecting itself — see fasting/profile/reports
// screens — so this bootstrap stays framework-free and dumb.
function bootstrap(): void {
  const app = document.getElementById("app");
  if (!app) return;
  app.innerHTML = "";

  const main = document.createElement("main");
  main.className = "app-main";
  const navEl = document.createElement("nav");

  app.appendChild(main);
  app.appendChild(navEl);

  mountNav(navEl);
  startRouter(main);
}

bootstrap();
