import { api, type Fast } from "../api";
import { renderMiniChart } from "../components/mini-chart";
import { moodEmoji } from "../data/mood";
import type { Screen } from "../router";
import { formatClock, formatDate } from "../time";
import { formatWeight } from "../units";

export const reportsScreen: Screen = {
  async mount(container, ctx) {
    container.innerHTML = `<div class="screen"><p>Loading…</p></div>`;

    const hasProfile = await api.hasProfile();
    if (!hasProfile) {
      ctx.navigate("onboarding");
      return;
    }

    const [stats, fasts, weightLogs, profile] = await Promise.all([
      api.getStats(),
      api.listFasts(100, 0),
      api.listWeightLogs(),
      api.getProfile(),
    ]);

    const weightUnit = (profile?.weightUnit as "kg" | "lb") ?? "kg";

    container.innerHTML = `
      <div class="screen">
        <h1>Reports</h1>

        <div class="stat-grid">
          <div class="stat-card"><div class="stat-card__value">${stats.totalFasts}</div><div class="stat-card__label">Total fasts</div></div>
          <div class="stat-card"><div class="stat-card__value">${stats.currentStreakDays}</div><div class="stat-card__label">Day streak</div></div>
          <div class="stat-card"><div class="stat-card__value">${stats.completionRatePercent.toFixed(0)}%</div><div class="stat-card__label">Completion rate</div></div>
          <div class="stat-card"><div class="stat-card__value">${stats.longestFastHours.toFixed(0)}h</div><div class="stat-card__label">Longest fast</div></div>
          <div class="stat-card"><div class="stat-card__value">${stats.averageDurationHours.toFixed(0)}h</div><div class="stat-card__label">Average duration</div></div>
        </div>

        <div class="card">
          <h2>Weight trend</h2>
          <div id="weight-chart"></div>
        </div>

        <div class="card">
          <h2>Mood over time</h2>
          <div id="mood-chart"></div>
        </div>

        <div class="card">
          <h2>History</h2>
          <div id="history-list"></div>
        </div>
      </div>
    `;

    const weightChartEl = container.querySelector<HTMLElement>("#weight-chart")!;
    if (weightLogs.length >= 2) {
      const points = weightLogs.map((w) => ({ x: new Date(w.loggedAt).getTime(), y: w.weightKg }));
      weightChartEl.appendChild(renderMiniChart(points));

      const delta = weightLogs[weightLogs.length - 1].weightKg - weightLogs[0].weightKg;
      const note = document.createElement("p");
      note.style.cssText = "margin-top:8px;font-size:12px;color:var(--color-text-muted);";
      note.textContent = `${delta >= 0 ? "+" : "-"}${formatWeight(Math.abs(delta), weightUnit)} since first log`;
      weightChartEl.appendChild(note);
    } else {
      weightChartEl.innerHTML = `<p class="empty-state" style="padding:16px 0;">Log weight from the Profile screen to see a trend here.</p>`;
    }

    const moodChartEl = container.querySelector<HTMLElement>("#mood-chart")!;
    const moodFasts = fasts.filter((f): f is Fast & { mood: number; endTime: string } => Boolean(f.mood && f.endTime));
    if (moodFasts.length >= 2) {
      // fasts are newest-first; the chart reads left-to-right as oldest-first.
      const points = moodFasts
        .slice()
        .reverse()
        .map((f) => ({ x: new Date(f.endTime).getTime(), y: f.mood }));
      moodChartEl.appendChild(renderMiniChart(points));
    } else {
      moodChartEl.innerHTML = `<p class="empty-state" style="padding:16px 0;">Complete a few fasts with a mood logged to see a trend here.</p>`;
    }

    const historyEl = container.querySelector<HTMLElement>("#history-list")!;
    if (fasts.length === 0) {
      historyEl.innerHTML = `<p class="empty-state">No fasts yet — start one from the Fasting tab.</p>`;
    } else {
      for (const fast of fasts) {
        historyEl.appendChild(renderHistoryItem(fast));
      }
    }
  },
};

export default reportsScreen;

function renderHistoryItem(fast: Fast): HTMLElement {
  const item = document.createElement("div");
  item.className = "history-item";
  const durationLabel = fast.actualDurationSeconds != null ? formatClock(fast.actualDurationSeconds) : "—";
  const badgeClass = fast.status === "completed" ? "badge-completed" : "badge-abandoned";
  item.innerHTML = `
    <div>
      <div class="history-item__date">${formatDate(fast.startTime)}</div>
      <div class="history-item__meta">${durationLabel} of ${fast.plannedDurationHours}h goal</div>
    </div>
    <div>
      <span class="badge ${badgeClass}">${fast.status}</span>
      <span class="mood-emoji">${moodEmoji(fast.mood)}</span>
    </div>
  `;
  return item;
}
