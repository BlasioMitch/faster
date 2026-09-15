import { MOOD_OPTIONS } from "../data/mood";

export interface MoodNoteResult {
  mood: number; // 0 = not set
  note: string;
}

/** Shows the mood + optional note modal used whenever a fast ends. Resolves to null if cancelled. */
export function promptMoodAndNote(title: string, confirmLabel = "Save"): Promise<MoodNoteResult | null> {
  return new Promise((resolve) => {
    let selectedMood = 0;

    const overlay = document.createElement("div");
    overlay.className = "modal-overlay";
    overlay.innerHTML = `
      <div class="modal">
        <h3>${title}</h3>
        <p>How did it go?</p>
        <div class="mood-picker"></div>
        <div class="form-row">
          <label for="fast-note">Notes (optional)</label>
          <textarea id="fast-note" placeholder="Anything worth remembering about this fast..."></textarea>
        </div>
        <div class="form-row-inline">
          <button type="button" class="btn btn-ghost" data-action="cancel">Cancel</button>
          <button type="button" class="btn btn-primary" data-action="confirm">${confirmLabel}</button>
        </div>
      </div>
    `;

    const moodPicker = overlay.querySelector<HTMLElement>(".mood-picker")!;
    for (const opt of MOOD_OPTIONS) {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "mood-picker__btn";
      btn.textContent = opt.emoji;
      btn.title = opt.label;
      btn.addEventListener("click", () => {
        selectedMood = opt.value;
        moodPicker.querySelectorAll(".mood-picker__btn").forEach((el) => el.classList.remove("selected"));
        btn.classList.add("selected");
      });
      moodPicker.appendChild(btn);
    }

    const noteEl = overlay.querySelector<HTMLTextAreaElement>("#fast-note")!;

    function close(result: MoodNoteResult | null) {
      document.body.removeChild(overlay);
      resolve(result);
    }

    overlay.querySelector('[data-action="cancel"]')!.addEventListener("click", () => close(null));
    overlay.querySelector('[data-action="confirm"]')!.addEventListener("click", () =>
      close({ mood: selectedMood, note: noteEl.value.trim() }),
    );

    document.body.appendChild(overlay);
  });
}
