import { api, type Fast } from "../api";
import { createRingTimer } from "../components/ring-timer";
import { promptMoodAndNote } from "../components/modal";
import { currentStage, FASTING_DISCLAIMER, FASTING_STAGES } from "../data/fasting-stages";
import { CUSTOM_DURATION_MAX_HOURS, CUSTOM_DURATION_MIN_HOURS, FAST_PRESETS, MAX_BACKDATE_HOURS } from "../data/presets";
import type { Screen, ScreenContext, UnmountFn } from "../router";
import { setActiveFast } from "../state";
import { formatClock, formatDateTime, toDatetimeLocalValue } from "../time";

export const fastingScreen: Screen = {
  async mount(container, ctx) {
    container.innerHTML = `<div class="screen"><p>Loading…</p></div>`;

    const hasProfile = await api.hasProfile();
    if (!hasProfile) {
      ctx.navigate("onboarding");
      return;
    }

    const active = await api.getActiveFast();
    setActiveFast(active);

    return active ? mountActive(container, ctx, active) : mountIdle(container, ctx);
  },
};

export default fastingScreen;

function mountIdle(container: HTMLElement, ctx: ScreenContext): void {
  container.innerHTML = `
    <div class="screen">
      <h1>Start a fast</h1>
      <p>Pick a duration to begin your countdown.</p>
      <div class="error-banner" id="error" hidden></div>
      <div class="preset-grid" id="preset-grid"></div>
      <div class="card" style="margin-top:16px;">
        <h2>Custom duration</h2>
        <div class="form-row-inline">
          <input id="custom-hours" type="number" min="${CUSTOM_DURATION_MIN_HOURS}" max="${CUSTOM_DURATION_MAX_HOURS}" step="1" placeholder="Hours (1–240)" />
          <button id="start-custom" type="button" class="btn btn-primary">Start</button>
        </div>
      </div>
      <div class="card" style="margin-top:16px;">
        <label style="display:flex;align-items:center;gap:10px;cursor:pointer;margin-bottom:0;">
          <input id="backdate-toggle" type="checkbox" style="width:auto;" />
          <span>I already started fasting earlier</span>
        </label>
        <div class="form-row" id="backdate-row" hidden style="margin-top:12px;">
          <label for="backdate-time">When did you actually start?</label>
          <input id="backdate-time" type="datetime-local" max="${toDatetimeLocalValue(new Date())}" />
          <p style="margin:6px 0 0;font-size:12px;color:var(--color-text-muted);">Whichever duration you start below, the countdown will begin from this time instead of now. Up to ${MAX_BACKDATE_HOURS / 24} days in the past.</p>
        </div>
      </div>
    </div>
  `;

  const errorEl = container.querySelector<HTMLElement>("#error")!;
  function showError(message: string): void {
    errorEl.textContent = message;
    errorEl.hidden = false;
  }

  const grid = container.querySelector<HTMLElement>("#preset-grid")!;
  for (const preset of FAST_PRESETS) {
    const tile = document.createElement("button");
    tile.type = "button";
    tile.className = "preset-tile";
    tile.innerHTML = `<div class="preset-tile__hours">${preset.hours}h</div><div class="preset-tile__label">${preset.label}</div>`;
    tile.addEventListener("click", () => start(preset.hours));
    grid.appendChild(tile);
  }

  container.querySelector<HTMLButtonElement>("#start-custom")!.addEventListener("click", () => {
    const raw = container.querySelector<HTMLInputElement>("#custom-hours")!.value;
    const hours = parseFloat(raw);
    if (!Number.isFinite(hours) || hours < CUSTOM_DURATION_MIN_HOURS || hours > CUSTOM_DURATION_MAX_HOURS) {
      showError(`Enter a duration between ${CUSTOM_DURATION_MIN_HOURS} and ${CUSTOM_DURATION_MAX_HOURS} hours.`);
      return;
    }
    start(hours);
  });

  const backdateToggle = container.querySelector<HTMLInputElement>("#backdate-toggle")!;
  const backdateRow = container.querySelector<HTMLElement>("#backdate-row")!;
  const backdateInput = container.querySelector<HTMLInputElement>("#backdate-time")!;

  backdateToggle.addEventListener("change", () => {
    backdateRow.hidden = !backdateToggle.checked;
    if (backdateToggle.checked && !backdateInput.value) {
      backdateInput.value = toDatetimeLocalValue(new Date());
    }
  });

  /** Reads the optional backdated start time as an RFC3339 string, or "" for "now". Returns undefined (after showing an error) if invalid. */
  function readStartTime(): string | undefined {
    if (!backdateToggle.checked) return "";
    if (!backdateInput.value) {
      showError("Enter when you actually started, or turn that off to start now.");
      return undefined;
    }
    const chosen = new Date(backdateInput.value);
    if (Number.isNaN(chosen.getTime())) {
      showError("Enter a valid start date and time.");
      return undefined;
    }
    const now = Date.now();
    if (chosen.getTime() > now + 60_000) {
      showError("Start time can't be in the future.");
      return undefined;
    }
    if (chosen.getTime() < now - MAX_BACKDATE_HOURS * 3_600_000) {
      showError(`Start time can't be more than ${MAX_BACKDATE_HOURS / 24} days in the past.`);
      return undefined;
    }
    return chosen.toISOString();
  }

  async function start(hours: number): Promise<void> {
    errorEl.hidden = true;
    const startTime = readStartTime();
    if (startTime === undefined) return;
    try {
      const fast = await api.startFast(hours, startTime);
      setActiveFast(fast);
      ctx.navigate("fasting"); // re-enter mount(), which now finds the active fast
    } catch (err) {
      showError(err instanceof Error ? err.message : "Couldn't start fast.");
    }
  }
}

