import { useEffect, useMemo, useState } from "react";
import "./BoxMap.css";

/* =========================================================================
 * Геометрия мойки (в метрах) — здание 24 × 46,5 м.
 * Ряд боксов сверху: пылесосы 19,20 → стена 1.5 м → боксы 11..18
 * Ряд боксов снизу: пылесосы 9,10  → стена 1.5 м → боксы 1..8
 * Точки сжатого воздуха 21,22,23 — на колоннах в центральном проезде,
 *   теперь тоже статусные (подсвечиваются), таймер рисуется снизу.
 * Координаты в "метрах" (x, y, w, h); компонент масштабирует во viewBox SVG.
 * ========================================================================= */

const BOX_W = 4.5;
const BAY_DEPTH = 6;
const WALL_T = 1.5;
const BUILDING_W = 46.5;
const BUILDING_H = 24;

const slotX = (i) => (i < 2 ? i * BOX_W : 2 * BOX_W + WALL_T + (i - 2) * BOX_W);

const TOP_Y = 0;
const BOT_Y = BUILDING_H - BAY_DEPTH;
const AIR_R = 1.1; // радиус метки воздуха

export const LAYOUT = [
  // Нижний ряд
  { number: 9,  kind: "vacuum", x: slotX(0), y: BOT_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 10, kind: "vacuum", x: slotX(1), y: BOT_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 1,  kind: "box",    x: slotX(2), y: BOT_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 2,  kind: "box",    x: slotX(3), y: BOT_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 3,  kind: "box",    x: slotX(4), y: BOT_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 4,  kind: "box",    x: slotX(5), y: BOT_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 5,  kind: "box",    x: slotX(6), y: BOT_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 6,  kind: "box",    x: slotX(7), y: BOT_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 7,  kind: "box",    x: slotX(8), y: BOT_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 8,  kind: "box",    x: slotX(9), y: BOT_Y, w: BOX_W, h: BAY_DEPTH },
  // Верхний ряд
  { number: 19, kind: "vacuum", x: slotX(0), y: TOP_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 20, kind: "vacuum", x: slotX(1), y: TOP_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 11, kind: "box",    x: slotX(2), y: TOP_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 12, kind: "box",    x: slotX(3), y: TOP_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 13, kind: "box",    x: slotX(4), y: TOP_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 14, kind: "box",    x: slotX(5), y: TOP_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 15, kind: "box",    x: slotX(6), y: TOP_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 16, kind: "box",    x: slotX(7), y: TOP_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 17, kind: "box",    x: slotX(8), y: TOP_Y, w: BOX_W, h: BAY_DEPTH },
  { number: 18, kind: "box",    x: slotX(9), y: TOP_Y, w: BOX_W, h: BAY_DEPTH },
];

// Точки воздуха — теперь статусные (cx в метрах, cy = центр проезда)
const AIR_POINTS = [
  { number: 21, kind: "air", cx: slotX(2) },
  { number: 22, kind: "air", cx: slotX(6) },
  { number: 23, kind: "air", cx: slotX(8) },
];
const AIR_CY = BUILDING_H / 2;

const STATUS_COLORS = {
  free:        { fill: "#16a34a", text: "#ffffff", label: "Свободен" },
  busy:        { fill: "#dc2626", text: "#ffffff", label: "Идёт мойка" },
  reserved:    { fill: "#f59e0b", text: "#1f2937", label: "Назначен" },
  maintenance: { fill: "#6b7280", text: "#ffffff", label: "В сервисе" },
  cleaning:    { fill: "#38bdf8", text: "#0b3a52", label: "Уборка" },
};
const COOLDOWN = { fill: "#a855f7", text: "#ffffff", label: "Кулдаун" };
const NEUTRAL = { fill: "#cbd5e1", text: "#475569", label: "Нет данных" };

const fmt = (s) => {
  const v = Math.max(0, Math.floor(s));
  const m = Math.floor(v / 60);
  const sec = v % 60;
  return `${m}:${sec.toString().padStart(2, "0")}`;
};

