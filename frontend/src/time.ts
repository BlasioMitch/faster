// Native Date/Intl only — no date library, matching the low-memory,
// no-framework approach used throughout the frontend.

function pad(n: number): string {
  return String(n).padStart(2, "0");
}

/** Formats a non-negative duration in seconds as "HH:MM:SS", or "Nd HH:MM:SS" once it reaches a full day (fasts run up to 120h). */
export function formatClock(totalSeconds: number): string {
  const abs = Math.max(0, Math.floor(totalSeconds));
  const days = Math.floor(abs / 86400);
  const hours = Math.floor((abs % 86400) / 3600);
  const minutes = Math.floor((abs % 3600) / 60);
  const seconds = abs % 60;
  const clock = `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
  return days > 0 ? `${days}d ${clock}` : clock;
}

/** Formats an ISO/RFC3339 date string as a short local date, e.g. "Mar 4, 2026". */
export function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

/** Formats an ISO/RFC3339 date string as a short local date + time. */
export function formatDateTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString(undefined, { year: "numeric", month: "short", day: "numeric", hour: "numeric", minute: "2-digit" });
}

/** Formats a Date as the "YYYY-MM-DDTHH:mm" value a `<input type="datetime-local">` expects, in local time. */
export function toDatetimeLocalValue(date: Date): string {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}
