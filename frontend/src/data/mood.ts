// 5-point mood scale captured when a fast ends (completed or abandoned).
export interface MoodOption {
  value: number; // 1..5, stored directly in fasts.mood
  emoji: string;
  label: string;
}

export const MOOD_OPTIONS: MoodOption[] = [
  { value: 1, emoji: "😩", label: "Struggled" },
  { value: 2, emoji: "😕", label: "Tough" },
  { value: 3, emoji: "🙂", label: "Okay" },
  { value: 4, emoji: "😀", label: "Good" },
  { value: 5, emoji: "🤩", label: "Great" },
];

export function moodEmoji(value: number | undefined): string {
  return MOOD_OPTIONS.find((m) => m.value === value)?.emoji ?? "";
}

export function moodLabel(value: number | undefined): string {
  return MOOD_OPTIONS.find((m) => m.value === value)?.label ?? "";
}
