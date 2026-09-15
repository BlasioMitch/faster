import { api } from "../api";
import { escapeHtml } from "../dom";
import type { Screen } from "../router";
import { setProfile } from "../state";
import { cmToFeetInches, feetInchesToCm, lbToKg } from "../units";

function calcAge(dob: string): number | null {
  const d = new Date(dob);
  if (Number.isNaN(d.getTime())) return null;
  const today = new Date();
  let age = today.getFullYear() - d.getFullYear();
  const monthDiff = today.getMonth() - d.getMonth();
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < d.getDate())) age--;
  return age;
}

export const profileScreen: Screen = {
  async mount(container, ctx) {
    container.innerHTML = `<div class="screen"><p>Loading…</p></div>`;
    const loaded = await api.getProfile();
    if (!loaded) {
      ctx.navigate("onboarding");
      return;
    }
    // Re-bind to a const whose declared type is non-null Profile: TS control-flow
    // narrowing on `loaded` doesn't survive into the closures below (they could in
    // principle run after further reassignment), but a fresh const's inferred type
    // does, so nested functions referencing `profile` see it as non-null too.
    const profile = loaded;

    let weightUnit = profile.weightUnit as "kg" | "lb";
    let heightUnit = profile.heightUnit as "cm" | "ft_in";
    const age = calcAge(profile.dateOfBirth);

    container.innerHTML = `
      <div class="screen">
        <h1>Profile</h1>
        <p>${escapeHtml(profile.username)}${age !== null ? ` · ${age} years old` : ""}</p>
        <div class="error-banner" id="error" hidden></div>

        <div class="card">
          <h2>Details</h2>
          <div class="form-row">
            <label for="username">Username</label>
            <input id="username" type="text" value="${escapeHtml(profile.username)}" maxlength="40" />
          </div>
          <div class="form-row">
            <label for="dob">Date of birth</label>
            <input id="dob" type="date" value="${profile.dateOfBirth}" />
          </div>
          <div class="form-row">
            <label>Weight unit</label>
            <div class="unit-toggle" id="weight-unit-toggle">
              <button type="button" data-unit="kg" class="unit-toggle__btn${weightUnit === "kg" ? " selected" : ""}">kg</button>
              <button type="button" data-unit="lb" class="unit-toggle__btn${weightUnit === "lb" ? " selected" : ""}">lb</button>
            </div>
          </div>
          <div class="form-row">
            <label>Height</label>
            <div class="form-row-inline" id="height-inputs"></div>
            <div class="unit-toggle" id="height-unit-toggle" style="margin-top:8px">
              <button type="button" data-unit="cm" class="unit-toggle__btn${heightUnit === "cm" ? " selected" : ""}">cm</button>
              <button type="button" data-unit="ft_in" class="unit-toggle__btn${heightUnit === "ft_in" ? " selected" : ""}">ft / in</button>
            </div>
          </div>
          <button id="save-profile" class="btn btn-primary btn-block">Save changes</button>
        </div>

        <div class="card">
          <h2>Log today's weight</h2>
          <div class="form-row-inline">
            <input id="new-weight" type="number" min="1" step="0.1" placeholder="e.g. 70" />
            <button id="log-weight" class="btn btn-primary">Log</button>
          </div>
          <p style="margin:8px 0 0;font-size:12px;color:var(--color-text-muted);">Your weight trend appears on the Reports screen.</p>
        </div>
      </div>
    `;

    const errorEl = container.querySelector<HTMLElement>("#error")!;
    function showError(message: string): void {
      errorEl.textContent = message;
      errorEl.hidden = false;
    }

    const weightUnitToggle = container.querySelector<HTMLElement>("#weight-unit-toggle")!;
    const heightUnitToggle = container.querySelector<HTMLElement>("#height-unit-toggle")!;
    const heightInputs = container.querySelector<HTMLElement>("#height-inputs")!;

    function selectToggle(toggle: HTMLElement, value: string): void {
      toggle.querySelectorAll<HTMLButtonElement>(".unit-toggle__btn").forEach((btn) => {
        btn.classList.toggle("selected", btn.dataset.unit === value);
      });
    }

    function renderHeightInputs(): void {
      if (heightUnit === "cm") {
        heightInputs.innerHTML = `<input id="height-cm" type="number" min="1" step="0.1" placeholder="e.g. 175" value="${profile.heightValue ?? ""}" />`;
      } else {
        const current = profile.heightValue != null ? cmToFeetInches(profile.heightValue) : null;
        heightInputs.innerHTML = `
          <input id="height-ft" type="number" min="0" step="1" placeholder="feet" value="${current ? Math.floor(current.feet) : ""}" />
          <input id="height-in" type="number" min="0" max="11" step="0.1" placeholder="inches" value="${current ? current.inches.toFixed(0) : ""}" />
        `;
      }
    }
    renderHeightInputs();

    weightUnitToggle.addEventListener("click", (e) => {
      const btn = (e.target as HTMLElement).closest<HTMLButtonElement>("[data-unit]");
      if (!btn) return;
      weightUnit = btn.dataset.unit as "kg" | "lb";
      selectToggle(weightUnitToggle, weightUnit);
    });

    heightUnitToggle.addEventListener("click", (e) => {
      const btn = (e.target as HTMLElement).closest<HTMLButtonElement>("[data-unit]");
      if (!btn) return;
      heightUnit = btn.dataset.unit as "cm" | "ft_in";
      selectToggle(heightUnitToggle, heightUnit);
      renderHeightInputs();
    });

    container.querySelector<HTMLButtonElement>("#save-profile")!.addEventListener("click", async () => {
      errorEl.hidden = true;
      const username = container.querySelector<HTMLInputElement>("#username")!.value.trim();
      const dob = container.querySelector<HTMLInputElement>("#dob")!.value;
      if (!username) return showError("Username can't be empty.");
      if (!dob || new Date(dob) > new Date()) return showError("Enter a valid date of birth.");

      let heightValue: number | undefined;
      if (heightUnit === "cm") {
        const raw = container.querySelector<HTMLInputElement>("#height-cm")?.value;
        if (raw) heightValue = parseFloat(raw);
      } else {
        const ftRaw = container.querySelector<HTMLInputElement>("#height-ft")?.value;
        const inRaw = container.querySelector<HTMLInputElement>("#height-in")?.value;
        if (ftRaw || inRaw) heightValue = feetInchesToCm(parseFloat(ftRaw || "0"), parseFloat(inRaw || "0"));
      }

      try {
        const updated = await api.updateProfile({ username, dateOfBirth: dob, weightUnit, heightUnit, heightValue });
        setProfile(updated);
        profile.username = updated.username;
        profile.dateOfBirth = updated.dateOfBirth;
        profile.weightUnit = updated.weightUnit;
        profile.heightUnit = updated.heightUnit;
        profile.heightValue = updated.heightValue;
      } catch (err) {
        showError(err instanceof Error ? err.message : "Couldn't save changes.");
      }
    });

    container.querySelector<HTMLButtonElement>("#log-weight")!.addEventListener("click", async () => {
      errorEl.hidden = true;
      const input = container.querySelector<HTMLInputElement>("#new-weight")!;
      const raw = input.value;
      if (!raw) return showError("Enter a weight to log.");
      const value = parseFloat(raw);
      if (!Number.isFinite(value) || value <= 0) return showError("Enter a valid weight.");
      const weightKg = weightUnit === "lb" ? lbToKg(value) : value;
      try {
        await api.addWeightLog(weightKg, new Date().toISOString());
        input.value = "";
      } catch (err) {
        showError(err instanceof Error ? err.message : "Couldn't log weight.");
      }
    });
  },
};

export default profileScreen;
