// Circular countdown: a plain SVG ring, updated once per second. No
// requestAnimationFrame loop and no animation library — the CSS
// `transition: stroke-dashoffset 1s linear` (see style.css .ring-progress)
// lets the browser compositor interpolate smoothly between the once-a-second
// writes, which is the whole point of this technique for staying low-CPU.
const RADIUS = 45;
const CIRCUMFERENCE = 2 * Math.PI * RADIUS;

export interface RingTimerUpdate {
  /** remaining/total, 1 = just started (ring full), 0 = time's up (ring drained). */
  fraction: number;
  timeLabel: string;
  subLabel: string;
  overtime?: boolean;
}

export interface RingTimerHandle {
  element: HTMLElement;
  update(opts: RingTimerUpdate): void;
}

export function createRingTimer(): RingTimerHandle {
  const wrapper = document.createElement("div");
  wrapper.className = "ring-timer";
  wrapper.innerHTML = `
    <svg viewBox="0 0 100 100">
      <circle class="ring-track" cx="50" cy="50" r="${RADIUS}" fill="none" stroke-width="8" />
      <circle class="ring-progress" cx="50" cy="50" r="${RADIUS}" fill="none" stroke-width="8"
              stroke-dasharray="${CIRCUMFERENCE}" stroke-dashoffset="0" transform="rotate(-90 50 50)" />
    </svg>
    <div class="ring-timer__label">
      <span class="ring-timer__time">00:00:00</span>
      <span class="ring-timer__sub"></span>
    </div>
  `;

  const progress = wrapper.querySelector<SVGCircleElement>(".ring-progress")!;
  const timeEl = wrapper.querySelector<HTMLElement>(".ring-timer__time")!;
  const subEl = wrapper.querySelector<HTMLElement>(".ring-timer__sub")!;

  return {
    element: wrapper,
    update({ fraction, timeLabel, subLabel, overtime }) {
      const clamped = Math.min(1, Math.max(0, fraction));
      // Ring starts full (offset 0) and drains toward empty (offset = full
      // circumference) as remaining time runs out — "full to empty",
      // showing remaining, not elapsed.
      progress.style.strokeDashoffset = String(CIRCUMFERENCE * (1 - clamped));
      progress.classList.toggle("is-overtime", Boolean(overtime));
      timeEl.textContent = timeLabel;
      subEl.textContent = subLabel;
    },
  };
}
