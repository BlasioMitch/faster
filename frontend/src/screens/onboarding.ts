import { api } from "../api";
import type { Screen } from "../router";
import { setProfile } from "../state";
import { feetInchesToCm, lbToKg } from "../units";

export const onboardingScreen: Screen = {
  mount(container, ctx) {
    container.innerHTML = `
      <div class="screen">
        <h1>Welcome to Faster</h1>
        <p>Let's set up your profile. Only your username and date of birth are required — weight and height are optional and stay on this device.</p>
        <div class="error-banner" id="error" hidden></div>
        <form id="onboarding-form">
          <div class="form-row">
            <label for="username">Username</label>
            <input id="username" type="text" required maxlength="40" autocomplete="off" />
          </div>
          <div class="form-row">
            <label for="dob">Date of birth</label>
            <input id="dob" type="date" required />
          </div>
          <div class="form-row">
            <label>Weight (optional)</label>
            <div class="form-row-inline">
              <input id="weight" type="number" min="1" step="0.1" placeholder="e.g. 70" />
              <div class="unit-toggle" id="weight-unit-toggle">
                <button type="button" data-unit="kg" class="unit-toggle__btn selected">kg</button>
                <button type="button" data-unit="lb" class="unit-toggle__btn">lb</button>
              </div>
            </div>
          </div>
          <div class="form-row">
            <label>Height (optional)</label>
            <div class="form-row-inline" id="height-inputs"></div>
            <div class="unit-toggle" id="height-unit-toggle" style="margin-top:8px">
              <button type="button" data-unit="cm" class="unit-toggle__btn selected">cm</button>
              <button type="button" data-unit="ft_in" class="unit-toggle__btn">ft / in</button>
            </div>
          </div>
          <button type="submit" class="btn btn-primary btn-block">Get started</button>
        </form>
      </div>
    `;

    const form = container.querySelector<HTMLFormElement>("#onboarding-form")!;
    const errorEl = container.querySelector<HTMLElement>("#error")!;
    const weightUnitToggle = container.querySelector<HTMLElement>("#weight-unit-toggle")!;
    const heightUnitToggle = container.querySelector<HTMLElement>("#height-unit-toggle")!;
    const heightInputs = container.querySelector<HTMLElement>("#height-inputs")!;

    let weightUnit: "kg" | "lb" = "kg";
    let heightUnit: "cm" | "ft_in" = "cm";

    function selectToggle(toggle: HTMLElement, value: string): void {
      toggle.querySelectorAll<HTMLButtonElement>(".unit-toggle__btn").forEach((btn) => {
        btn.classList.toggle("selected", btn.dataset.unit === value);
      });
    }

    function renderHeightInputs(): void {
      heightInputs.innerHTML =
        heightUnit === "cm"
          ? `<input id="height-cm" type="number" min="1" step="0.1" placeholder="e.g. 175" />`
          : `<input id="height-ft" type="number" min="0" step="1" placeholder="feet" />
             <input id="height-in" type="number" min="0" max="11" step="0.1" placeholder="inches" />`;
    }

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

    function showError(message: string): void {
      errorEl.textContent = message;
      errorEl.hidden = false;
    }

    form.addEventListener("submit", async (e) => {
      e.preventDefault();
      errorEl.hidden = true;

      const username = container.querySelector<HTMLInputElement>("#username")!.value.trim();
      const dob = container.querySelector<HTMLInputElement>("#dob")!.value;

      if (!username) return showError("Please enter a username.");
      if (!dob) return showError("Please enter your date of birth.");
      if (new Date(dob) > new Date()) return showError("Date of birth can't be in the future.");

      let weightKg: number | undefined;
      const weightRaw = container.querySelector<HTMLInputElement>("#weight")!.value;
      if (weightRaw) {
        const value = parseFloat(weightRaw);
        if (!Number.isFinite(value) || value <= 0) return showError("Enter a valid weight.");
        weightKg = weightUnit === "lb" ? lbToKg(value) : value;
      }

      let heightValue: number | undefined;
      if (heightUnit === "cm") {
        const raw = container.querySelector<HTMLInputElement>("#height-cm")?.value;
        if (raw) {
          const value = parseFloat(raw);
          if (!Number.isFinite(value) || value <= 0) return showError("Enter a valid height.");
          heightValue = value;
        }
      } else {
        const ftRaw = container.querySelector<HTMLInputElement>("#height-ft")?.value;
        const inRaw = container.querySelector<HTMLInputElement>("#height-in")?.value;
        if (ftRaw || inRaw) {
          const feet = parseFloat(ftRaw || "0");
          const inches = parseFloat(inRaw || "0");
          if (!Number.isFinite(feet) || !Number.isFinite(inches) || feet < 0 || inches < 0) {
            return showError("Enter a valid height.");
          }
          heightValue = feetInchesToCm(feet, inches);
        }
      }

      const submitBtn = form.querySelector<HTMLButtonElement>('button[type="submit"]')!;
      submitBtn.disabled = true;
      try {
        const profile = await api.createProfile({
          username,
          dateOfBirth: dob,
          weightUnit,
          heightUnit,
          weightKg,
          heightValue,
        });
        setProfile(profile);
        ctx.navigate("fasting");
      } catch (err) {
        showError(err instanceof Error ? err.message : "Something went wrong.");
        submitBtn.disabled = false;
      }
    });

    renderHeightInputs();
  },
};

export default onboardingScreen;
