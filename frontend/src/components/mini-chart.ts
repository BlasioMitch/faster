// Hand-rolled SVG polyline chart — no charting library. Reports only ever
// needs a few dozen points at most (weight logs, mood-over-time), so a
// ~50-line component is plenty and avoids the bundle/memory cost of a real
// charting dependency.
export interface ChartPoint {
  x: number; // typically a timestamp in ms
  y: number;
}

const VIEW_W = 300;
const VIEW_H = 90;
const PAD = 10;

export function renderMiniChart(points: ChartPoint[]): SVGSVGElement {
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("viewBox", `0 0 ${VIEW_W} ${VIEW_H}`);
  svg.setAttribute("preserveAspectRatio", "none");
  svg.classList.add("mini-chart");

  if (points.length < 2) return svg; // caller shows an empty-state message instead

  const xs = points.map((p) => p.x);
  const ys = points.map((p) => p.y);
  const minX = Math.min(...xs);
  const maxX = Math.max(...xs);
  const minY = Math.min(...ys);
  const maxY = Math.max(...ys);
  const rangeX = maxX - minX || 1;
  const rangeY = maxY - minY || 1;

  const w = VIEW_W - PAD * 2;
  const h = VIEW_H - PAD * 2;

  const coords = points.map((p) => {
    const px = PAD + ((p.x - minX) / rangeX) * w;
    const py = PAD + h - ((p.y - minY) / rangeY) * h;
    return [px, py] as const;
  });

  const linePoints = coords.map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`).join(" ");

  // Subtle fill under the line for a bit of visual life, staying monochromatic (accent, low opacity).
  const areaPoints = `${PAD},${VIEW_H - PAD} ${linePoints} ${VIEW_W - PAD},${VIEW_H - PAD}`;
  const area = document.createElementNS("http://www.w3.org/2000/svg", "polygon");
  area.setAttribute("points", areaPoints);
  area.style.fill = "var(--color-accent-dim)";
  area.style.opacity = "0.25";
  svg.appendChild(area);

  const line = document.createElementNS("http://www.w3.org/2000/svg", "polyline");
  line.setAttribute("points", linePoints);
  line.setAttribute("fill", "none");
  line.setAttribute("stroke-linejoin", "round");
  line.setAttribute("stroke-linecap", "round");
  line.style.stroke = "var(--color-accent-strong)";
  line.style.strokeWidth = "2.5";
  svg.appendChild(line);

  for (const [x, y] of coords) {
    const dot = document.createElementNS("http://www.w3.org/2000/svg", "circle");
    dot.setAttribute("cx", x.toFixed(1));
    dot.setAttribute("cy", y.toFixed(1));
    dot.setAttribute("r", "2.5");
    dot.style.fill = "var(--color-accent-strong)";
    svg.appendChild(dot);
  }

  return svg;
}
