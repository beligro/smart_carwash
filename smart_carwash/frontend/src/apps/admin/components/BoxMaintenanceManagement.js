import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';
import { LAYOUT } from '../../web/components/BoxMap/BoxMap.jsx';
import CloseTicketModal from './CloseTicketModal';

const BOX_TYPE_LABEL = { wash: 'Мойка', vacuum: 'Пылесос', air: 'Воздух' };

// Физическая геометрия здания (повторяет константы публичной схемы /status,
// сам публичный BoxMap.jsx не трогаем — переиспользуем только LAYOUT).
const BOX_W = 4.5;
const BAY_DEPTH = 6;
const BUILDING_W = 46.5;
const BUILDING_H = 24;
const PAD = 2;
const VB_W = BUILDING_W + PAD * 2;
const VB_H = BUILDING_H + PAD * 2;
const TOP_Y = 0;
const BOT_Y = BUILDING_H - BAY_DEPTH;
const WALL_T = 1.5;
const AIR_CY = BUILDING_H / 2; // центр проезда — там метки воздуха
const AIR_R = 1.6;

// Позиции колонок (как в публичной схеме BoxMap): нужны для меток воздуха.
const slotX = (i) => (i < 2 ? i * BOX_W : 2 * BOX_W + WALL_T + (i - 2) * BOX_W);
const AIR_POINTS = [
  { number: 21, cx: slotX(2) },
  { number: 22, cx: slotX(6) },
  { number: 23, cx: slotX(8) },
];

const TO_LIMIT = 500;

const WORK_GREEN = '#16a34a';
const SERVICE_RED = '#dc2626';

// Палитра статусов ТО
const STATUS_COLORS = {
  ok: { fill: '#16a34a', text: '#ffffff', label: 'В норме (< 450 мч)' },
  soon: { fill: '#f59e0b', text: '#1f2937', label: 'Скоро ТО (450–499 мч)' },
  overdue: { fill: '#dc2626', text: '#ffffff', label: 'Просрочено (≥ 500 мч)' },
  working: { fill: WORK_GREEN, text: '#ffffff', label: 'В работе' },
  service: { fill: SERVICE_RED, text: '#ffffff', label: 'В сервисе / сломан' },
  none: { fill: '#cbd5e1', text: '#475569', label: 'Нет данных' },
};

const REASON_OPTIONS = [
  'Плановое ТО (500 мч)',
  'Поломка / ремонт',
  'Замена масла',
  'Другое',
];

const Container = styled.div`
  background: ${p => p.theme.cardBackground};
  border: 1px solid ${p => p.theme.borderColor};
  border-radius: 8px;
  padding: 16px;
`;

const HeaderRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 12px;
`;

const Title = styled.h2`
  margin: 0;
`;

const Button = styled.button`
  padding: 8px 14px;
  background: ${p => (p.primary ? p.theme.primaryColor : p.theme.cardBackground)};
  color: ${p => (p.primary ? 'white' : p.theme.textColor)};
  border: 1px solid ${p => (p.primary ? p.theme.primaryColor : p.theme.borderColor)};
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.95rem;
  &:disabled { opacity: 0.6; cursor: not-allowed; }
`;

const ErrorText = styled.div`
  color: #d33;
  margin-bottom: 8px;
`;

const SuccessText = styled.div`
  color: #2e7d32;
  margin-bottom: 8px;
`;

const MapWrap = styled.div`
  width: 100%;
  background: ${p => p.theme.backgroundColor};
  border: 1px solid ${p => p.theme.borderColor};
  border-radius: 8px;
  padding: 8px;
  svg { width: 100%; height: auto; display: block; }
`;

const Legend = styled.ul`
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  list-style: none;
  padding: 12px 4px 0;
  margin: 0;
  font-size: 0.9rem;
  li { display: flex; align-items: center; gap: 6px; }
`;

const Chip = styled.span`
  width: 14px;
  height: 14px;
  border-radius: 3px;
  display: inline-block;
  background: ${p => p.color};