function BoxMap({ boxes = [] }) {
  const byNumber = useMemo(() => {
    const map = new Map();
    for (const b of boxes) map.set(b.number, b);
    return map;
  }, [boxes]);

  // Локальные таймеры: для каждого бокса храним {secondsLeft, cooldownLeft}
  const [timers, setTimers] = useState({});

  // Пересинхронизация при обновлении данных (включая рост при продлении)
  useEffect(() => {
    const init = {};
    for (const b of boxes) {
      init[b.number] = {
        secondsLeft: typeof b.secondsLeft === "number" ? b.secondsLeft : null,
        cooldownLeft: typeof b.cooldownSecondsLeft === "number" ? b.cooldownSecondsLeft : null,
      };
    }
    setTimers(init);
  }, [boxes]);

  // Локальный тик раз в секунду (плавность между опросами)
  useEffect(() => {
    const id = setInterval(() => {
      setTimers((prev) => {
        const next = {};
        let changed = false;
        for (const k in prev) {
          const t = prev[k];
          const nt = { secondsLeft: t.secondsLeft, cooldownLeft: t.cooldownLeft };
          if (typeof t.secondsLeft === "number" && t.secondsLeft > 0) { nt.secondsLeft = t.secondsLeft - 1; changed = true; }
          if (typeof t.cooldownLeft === "number" && t.cooldownLeft > 0) { nt.cooldownLeft = t.cooldownLeft - 1; changed = true; }
          next[k] = nt;
        }
        return changed ? next : prev;
      });
    }, 1000);
    return () => clearInterval(id);
  }, []);

  const PAD = 2;
  const VB_W = BUILDING_W + PAD * 2;
  const VB_H = BUILDING_H + PAD * 2;

  // Определяет палитру: кулдаун имеет приоритет визуализации над free
  const paletteFor = (data, t) => {
    const status = data?.status;
    const cdLeft = t?.cooldownLeft;
    if (typeof cdLeft === "number" && cdLeft > 0 && (status === "free" || !status)) {
      return { palette: COOLDOWN, isCooldown: true };
    }
    return { palette: (status && STATUS_COLORS[status]) || NEUTRAL, isCooldown: false };
  };

  return (
    <div className="boxmap">
      <div className="boxmap__canvas">
        <svg
          viewBox={`0 0 ${VB_W} ${VB_H}`}
          preserveAspectRatio="xMidYMid meet"
          role="img"
          aria-label="План автомойки со статусами боксов"
        >
          <rect x={PAD - 0.3} y={PAD - 0.3} width={BUILDING_W + 0.6} height={BUILDING_H + 0.6} fill="none" stroke="#1f2937" strokeWidth={0.35} />
          <rect x={PAD} y={PAD + BAY_DEPTH} width={BUILDING_W} height={BUILDING_H - 2 * BAY_DEPTH} fill="#e5e7eb" />
          <rect x={PAD + 2 * BOX_W} y={PAD + TOP_Y} width={WALL_T} height={BAY_DEPTH} fill="#1f2937" />
          <rect x={PAD + 2 * BOX_W} y={PAD + BOT_Y} width={WALL_T} height={BAY_DEPTH} fill="#1f2937" />

          {/* Боксы / пылесосы */}
          {LAYOUT.map((slot) => {
            const data = byNumber.get(slot.number);
            const t = timers[slot.number];
            const { palette, isCooldown } = paletteFor(data, t);
            const isBusy = data?.status === "busy";
            const secs = isCooldown ? t?.cooldownLeft : t?.secondsLeft;
            const showTimer = (isBusy && typeof t?.secondsLeft === "number") || isCooldown;

            return (
              <g key={slot.number} data-box={slot.number} data-status={data?.status || "unknown"} className={isBusy ? "boxmap__slot boxmap__slot--busy" : "boxmap__slot"}>
                <rect x={PAD + slot.x + 0.1} y={PAD + slot.y + 0.1} width={slot.w - 0.2} height={slot.h - 0.2} fill={palette.fill} stroke="#0f172a" strokeWidth={0.08} rx={0.25} />
                <text x={PAD + slot.x + slot.w / 2} y={PAD + slot.y + 1} textAnchor="middle" fontSize={0.7} fill={palette.text} opacity={0.85}>
                  {slot.kind === "vacuum" ? "ПЫЛЕСОС" : "БОКС"}
                </text>
                <text x={PAD + slot.x + slot.w / 2} y={PAD + slot.y + slot.h / 2 + 0.4} textAnchor="middle" fontSize={2.4} fontWeight="800" fill={palette.text}>
                  {slot.number}
                </text>
                {showTimer && (
                  <text x={PAD + slot.x + slot.w / 2} y={PAD + slot.y + slot.h - 0.6} textAnchor="middle" fontSize={1.1} fontWeight="700" fill={palette.text}>
                    {fmt(secs ?? 0)}
                  </text>
                )}
              </g>
            );
          })}

          {/* Точки воздуха — статусные, таймер снизу */}
          {AIR_POINTS.map((p) => {
            const data = byNumber.get(p.number);
            const t = timers[p.number];
            const { palette, isCooldown } = paletteFor(data, t);
            const isBusy = data?.status === "busy";
            const secs = isCooldown ? t?.cooldownLeft : t?.secondsLeft;
            const showTimer = (isBusy && typeof t?.secondsLeft === "number") || isCooldown;
            return (
              <g key={p.number} data-box={p.number} data-status={data?.status || "unknown"} className={isBusy ? "boxmap__slot boxmap__slot--busy" : "boxmap__slot"}>
                <circle cx={PAD + p.cx} cy={PAD + AIR_CY} r={AIR_R} fill={palette.fill} stroke="#ffffff" strokeWidth={0.15} />
                <text x={PAD + p.cx} y={PAD + AIR_CY + 0.4} textAnchor="middle" fontSize={1.1} fontWeight="800" fill={palette.text}>
                  {p.number}
                </text>
                {showTimer && (
                  <text x={PAD + p.cx} y={PAD + AIR_CY + AIR_R + 1.1} textAnchor="middle" fontSize={1.0} fontWeight="700" fill="#0f172a">
                    {fmt(secs ?? 0)}
                  </text>
                )}
              </g>
            );
          })}
        </svg>
      </div>

      <ul className="boxmap__legend">
        {Object.entries(STATUS_COLORS).map(([key, v]) => (
          <li key={key}>
            <span className="boxmap__chip" style={{ background: v.fill }} />
            {v.label}
          </li>
        ))}
        <li><span className="boxmap__chip" style={{ background: COOLDOWN.fill }} />{COOLDOWN.label}</li>
        <li><span className="boxmap__chip" style={{ background: NEUTRAL.fill }} />{NEUTRAL.label}</li>
      </ul>
    </div>
  );
}

export default BoxMap;