function mountActive(container: HTMLElement, ctx: ScreenContext, fast: Fast): UnmountFn {
  container.innerHTML = `
    <div class="screen">
      <h1>Fasting</h1>
      <div id="ring-container"></div>
      <p style="text-align:center;margin-top:-16px;font-size:12px;color:var(--color-text-muted);">Started ${formatDateTime(fast.startTime)}</p>
      <div id="stage-callout"></div>
      <div style="text-align:center;margin-bottom:24px;">
        <button id="end-fast" type="button" class="btn btn-block">End fast</button>
      </div>
      <h2>What's happening in your body</h2>
      <ul class="stage-list" id="stage-list"></ul>
      <p class="disclaimer">${FASTING_DISCLAIMER}</p>
    </div>
  `;

  const ring = createRingTimer();
  container.querySelector<HTMLElement>("#ring-container")!.appendChild(ring.element);

  const stageCallout = container.querySelector<HTMLElement>("#stage-callout")!;
  const stageListEl = container.querySelector<HTMLElement>("#stage-list")!;
  const endBtn = container.querySelector<HTMLButtonElement>("#end-fast")!;

  for (const stage of FASTING_STAGES) {
    const li = document.createElement("li");
    li.className = "stage-list__item";
    li.dataset.minHour = String(stage.minHour);
    const hoursLabel = stage.maxHour === null ? `${stage.minHour}h+` : `${stage.minHour}–${stage.maxHour}h`;
    li.innerHTML = `
      <div class="stage-list__hours">${hoursLabel}</div>
      <div>
        <div class="stage-list__title">${stage.title}</div>
        <p class="stage-list__body">${stage.body}</p>
      </div>
    `;
    stageListEl.appendChild(li);
  }

  const startMs = new Date(fast.startTime).getTime();
  const totalMs = fast.plannedDurationHours * 3_600_000;
  const targetEndMs = startMs + totalMs;

  function tick(): void {
    // Always recompute from wall-clock time (never decrement a counter) so
    // background throttling or system sleep self-corrects on the next tick.
    const now = Date.now();
    const remainingMs = targetEndMs - now;
    const elapsedHours = (now - startMs) / 3_600_000;
    const overtime = remainingMs <= 0;

    ring.update({
      fraction: totalMs > 0 ? remainingMs / totalMs : 0,
      timeLabel: overtime ? `+${formatClock(Math.abs(remainingMs) / 1000)}` : formatClock(remainingMs / 1000),
      subLabel: overtime ? `over your ${fast.plannedDurationHours}h goal` : `remaining of ${fast.plannedDurationHours}h`,
      overtime,
    });

    endBtn.className = overtime ? "btn btn-primary btn-block" : "btn btn-block";
    endBtn.textContent = overtime ? "Complete fast" : "End fast early";

    const stage = currentStage(Math.max(0, elapsedHours));
    stageCallout.innerHTML = `
      <div class="stage-callout">
        <div class="stage-callout__title">${stage.title}</div>
        <div>${stage.body}</div>
      </div>
    `;

    stageListEl.querySelectorAll<HTMLElement>(".stage-list__item").forEach((li) => {
      const minHour = Number(li.dataset.minHour);
      li.classList.toggle("is-current", stage.minHour === minHour);
      li.classList.toggle("is-past", minHour < stage.minHour);
    });
  }

  tick();
  const intervalId = window.setInterval(tick, 1000);

  endBtn.addEventListener("click", async () => {
    const reachedGoal = Date.now() >= targetEndMs;
    const result = await promptMoodAndNote(reachedGoal ? "Nice work — fast complete!" : "End this fast early?");
    if (!result) return;

    try {
      if (reachedGoal) {
        await api.completeFast(fast.id, result.mood, result.note);
      } else {
        await api.abandonFast(fast.id, result.mood, result.note);
      }
      setActiveFast(null);
      ctx.navigate("fasting"); // re-enter mount(), which now finds no active fast
    } catch (err) {
      window.alert(err instanceof Error ? err.message : "Couldn't end fast.");
    }
  });

  return () => clearInterval(intervalId);
}