`;

// ===== Модалка =====
const Overlay = styled.div`
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
`;

const Modal = styled.div`
  background: ${p => p.theme.cardBackground};
  color: ${p => p.theme.textColor};
  border-radius: 10px;
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
  padding: 20px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
`;

const ModalHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
`;

const CloseButton = styled.button`
  background: none;
  border: none;
  color: ${p => p.theme.textColor};
  font-size: 1.5rem;
  cursor: pointer;
  line-height: 1;
`;

const InfoGrid = styled.div`
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 6px 12px;
  margin-bottom: 16px;
  font-size: 0.95rem;
  div:nth-child(odd) { color: ${p => p.theme.textColor}; opacity: 0.7; }
  div:nth-child(even) { font-weight: 600; }
`;

const Label = styled.label`
  display: block;
  font-size: 0.9rem;
  margin: 10px 0 4px;
  font-weight: 600;
`;

const Select = styled.select`
  width: 100%;
  padding: 8px 10px;
  border: 1px solid ${p => p.theme.borderColor};
  background: ${p => p.theme.backgroundColor};
  color: ${p => p.theme.textColor};
  border-radius: 6px;
`;

const TextArea = styled.textarea`
  width: 100%;
  padding: 8px 10px;
  border: 1px solid ${p => p.theme.borderColor};
  background: ${p => p.theme.backgroundColor};
  color: ${p => p.theme.textColor};
  border-radius: 6px;
  min-height: 60px;
  resize: vertical;
  box-sizing: border-box;
`;

const HistoryTitle = styled.h4`
  margin: 18px 0 8px;
`;

const HistoryList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
`;

const HistoryItem = styled.div`
  border: 1px solid ${p => p.theme.borderColor};
  border-radius: 6px;
  padding: 8px 10px;
  font-size: 0.85rem;
`;

const ModalActions = styled.div`
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 18px;
`;

const ConfirmBox = styled.div`
  background: ${p => p.theme.backgroundColor};
  border: 1px solid ${p => p.theme.borderColor};
  border-radius: 6px;
  padding: 12px;
  margin-top: 12px;
`;

const SectionTitle = styled.h3`
  margin: 22px 0 10px;
`;

const BoxTable = styled.table`
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
  th, td {
    text-align: left;
    padding: 8px 10px;
    border-bottom: 1px solid ${p => p.theme.borderColor};
    vertical-align: top;
  }
  th { color: ${p => p.theme.textColor}; opacity: 0.7; font-weight: 600; }
  tr.in-service { background: rgba(220, 38, 38, 0.06); }
`;

const SmallButton = styled.button`
  padding: 6px 12px;
  background: ${p => p.theme.primaryColor};
  color: #fff;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  white-space: nowrap;
  &:disabled { opacity: 0.6; cursor: not-allowed; }
