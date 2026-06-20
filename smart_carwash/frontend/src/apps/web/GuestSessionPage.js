import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import GuestApiService, { getGuestToken, clearGuestToken } from './GuestApiService';

const SERVICES = { wash: 'Мойка', air_dry: 'Обдув', vacuum: 'Пылесос' };
const STATUS_LABELS = {
  created: 'Ожидание оплаты',
  in_queue: 'В очереди',
  assigned: 'Бокс назначен',
  active: 'Идёт мойка',
  complete: 'Завершена',
  canceled: 'Отменена',
  payment_failed: 'Ошибка оплаты',
};
const STATUS_COLORS = {
  in_queue: '#f59e0b',
  assigned: '#3b82f6',
  active: '#22c55e',
  complete: '#6b7280',
  canceled: '#ef4444',
  payment_failed: '#ef4444',
};

const formatTime = (seconds) => {
  if (seconds <= 0) return '0:00';
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, '0')}`;
};

const isTerminal = (status) =>
  ['complete', 'canceled', 'payment_failed'].includes(status);

/**
 * GuestSessionPage — таймер и управление гостевой сессией.
 * Показывает оставшееся время, статус, кнопку продления.
 * После завершения предлагает зарегистрироваться.
 */
const GuestSessionPage = ({ session: sessionProp, payment, onExtend, onComplete }) => {
  const navigate = useNavigate();
  const [session, setSession] = useState(sessionProp);
  const [timeLeft, setTimeLeft] = useState(null);
  const [extendLoading, setExtendLoading] = useState(false);
  const [extendError, setExtendError] = useState(null);
  const [availableTimes, setAvailableTimes] = useState([15, 30]);
  const timerRef = useRef(null);

  // Обновляем session при изменении пропа
  useEffect(() => {
    setSession(sessionProp);
  }, [sessionProp]);

  // Вычисляем оставшееся время
  useEffect(() => {
    if (!session || isTerminal(session.status)) {
      setTimeLeft(null);
      return;
    }
    if (session.status !== 'active' && session.status !== 'in_queue' && session.status !== 'assigned') {
      setTimeLeft(null);
      return;
    }

    const updateTimer = () => {
      const startStr = session.active_started_at || session.status_updated_at;
      if (!startStr) { setTimeLeft(null); return; }
      const start = new Date(startStr);
      const totalMins =
        (session.rental_time_minutes || 0) + (session.extension_time_minutes || 0);
      const totalSec = totalMins * 60;
      const elapsed = Math.floor((Date.now() - start.getTime()) / 1000);
      const left = totalSec - elapsed;
      setTimeLeft(left > 0 ? left : 0);
    };

    updateTimer();
    timerRef.current = setInterval(updateTimer, 1000);
    return () => clearInterval(timerRef.current);
  }, [session]);

  // Загружаем доступные времена продления
  useEffect(() => {
    if (!session) return;
    GuestApiService.getAvailableRentalTimes(session.service_type)
      .then((d) => {
        if (d?.availableTimes?.length) setAvailableTimes(d.availableTimes);
      })
      .catch(() => {});
  }, [session?.service_type]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleExtend = async (minutes) => {
    setExtendError(null);
    setExtendLoading(true);
    try {
      const token = getGuestToken();
      const resp = await GuestApiService.extendSession(minutes, 0, token);
      onExtend?.(resp.session, resp.payment);
    } catch (err) {
      setExtendError(err.response?.data?.error || 'Ошибка продления');
    } finally {
      setExtendLoading(false);
    }
  };

  const handleDone = () => {
    clearGuestToken();
    onComplete?.();
  };

  if (!session) {
    return <div style={{ padding: 24, textAlign: 'center' }}>Сессия не найдена.</div>;
  }

  const statusLabel = STATUS_LABELS[session.status] || session.status;
  const statusColor = STATUS_COLORS[session.status] || '#6b7280';
  const serviceName = SERVICES[session.service_type] || session.service_type;
  const done = isTerminal(session.status);

  return (
    <div style={{ padding: 16, maxWidth: 480, margin: '0 auto' }}>
      {/* Статус */}
      <div style={cardStyle}>
        <p style={{ margin: 0, fontSize: 13, color: '#888' }}>Статус</p>
        <p style={{ margin: '4px 0 0', fontSize: 18, fontWeight: 700, color: statusColor }}>
          {statusLabel}
        </p>
        {session.box_number && (
          <p style={{ margin: '4px 0 0', fontSize: 14, color: '#555' }}>
            Бокс № {session.box_number}
          </p>
        )}
      </div>

      {/* Таймер */}
      {!done && session.status === 'active' && timeLeft !== null && (
        <div style={{ ...cardStyle, textAlign: 'center', marginTop: 12 }}>
          <p style={{ margin: 0, fontSize: 13, color: '#888' }}>Осталось</p>
          <p style={{ margin: '4px 0 0', fontSize: 52, fontWeight: 700, fontVariantNumeric: 'tabular-nums', color: timeLeft < 120 ? '#ef4444' : '#111' }}>
            {formatTime(timeLeft)}
          </p>
          <p style={{ margin: '4px 0 0', fontSize: 13, color: '#888' }}>
            {serviceName}
            {session.car_number ? ` · ${session.car_number}` : ''}
          </p>
        </div>
      )}

      {session.status === 'in_queue' && (
        <div style={{ ...cardStyle, textAlign: 'center', marginTop: 12 }}>
          <p style={{ margin: 0 }}>Ваш автомобиль в очереди.<br />Мойка начнётся автоматически.</p>
        </div>
      )}

      {session.status === 'assigned' && (
        <div style={{ ...cardStyle, textAlign: 'center', marginTop: 12 }}>
          <p style={{ margin: 0 }}>
            Подъезжайте к боксу {session.box_number ? `№ ${session.box_number}` : ''}.
          </p>
        </div>
      )}

      {/* Продление — только если сессия активна */}
      {session.status === 'active' && !done && (
        <div style={{ marginTop: 16 }}>
          <p style={{ fontSize: 14, fontWeight: 600, marginBottom: 8 }}>Продлить мойку:</p>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            {availableTimes.map((min) => (
              <button
                key={min}
                onClick={() => handleExtend(min)}
                disabled={extendLoading}
                style={{ ...btnStyle, flex: '1 1 80px', minWidth: 80 }}
              >
                +{min} мин
              </button>
            ))}
          </div>
          {extendError && (
            <p style={{ color: 'red', fontSize: 13, marginTop: 8 }}>{extendError}</p>
          )}
        </div>
      )}

      {/* После завершения */}
      {done && (
        <div style={{ marginTop: 20 }}>
          <div style={{ ...cardStyle, background: '#f0fdf4', textAlign: 'center' }}>
            <p style={{ fontSize: 16, fontWeight: 600, marginBottom: 8 }}>
              {session.status === 'complete' ? 'Мойка завершена!' : 'Сессия закрыта'}
            </p>
            <p style={{ fontSize: 14, color: '#555', marginBottom: 0 }}>
              Зарегистрируйтесь, чтобы сохранять историю,
              получать уведомления и участвовать в программе лояльности.
            </p>
          </div>
          <a
            href="/web/login"
            style={{ ...btnStyle, display: 'block', textAlign: 'center', textDecoration: 'none', background: '#1a73e8', color: '#fff', marginTop: 12, padding: '14px 0', borderRadius: 10 }}
          >
            Зарегистрироваться
          </a>
          <button onClick={handleDone} style={{ ...btnStyle, marginTop: 8, background: 'transparent', color: '#888' }}>
            Без регистрации
          </button>
        </div>
      )}
    </div>
  );
};

const cardStyle = {
  background: '#f8f9fa',
  borderRadius: 12,
  padding: '16px 20px',
};

const btnStyle = {
  padding: '12px 0',
  border: 'none',
  borderRadius: 10,
  cursor: 'pointer',
  fontSize: 14,
  background: '#e8f0fe',
  color: '#1a73e8',
  fontWeight: 600,
};

export default GuestSessionPage;
