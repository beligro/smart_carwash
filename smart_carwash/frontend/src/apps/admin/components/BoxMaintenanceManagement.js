import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';
import { LAYOUT } from '../../web/components/BoxMap/BoxMap.jsx';

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

const TO_LIMIT = 500;

// Палитра статусов ТО
const STATUS_COLORS = {
  ok: { fill: '#16a34a', text: '#ffffff', label: 'В норме (< 450 мч)' },
  soon: { fill: '#f59e0b', text: '#1f2937', label: 'Скоро ТО (450–499 мч)' },
  overdue: { fill: '#dc2626', text: '#ffffff', label: 'Просрочено (≥ 500 мч)' },
  service: { fill: '#6b7280', text: '#ffffff', label: 'В сервисе' },
  none: { fill: '#cbd5e1', text: '#475569', label: 'Без учёта ТО' },
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

  // Состояние формы отметки ТО
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [reason, setReason] = useState(REASON_OPTIONS[0]);
  const [comment, setComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [feedback, setFeedback] = useState('');

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

  const paletteFor = (data) => {
    if (!data) return STATUS_COLORS.none;
    if (data.in_service) return STATUS_COLORS.service;
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
        <Button theme={theme} onClick={() => load()} disabled={loading}>
          {loading ? 'Обновление…' : 'Обновить'}
        </Button>
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

          {LAYOUT.map((slot) => {
            const data = byNumber.get(slot.number);
            const isVacuum = slot.kind === 'vacuum';
            const palette = paletteFor(data);
            const fill = data && data.in_service ? 'url(#serviceStripes)' : palette.fill;
            const clickable = !!data;
            const mh = data ? (data.motor_hours ?? data.motorHours ?? 0) : null;
            const svcRaw = data ? (data.service_minutes_30d ?? data.serviceMinutes30d) : null;
            const svcHours = (svcRaw === null || svcRaw === undefined || Number.isNaN(Number(svcRaw)))
              ? null
              : Math.round(Number(svcRaw) / 60);
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
                <text x={PAD + slot.x + slot.w / 2} y={PAD + slot.y + 1} textAnchor="middle" fontSize={0.7} fill={palette.text} opacity={0.85}>
                  {isVacuum ? 'ПЫЛЕСОС' : 'БОКС'}
                </text>
                <text x={PAD + slot.x + slot.w / 2} y={PAD + slot.y + slot.h / 2 - 0.4} textAnchor="middle" fontSize={2.0} fontWeight="800" fill={palette.text}>
                  {slot.number}
                </text>
                {data && !data.in_service && (
                  <text x={PAD + slot.x + slot.w / 2} y={PAD + slot.y + slot.h - 1.5} textAnchor="middle" fontSize={1.25} fontWeight="700" fill={palette.text}>
                    {mh}
                  </text>
                )}
                {data && data.in_service && (
                  <text x={PAD + slot.x + slot.w / 2} y={PAD + slot.y + slot.h - 1.5} textAnchor="middle" fontSize={0.95} fontWeight="700" fill="#ffffff">
                    {formatDuration(data.in_service_since || data.inServiceSince, nowMs)}
                  </text>
                )}
                {data && svcHours !== null && (
                  <text x={PAD + slot.x + slot.w / 2} y={PAD + slot.y + slot.h - 0.4} textAnchor="middle" fontSize={0.8} fontWeight="600" fill={data.in_service ? '#ffffff' : palette.text} opacity={0.85}>
                    {`серв ${svcHours}ч/30д`}
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

      {selected && (
        <Overlay onClick={closeModal}>
          <Modal theme={theme} onClick={(e) => e.stopPropagation()}>
            <ModalHeader>
              <h3 style={{ margin: 0 }}>Бокс №{selected.box_number ?? selected.boxNumber}</h3>
              <CloseButton theme={theme} onClick={closeModal}>×</CloseButton>
            </ModalHeader>

            {feedback && (feedback.startsWith('Ошибка') ? <ErrorText>{feedback}</ErrorText> : <SuccessText>{feedback}</SuccessText>)}

            <InfoGrid theme={theme}>
              <div>Моточасы</div>
              <div>{(selected.motor_hours ?? 0)} / {TO_LIMIT} мч{selected.status === 'overdue' ? ' (просрочено)' : selected.status === 'soon' ? ' (скоро ТО)' : ''}</div>
              <div>Последнее ТО</div>
              <div>{formatDate(selected.last_to_at ?? selected.lastToAt)}</div>
              <div>Статус бокса</div>
              <div>{selected.in_service ? 'В сервисе сейчас' : 'В работе'}</div>
              {selected.in_service && (
                <>
                  <div>В сервисе уже</div>
                  <div>{formatDuration(selected.in_service_since ?? selected.inServiceSince, nowMs)}</div>
                </>
              )}
              <div>В сервисе за последние 30 дней</div>
              <div>{formatServiceMinutes(selected.service_minutes_30d ?? selected.serviceMinutes30d)}</div>
            </InfoGrid>

            {!confirmOpen ? (
              <Button primary theme={theme} onClick={() => setConfirmOpen(true)}>
                ТО выполнено
              </Button>
            ) : (
              <ConfirmBox theme={theme}>
                <div style={{ fontWeight: 600, marginBottom: 8 }}>
                  Точно отметить ТО для бокса №{selected.box_number ?? selected.boxNumber}?
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
            )}

            <HistoryTitle>История ТО (последние 5)</HistoryTitle>
            <HistoryList>
              {(selected.history && selected.history.length > 0) ? (
                selected.history.map((h, idx) => (
                  <HistoryItem key={idx} theme={theme}>
                    <div style={{ fontWeight: 600 }}>{formatDateTime(h.performed_at ?? h.performedAt)}</div>
                    <div>Кто: {h.performed_by ?? h.performedBy ?? '—'}</div>
                    <div>Было моточасов: {h.motor_hours_at_reset ?? h.motorHoursAtReset ?? '—'}</div>
                    <div>Причина: {h.reason || '—'}</div>
                    {(h.comment) && <div>Комментарий: {h.comment}</div>}
                  </HistoryItem>
                ))
              ) : (
                <div style={{ opacity: 0.7 }}>Записей пока нет</div>
              )}
            </HistoryList>
          </Modal>
        </Overlay>
      )}
    </Container>
  );
};

export default BoxMaintenanceManagement;