`;

const pad2 = (n) => String(n).padStart(2, '0');

const formatDate = (iso) => {
  if (!iso) return '—';
  try {
    const d = new Date(iso);
    return `${pad2(d.getDate())}.${pad2(d.getMonth() + 1)}.${d.getFullYear()}`;
  } catch {
    return '—';
  }
};

const formatDateTime = (iso) => {
  if (!iso) return '—';
  try {
    const d = new Date(iso);
    return `${pad2(d.getDate())}.${pad2(d.getMonth() + 1)}.${d.getFullYear()} ${pad2(d.getHours())}:${pad2(d.getMinutes())}`;
  } catch {
    return '—';
  }
};

// Длительность простоя в формате HH:MM:SS
const formatDuration = (fromIso, nowMs) => {
  if (!fromIso) return '—';
  const start = new Date(fromIso).getTime();
  if (Number.isNaN(start)) return '—';
  let s = Math.max(0, Math.floor((nowMs - start) / 1000));
  const h = Math.floor(s / 3600);
  s -= h * 3600;
  const m = Math.floor(s / 60);
  s -= m * 60;
  return `${pad2(h)}:${pad2(m)}:${pad2(s)}`;
};

const formatServiceMinutes = (mins) => {
  if (mins === null || mins === undefined) return '—';
  const total = Number(mins);
  if (Number.isNaN(total)) return '—';
  const h = Math.floor(total / 60);
  const m = total % 60;
  if (h > 0) return `${h} ч ${m} мин`;
  return `${m} мин`;
};

const BoxMaintenanceManagement = () => {
  const theme = getTheme('light');
  const [boxes, setBoxes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [selectedNumber, setSelectedNumber] = useState(null);
  const [nowMs, setNowMs] = useState(Date.now());
  const [closingTicket, setClosingTicket] = useState(null);

  // Состояние формы отметки ТО
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [reason, setReason] = useState(REASON_OPTIONS[0]);
  const [comment, setComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [feedback, setFeedback] = useState('');

  // Состояние формы ротации аппаратов
  const [swapOpen, setSwapOpen] = useState(false);
  const [swapA, setSwapA] = useState('');
  const [swapB, setSwapB] = useState('');
  const [swapComment, setSwapComment] = useState('');
  const [swapSubmitting, setSwapSubmitting] = useState(false);
  const [swapFeedback, setSwapFeedback] = useState('');

  const isMounted = useRef(true);

  const load = useCallback(async (silent = false) => {
    if (!silent) setLoading(true);
    try {
      const data = await ApiService.getBoxMaintenance();
      if (!isMounted.current) return;
      setBoxes(Array.isArray(data) ? data : (data?.boxes || []));
      setError('');
    } catch (e) {
      if (!isMounted.current) return;
      setError('Не удалось загрузить данные ТО. Попробуйте обновить.');
    } finally {
      if (isMounted.current && !silent) setLoading(false);
    }
  }, []);

  useEffect(() => {
    isMounted.current = true;
    load();
    const poll = setInterval(() => load(true), 8000);
    return () => {
      isMounted.current = false;
      clearInterval(poll);
    };
  }, [load]);

  // Локальный тик раз в секунду для живого счётчика простоя
  useEffect(() => {
    const id = setInterval(() => setNowMs(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  const byNumber = useMemo(() => {
    const map = new Map();
    for (const b of boxes) map.set(b.box_number ?? b.boxNumber, b);
    return map;
  }, [boxes]);

  const selected = selectedNumber != null ? byNumber.get(selectedNumber) : null;

  const boxNumbers = useMemo(
    () => boxes
      .map((b) => b.box_number ?? b.boxNumber)
      .filter((n) => n != null)
      .sort((a, b) => a - b),
    [boxes],
  );

  const paletteFor = (data) => {
    if (!data) return STATUS_COLORS.none;
    if (data.in_service) return STATUS_COLORS.service;
    if (!data.has_maintenance) return STATUS_COLORS.working; // пылесос/воздух в работе
    return STATUS_COLORS[data.status] || STATUS_COLORS.none;
  };

  const openBox = (number) => {
    if (!byNumber.get(number)) return; // нет учёта ТО (пылесос/воздух)
    setSelectedNumber(number);
    setConfirmOpen(false);
    setReason(REASON_OPTIONS[0]);
    setComment('');
    setFeedback('');
  };

  const closeModal = () => {
    setSelectedNumber(null);
    setConfirmOpen(false);
  };

  const openSwap = () => {
    setSwapOpen(true);
    setSwapA('');
    setSwapB('');
    setSwapComment('');
    setSwapFeedback('');
  };

  const closeSwap = () => {
    setSwapOpen(false);
  };

  const submitSwap = async () => {
    const a = Number(swapA);
    const b = Number(swapB);
    if (!a || !b || a === b) {
      setSwapFeedback('Ошибка: выберите два разных бокса.');
      return;
    }
    setSwapSubmitting(true);
    setSwapFeedback('');
    try {
      await ApiService.swapBoxMaintenance(a, b, swapComment);
      await load(true);
      setSwapOpen(false);
    } catch (e) {
      setSwapFeedback('Ошибка: не удалось выполнить ротацию.');
    } finally {
      setSwapSubmitting(false);
    }
  };

  const submitReset = async () => {
    if (selectedNumber == null) return;
    setSubmitting(true);
    setFeedback('');
    try {
      await ApiService.resetBoxMaintenance(selectedNumber, { reason, comment });
      await load(true);
      setConfirmOpen(false);
      setFeedback('ТО отмечено, счётчик сброшен.');
      setComment('');
    } catch (e) {
      setFeedback('Ошибка: не удалось отметить ТО.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Container theme={theme}>
      <HeaderRow>
        <Title>ТО аппаратов</Title>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          <Button theme={theme} onClick={openSwap} disabled={boxNumbers.length < 2}>
            Ротация аппаратов
          </Button>
          <Button theme={theme} onClick={() => load()} disabled={loading}>
            {loading ? 'Обновление…' : 'Обновить'}
          </Button>
        </div>
      </HeaderRow>

      {error && <ErrorText>{error}</ErrorText>}

      <MapWrap theme={theme}>
        <svg viewBox={`0 0 ${VB_W} ${VB_H}`} preserveAspectRatio="xMidYMid meet" role="img" aria-label="Схема ТО аппаратов">
          <defs>
            <pattern id="serviceStripes" width="1.4" height="1.4" patternTransform="rotate(45)" patternUnits="userSpaceOnUse">
              <rect width="1.4" height="1.4" fill="#6b7280" />
              <rect width="0.7" height="1.4" fill="#4b5563" />
            </pattern>
          </defs>

          <rect x={PAD - 0.3} y={PAD - 0.3} width={BUILDING_W + 0.6} height={BUILDING_H + 0.6} fill="none" stroke="#1f2937" strokeWidth={0.35} />
          <rect x={PAD} y={PAD + BAY_DEPTH} width={BUILDING_W} height={BUILDING_H - 2 * BAY_DEPTH} fill="#e5e7eb" />
          <rect x={PAD + 2 * BOX_W} y={PAD + TOP_Y} width={WALL_T} height={BAY_DEPTH} fill="#1f2937" />
          <rect x={PAD + 2 * BOX_W} y={PAD + BOT_Y} width={WALL_T} height={BAY_DEPTH} fill="#1f2937" />

          {/* Тайлы боксов и пылесосов (у воздуха — отдельная геометрия ниже) */}
          {LAYOUT.filter((s) => s.kind !== 'air').map((slot) => {
            const data = byNumber.get(slot.number);
            const isVacuum = slot.kind === 'vacuum';
            const hasTO = !!(data && data.has_maintenance);
            const inSvc = !!(data && data.in_service);
            const palette = paletteFor(data);
            const fill = palette.fill; // зелёный = работает, красный = в сервисе/сломан
            const clickable = !!data;
            const mh = data ? (data.motor_hours ?? 0) : null;
            const svcRaw = data ? data.service_minutes_30d : null;
            const svcHours = (svcRaw === null || svcRaw === undefined || Number.isNaN(Number(svcRaw)))
              ? null
              : Math.round(Number(svcRaw) / 60);
            const cx = PAD + slot.x + slot.w / 2;
            return (
              <g
                key={slot.number}
                style={{ cursor: clickable ? 'pointer' : 'default' }}
                onClick={() => openBox(slot.number)}
              >
                <rect
                  x={PAD + slot.x + 0.1}
                  y={PAD + slot.y + 0.1}
                  width={slot.w - 0.2}
                  height={slot.h - 0.2}
                  fill={fill}
                  stroke="#0f172a"
                  strokeWidth={0.08}
                  rx={0.25}
                />
                <text x={cx} y={PAD + slot.y + 0.95} textAnchor="middle" fontSize={0.6} fontWeight="700" fill={palette.text} opacity={0.9}>
                  {isVacuum ? 'ПЫЛЕСОС' : `БОКС ${slot.number}`}
                </text>

                {/* По центру: моточасы (мойка), либо номер (пылесос) — как у боксов */}
                {!inSvc && hasTO && (
                  <text x={cx} y={PAD + slot.y + 3.1} textAnchor="middle" fontSize={1.7} fontWeight="800" fill={palette.text}>
                    {mh}
                  </text>
                )}
                {!inSvc && isVacuum && (
                  <text x={cx} y={PAD + slot.y + 3.3} textAnchor="middle" fontSize={1.9} fontWeight="800" fill={palette.text}>
                    {slot.number}
                  </text>
                )}

                {/* В сервисе — таймер простоя по центру */}
                {inSvc && (
                  <text x={cx} y={PAD + slot.y + 3.2} textAnchor="middle" fontSize={0.95} fontWeight="700" fill="#ffffff" textLength={slot.w - 0.8} lengthAdjust="spacingAndGlyphs">
                    {formatDuration(data.in_service_since, nowMs)}
                  </text>
                )}

                {/* Простой за 30 дней — снизу */}
                {svcHours !== null && svcHours > 0 && (
                  <text x={cx} y={PAD + slot.y + slot.h - 0.5} textAnchor="middle" fontSize={0.72} fontWeight="700" fill={inSvc ? '#ffffff' : '#111827'} textLength={slot.w - 0.8} lengthAdjust="spacingAndGlyphs">
                    {`в серв. ${svcHours}ч/30д`}
                  </text>
                )}
              </g>
            );
          })}

          {/* Воздух (21–23) — метки в центральном проезде (в LAYOUT их нет, задаём отдельно) */}
          {AIR_POINTS.map((slot) => {
            const data = byNumber.get(slot.number);
            const inSvc = !!(data && data.in_service);
            const palette = paletteFor(data);
            const cx = PAD + slot.cx;
            const cy = PAD + AIR_CY;
            return (
              <g
                key={slot.number}
                style={{ cursor: data ? 'pointer' : 'default' }}
                onClick={() => openBox(slot.number)}
              >
                <circle cx={cx} cy={cy} r={AIR_R} fill={palette.fill} stroke="#0f172a" strokeWidth={0.12} />
                <text x={cx} y={cy - 0.35} textAnchor="middle" fontSize={0.5} fontWeight="700" fill={palette.text}>ВОЗДУХ</text>
                <text x={cx} y={cy + 0.9} textAnchor="middle" fontSize={1.3} fontWeight="800" fill={palette.text}>{slot.number}</text>
                {inSvc && (
                  <text x={cx} y={cy + AIR_R + 1.2} textAnchor="middle" fontSize={0.85} fontWeight="700" fill="#111827">
                    {formatDuration(data.in_service_since, nowMs)}
                  </text>
                )}
              </g>
            );
          })}
        </svg>
      </MapWrap>

      <Legend>
        {Object.entries(STATUS_COLORS).filter(([k]) => k !== 'none').map(([key, v]) => (
          <li key={key}>
            <Chip color={v.fill} />
            {v.label}
          </li>
        ))}
      </Legend>

      <SectionTitle>Все боксы и наряды</SectionTitle>
      <div style={{ overflowX: 'auto' }}>
        <BoxTable theme={theme}>
          <thead>
            <tr>
              <th>Бокс</th>
              <th>Тип</th>
              <th>Статус</th>
              <th>Моточасы</th>
              <th>Наряд (причина · кто · сколько в сервисе)</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {boxes.slice().sort((a, b) => (a.box_number ?? 0) - (b.box_number ?? 0)).map((b) => {
              const t = b.open_ticket;
              const inSvc = b.in_service;
              return (
                <tr key={b.box_number} className={inSvc ? 'in-service' : ''}>
                  <td>№{b.box_number}</td>
                  <td>{BOX_TYPE_LABEL[b.box_type] || b.box_type || '—'}</td>
                  <td>{inSvc ? 'В сервисе' : 'В работе'}</td>
                  <td>{b.has_maintenance ? `${b.motor_hours ?? 0} мч` : '—'}</td>
                  <td>
                    {t ? (
                      <div>
                        <div style={{ fontWeight: 600 }}>
                          {t.symptom_name || '—'}{t.is_breakdown === false ? ' (не поломка)' : ''}
                        </div>
                        <div style={{ opacity: 0.7, fontSize: '0.82rem' }}>
                          {t.opened_by || '—'} · {formatDuration(t.opened_at, nowMs)}
                        </div>
                        {t.cashier_comment && (
                          <div style={{ opacity: 0.7, fontSize: '0.82rem' }}>«{t.cashier_comment}»</div>
                        )}
                      </div>
                    ) : inSvc ? (
                      <span style={{ opacity: 0.6 }}>в сервисе (без наряда)</span>
                    ) : '—'}
                  </td>
                  <td>
                    {t && (
                      <SmallButton
                        theme={theme}
                        onClick={() => setClosingTicket({ id: t.id, box_number: b.box_number, is_breakdown: t.is_breakdown })}
                      >
                        Закрыть наряд
                      </SmallButton>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </BoxTable>
      </div>

      {closingTicket && (
        <CloseTicketModal
          ticket={closingTicket}
          onClose={() => setClosingTicket(null)}
          onClosed={() => { setClosingTicket(null); load(true); }}
        />
      )}

      {swapOpen && (
        <Overlay onClick={closeSwap}>
          <Modal theme={theme} onClick={(e) => e.stopPropagation()}>
            <ModalHeader>
              <h3 style={{ margin: 0 }}>Ротация аппаратов</h3>
              <CloseButton theme={theme} onClick={closeSwap}>×</CloseButton>
            </ModalHeader>

            {swapFeedback && <ErrorText>{swapFeedback}</ErrorText>}

            <div style={{ fontSize: '0.9rem', opacity: 0.8, marginBottom: 8 }}>
              Выберите два бокса, между которыми переставлены аппараты. Моточасы и дата ТО переедут вместе с аппаратами.
            </div>

            <Label>Первый бокс</Label>
            <Select theme={theme} value={swapA} onChange={(e) => setSwapA(e.target.value)}>
              <option value="">— выберите бокс —</option>
              {boxNumbers.map((n) => (
                <option key={n} value={n} disabled={String(n) === String(swapB)}>№{n}</option>
              ))}
            </Select>

            <Label>Второй бокс</Label>
            <Select theme={theme} value={swapB} onChange={(e) => setSwapB(e.target.value)}>
              <option value="">— выберите бокс —</option>
              {boxNumbers.map((n) => (
                <option key={n} value={n} disabled={String(n) === String(swapA)}>№{n}</option>
              ))}
            </Select>

            {swapA && swapB && swapA !== swapB && (
              <ConfirmBox theme={theme}>
                <div style={{ fontWeight: 600 }}>
                  Поменять местами учёт ТО боксов №{swapA} и №{swapB}?
                </div>
                <div style={{ fontSize: '0.85rem', opacity: 0.85, marginTop: 6 }}>
                  Моточасы и дата ТО переедут вместе с аппаратами.
                </div>
              </ConfirmBox>
            )}

            <Label>Комментарий (необязательно)</Label>
            <TextArea theme={theme} value={swapComment} onChange={(e) => setSwapComment(e.target.value)} placeholder="Доп. информация о перестановке" />

            <ModalActions>
              <Button theme={theme} onClick={closeSwap} disabled={swapSubmitting}>Отмена</Button>
              <Button primary theme={theme} onClick={submitSwap} disabled={swapSubmitting || !swapA || !swapB || swapA === swapB}>
                {swapSubmitting ? 'Выполнение…' : 'Переставить'}
              </Button>
            </ModalActions>
          </Modal>
        </Overlay>
      )}

      {selected && (
        <Overlay onClick={closeModal}>
          <Modal theme={theme} onClick={(e) => e.stopPropagation()}>
            <ModalHeader>
              <h3 style={{ margin: 0 }}>Бокс №{selected.box_number ?? selected.boxNumber}</h3>
              <CloseButton theme={theme} onClick={closeModal}>×</CloseButton>
            </ModalHeader>

            {feedback && (feedback.startsWith('Ошибка') ? <ErrorText>{feedback}</ErrorText> : <SuccessText>{feedback}</SuccessText>)}

            <InfoGrid theme={theme}>
              <div>Тип</div>
              <div>{BOX_TYPE_LABEL[selected.box_type] || selected.box_type || '—'}</div>
              {selected.has_maintenance && (
                <>
                  <div>Моточасы</div>
                  <div>{(selected.motor_hours ?? 0)} / {TO_LIMIT} мч{selected.status === 'overdue' ? ' (просрочено)' : selected.status === 'soon' ? ' (скоро ТО)' : ''}</div>
                  <div>Последнее ТО</div>
                  <div>{formatDate(selected.last_to_at)}</div>
                </>
              )}
              <div>Статус бокса</div>
              <div>{selected.in_service ? 'В сервисе сейчас' : 'В работе'}</div>
              {selected.in_service && (
                <>
                  <div>В сервисе уже</div>
                  <div>{formatDuration(selected.in_service_since, nowMs)}</div>
                </>
              )}
              {selected.open_ticket && (
                <>
                  <div>Причина</div>
                  <div>{(selected.open_ticket.symptom_name || 'не указана')}{selected.open_ticket.is_breakdown === false ? ' (не поломка)' : ''}</div>
                </>
              )}
              {selected.open_ticket && selected.open_ticket.cashier_comment && (
                <>
                  <div>Коммент кассира</div>
                  <div>{selected.open_ticket.cashier_comment}</div>
                </>
              )}
              <div>В сервисе за 30 дней</div>
              <div>{formatServiceMinutes(selected.service_minutes_30d)}</div>
            </InfoGrid>

            {selected.open_ticket && (
              <Button
                primary
                theme={theme}
                style={{ marginBottom: 10 }}
                onClick={() => {
                  setClosingTicket({ id: selected.open_ticket.id, box_number: selected.box_number, is_breakdown: selected.open_ticket.is_breakdown });
                  closeModal();
                }}
              >
                Закрыть наряд
              </Button>
            )}

            {selected.has_maintenance && (
              !confirmOpen ? (
                <Button primary theme={theme} onClick={() => setConfirmOpen(true)}>
                  Отметить плановое ТО
                </Button>
              ) : (
                <ConfirmBox theme={theme}>
                  <div style={{ fontWeight: 600, marginBottom: 8 }}>
                    Точно отметить ТО для бокса №{selected.box_number}?
                  </div>
                  <Label>Причина</Label>
                  <Select theme={theme} value={reason} onChange={(e) => setReason(e.target.value)}>
                    {REASON_OPTIONS.map((r) => (
                      <option key={r} value={r}>{r}</option>
                    ))}
                  </Select>
                  <Label>Комментарий (необязательно)</Label>
                  <TextArea theme={theme} value={comment} onChange={(e) => setComment(e.target.value)} placeholder="Доп. информация" />
                  <ModalActions>
                    <Button theme={theme} onClick={() => setConfirmOpen(false)} disabled={submitting}>Отмена</Button>
                    <Button primary theme={theme} onClick={submitReset} disabled={submitting}>
                      {submitting ? 'Сохранение…' : 'Подтвердить'}
                    </Button>
                  </ModalActions>
                </ConfirmBox>
              )
            )}

            {selected.has_maintenance && (
              <>
                <HistoryTitle>История ТО (последние 5)</HistoryTitle>
                <HistoryList>
                  {(selected.history && selected.history.length > 0) ? (
                    selected.history.map((h, idx) => (
                      <HistoryItem key={idx} theme={theme}>
                        <div style={{ fontWeight: 600 }}>{formatDateTime(h.performed_at)}</div>
                        <div>Кто: {h.performed_by ?? '—'}</div>
                        <div>Было моточасов: {h.motor_hours_at_reset ?? '—'}</div>
                        <div>Причина: {h.reason || '—'}</div>
                        {(h.comment) && <div>Комментарий: {h.comment}</div>}
                      </HistoryItem>
                    ))
                  ) : (
                    <div style={{ opacity: 0.7 }}>Записей пока нет</div>
                  )}
                </HistoryList>
              </>
            )}
          </Modal>
        </Overlay>
      )}
    </Container>
  );
};

export default BoxMaintenanceManagement;
