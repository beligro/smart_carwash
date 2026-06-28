import React, { useEffect, useRef, useState } from 'react';
import axios from 'axios';
import BoxMap from './components/BoxMap/BoxMap';
import logoUrl from './assets/h2o-logo.webp';

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
        <header style={headerStyle}>
          <a href="/" style={logoLinkStyle} aria-label="На главную H2O">
            <img src={logoUrl} alt="H2O — автомойка самообслуживания" style={logoImgStyle} />
          </a>
          <a href="/" style={homeLinkStyle}>← На главную</a>
        </header>

        <h1 style={titleStyle}>Автомойка H2O — статус боксов</h1>

        {/* Общая очередь */}
        <div style={hasQueue ? bannerQueueStyle : bannerFreeStyle}>
          {hasQueue
            ? `В очереди ${totalCars} ${plural(totalCars, 'машина', 'машины', 'машин')} · ожидание ~${waitMin} мин`
            : 'Свободно — заехать можно сразу после оплаты'}
        </div>

        {/* CTA — записаться на мойку */}
        <div style={ctaWrapStyle}>
          <a href="/web/login" style={ctaButtonStyle}>Помыть машину</a>
          <div style={ctaHintStyle}>Без регистрации, вход или регистрация — на следующем шаге</div>
        </div>

        {error && (
          <p style={{ color: '#ff8a8a', textAlign: 'center' }}>Не удалось загрузить статус, повтор…</p>
        )}

        <BoxMap boxes={boxes} />
      </div>
    </div>
  );
};

const pageStyle = {
  minHeight: '100vh',
  padding: '24px 16px',
  background:
    'radial-gradient(1200px 600px at 80% -10%, rgba(25,217,255,0.20), transparent 60%),' +
    'radial-gradient(900px 500px at 0% 10%, rgba(80,120,255,0.14), transparent 60%),' +
    'linear-gradient(180deg, #0b1a2b 0%, #0a1626 100%)',
  color: '#f3f7ff',
  boxSizing: 'border-box',
  fontFamily: 'Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif',
};
const headerStyle = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  gap: 16,
  marginBottom: 20,
};
const logoLinkStyle = { display: 'inline-flex', alignItems: 'center', textDecoration: 'none' };
const logoImgStyle = { height: 48, width: 'auto', display: 'block' };
const homeLinkStyle = {
  color: '#9fb2c8', textDecoration: 'none', fontSize: '0.95rem', fontWeight: 600,
  border: '1px solid rgba(255,255,255,0.08)', borderRadius: 999, padding: '8px 16px',
  background: 'rgba(255,255,255,0.04)',
};
const titleStyle = { textAlign: 'center', fontSize: '1.6rem', fontWeight: 800, color: '#f3f7ff', margin: '0 0 16px', letterSpacing: '-0.02em' };
const bannerBase = {
  textAlign: 'center', fontSize: '1.1rem', fontWeight: 700, borderRadius: 14,
  padding: '14px 20px', marginBottom: 16, border: '1px solid rgba(255,255,255,0.08)',
};
const bannerQueueStyle = { ...bannerBase, background: 'rgba(245,158,11,0.14)', color: '#ffd591', borderColor: 'rgba(245,158,11,0.35)' };
const bannerFreeStyle = { ...bannerBase, background: 'rgba(34,197,94,0.14)', color: '#86efac', borderColor: 'rgba(34,197,94,0.35)' };
const ctaWrapStyle = { textAlign: 'center', margin: '0 0 18px' };
const ctaButtonStyle = {
  display: 'inline-block', background: 'linear-gradient(135deg, #19d9ff 0%, #0aa5e8 100%)',
  color: '#0b1a2b', textDecoration: 'none', fontWeight: 800, fontSize: '1.2rem',
  padding: '16px 44px', borderRadius: 14, boxShadow: '0 20px 60px -20px rgba(25,217,255,0.7)',
};
const ctaHintStyle = { marginTop: 10, fontSize: '0.9rem', color: '#9fb2c8' };

export default StatusBoard;
