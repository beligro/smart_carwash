import React, { useEffect, useRef, useState } from 'react';
import axios from 'axios';
import BoxMap from './components/BoxMap/BoxMap';

const baseURL = process.env.REACT_APP_API_URL || '/api';
const POLL_MS = 4000;

// /queue-status -> пропсы BoxMap + общая очередь
function mapData(data) {
  const now = Date.now();
  const boxes = (data?.all_boxes || []).map((b) => {
    const box = { number: b.number, status: b.status };
    if (typeof b.seconds_left === 'number') box.secondsLeft = b.seconds_left;
    if (typeof b.reserved_seconds_left === 'number') box.reservedSecondsLeft = b.reserved_seconds_left;
    if (b.cooldown_until) {
      const left = Math.floor((new Date(b.cooldown_until).getTime() - now) / 1000);
      if (left > 0) box.cooldownSecondsLeft = left;
    }
    return box;
  });

  const hasQueue = !!data?.has_any_queue;
  const totalCars = data?.total_queue_size || 0;
  const waitMin = Math.max(
    data?.wash_queue?.wait_time_minutes || 0,
    data?.air_dry_queue?.wait_time_minutes || 0,
    data?.vacuum_queue?.wait_time_minutes || 0
  );
  return { boxes, hasQueue, totalCars, waitMin };
}

const plural = (n, one, few, many) => {
  const m10 = n % 10, m100 = n % 100;
  if (m10 === 1 && m100 !== 11) return one;
  if (m10 >= 2 && m10 <= 4 && (m100 < 10 || m100 >= 20)) return few;
  return many;
};

const StatusBoard = () => {
  const [state, setState] = useState({ boxes: [], hasQueue: false, totalCars: 0, waitMin: 0 });
  const [error, setError] = useState(false);
  const pollRef = useRef(null);

  useEffect(() => {
    let stopped = false;
    const fetchStatus = async () => {
      try {
        const res = await axios.get(`${baseURL}/queue-status`, { timeout: 8000 });
        if (!stopped) { setState(mapData(res.data)); setError(false); }
      } catch {
        if (!stopped) setError(true);
      }
    };
    fetchStatus();
    pollRef.current = setInterval(fetchStatus, POLL_MS);
    return () => { stopped = true; if (pollRef.current) clearInterval(pollRef.current); };
  }, []);

  const { boxes, hasQueue, totalCars, waitMin } = state;

  return (
    <div style={pageStyle}>
      <div style={{ maxWidth: 1100, margin: '0 auto' }}>
        <h1 style={titleStyle}>Автомойка H2O — статус боксов</h1>

        {/* Общая очередь */}
        <div style={hasQueue ? bannerQueueStyle : bannerFreeStyle}>
          {hasQueue
            ? `В очереди ${totalCars} ${plural(totalCars, 'машина', 'машины', 'машин')} · ожидание ~${waitMin} мин`
            : 'Свободно — заехать можно сразу после оплаты'}
        </div>

        {/* CTA — записаться на мойку */}
        <div style={ctaWrapStyle}>
          <a href="/web/login" style={ctaButtonStyle}>🚗 Помыть машину</a>
          <div style={ctaHintStyle}>Без регистрации, вход или регистрация — на следующем шаге</div>
        </div>

        {error && (
          <p style={{ color: '#c62828', textAlign: 'center' }}>Не удалось загрузить статус, повтор…</p>
        )}

        <BoxMap boxes={boxes} />
      </div>
    </div>
  );
};

const pageStyle = {
  minHeight: '100vh',
  padding: '24px 16px',
  background: 'linear-gradient(160deg, #e0eaf0 0%, #eef3f7 100%)',
  boxSizing: 'border-box',
};
const titleStyle = { textAlign: 'center', fontSize: '1.5rem', color: '#1f2937', margin: '0 0 16px' };
const bannerBase = {
  textAlign: 'center', fontSize: '1.1rem', fontWeight: 700, borderRadius: 12,
  padding: '14px 20px', marginBottom: 16,
};
const bannerQueueStyle = { ...bannerBase, background: '#fff3cd', color: '#7a5c00' };
const bannerFreeStyle = { ...bannerBase, background: '#e8f5e9', color: '#2e7d32' };
const ctaWrapStyle = { textAlign: 'center', margin: '0 0 18px' };
const ctaButtonStyle = {
  display: 'inline-block', background: 'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
  color: '#fff', textDecoration: 'none', fontWeight: 800, fontSize: '1.2rem',
  padding: '16px 40px', borderRadius: 14, boxShadow: '0 4px 16px rgba(2,132,199,0.4)',
};
const ctaHintStyle = { marginTop: 8, fontSize: '0.9rem', color: '#64748b' };

export default StatusBoard;
