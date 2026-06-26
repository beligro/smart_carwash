import React, { useEffect, useState, useRef, Suspense, lazy } from 'react';
import { Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom';
import GuestApiService, { getGuestToken, clearGuestToken } from './GuestApiService';
import styles from './WebApp.module.css';

const Header = lazy(() => import('./components/Header'));
const WashInfo = lazy(() => import('./components/WashInfo/WashInfo'));
const GuestPaymentPage = lazy(() => import('./GuestPaymentPage'));

const BASE = '/web/guest';

const GuestApp = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [washInfo, setWashInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [carwashStatus, setCarwashStatus] = useState(null);
  const queuePollingRef = useRef(null);
  const sessionPollingRef = useRef(null);
  const sessionResetTimer = useRef(null);

  const isTerminal = (s) =>
    s && ['complete', 'canceled', 'expired', 'payment_failed'].includes(s.status);

  const clearTimers = () => {
    if (queuePollingRef.current) clearInterval(queuePollingRef.current);
    if (sessionPollingRef.current) clearInterval(sessionPollingRef.current);
    if (sessionResetTimer.current) clearTimeout(sessionResetTimer.current);
    queuePollingRef.current = sessionPollingRef.current = sessionResetTimer.current = null;
  };

  // После завершения/отмены сессии: сброс на форму новой мойки
  const scheduleTerminalReset = (session) => {
    if (!session || !isTerminal(session)) return;
    if (sessionResetTimer.current) return;
    sessionResetTimer.current = setTimeout(() => {
      clearGuestToken();
      setWashInfo((prev) => (prev ? { ...prev, userSession: null, payment: null } : prev));
      sessionResetTimer.current = null;
    }, 5000);
  };

  const startSessionPolling = () => {
    if (sessionPollingRef.current) clearInterval(sessionPollingRef.current);
    sessionPollingRef.current = setInterval(async () => {
      const token = getGuestToken();
      if (!token) return;
      try {
        const data = await GuestApiService.getSession(token);
        if (data?.session) {
          setWashInfo((prev) => ({ ...prev, userSession: data.session, payment: data.payment }));
          if (isTerminal(data.session)) {
            clearInterval(sessionPollingRef.current);
            sessionPollingRef.current = null;
            scheduleTerminalReset(data.session);
          }
        }
      } catch {}
    }, 3000);
  };

  const fetchQueue = async () => {
    try {
      const q = await GuestApiService.getQueueStatus();
      setWashInfo((prev) => ({
        ...prev,
        allBoxes: q?.boxes || [],
        washQueue: q?.wash_queue || { queue_size: 0, has_queue: false },
        airDryQueue: q?.air_dry_queue || { queue_size: 0, has_queue: false },
        vacuumQueue: q?.vacuum_queue || { queue_size: 0, has_queue: false },
      }));
    } catch {}
  };

  // Инициализация
  useEffect(() => {
    const load = async () => {
      try {
        setCarwashStatus(await GuestApiService.getCarwashStatus());
      } catch {
        setCarwashStatus({ is_closed: false });
      }

      await fetchQueue();

      const token = getGuestToken();
      if (token) {
        try {
          const data = await GuestApiService.getSession(token);
          if (data?.session && !isTerminal(data.session)) {
            setWashInfo((prev) => ({ ...prev, userSession: data.session, payment: data.payment }));
            startSessionPolling();
          } else {
            clearGuestToken();
          }
        } catch {
          clearGuestToken();
        }
      }

      queuePollingRef.current = setInterval(fetchQueue, 10000);
      setLoading(false);
    };
    load();
    return clearTimers;
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleCreateSession = async (serviceData) => {
    if (carwashStatus?.is_closed) {
      setError('Мойка временно закрыта');
      return;
    }
    try {
      const resp = await GuestApiService.createSession(serviceData);
      setWashInfo((prev) => ({ ...prev, userSession: resp.session, payment: resp.payment }));
      navigate(`${BASE}/payment`, { state: { session: resp.session, payment: resp.payment } });
    } catch (err) {
      const msg = err.response?.data?.error || err.message || 'Ошибка создания сессии';
      alert(msg);
    }
  };

  const handleCancelSession = async () => {
    try {
      await GuestApiService.cancelSession();
      clearGuestToken();
      setWashInfo((prev) => (prev ? { ...prev, userSession: null, payment: null } : prev));
      navigate(BASE, { replace: true });
    } catch (err) {
      throw err;
    }
  };

  const updateSession = (s, p) =>
    setWashInfo((prev) => ({ ...prev, userSession: s, payment: p !== undefined ? p : prev?.payment }));

  const handlePaymentComplete = (updatedSession) => {
    updateSession(updatedSession);
    startSessionPolling();
    navigate(BASE, { replace: true });
  };

  const handlePaymentFailed = (updatedSession) => {
    updateSession(updatedSession);
    navigate(BASE, { replace: true });
  };

  const showBack = location.pathname !== BASE && location.pathname !== `${BASE}/`;

  if (loading) {
    return <div style={{ padding: 24, textAlign: 'center' }}>Загрузка…</div>;
  }

  if (carwashStatus?.is_closed) {
    return (
      <div className={styles.root}>
        <div className={styles.appContainer}>
          <Suspense fallback={null}>
            <Header theme="light" onBack={showBack ? () => navigate(BASE) : undefined} />
          </Suspense>
          <div className={styles.content} style={{ padding: 40, textAlign: 'center' }}>
            Мойка временно закрыта. Справки: 287-03-78
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.root}>
      <div className={styles.appContainer}>
        <Suspense fallback={null}>
          <Header theme="light" onBack={showBack ? () => navigate(BASE) : undefined} />
        </Suspense>
        <div className={styles.content}>
          <Suspense fallback={<div style={{ padding: 24 }}>Загрузка…</div>}>
            <Routes>
              <Route
                path="/"
                element={
                  error ? (
                    <p style={{ color: 'red', padding: 16 }}>{error}</p>
                  ) : washInfo ? (
                    <WashInfo
                      washInfo={washInfo}
                      theme="light"
                      apiService={GuestApiService}
                      user={null}
                      basePath={BASE}
                      onCreateSession={handleCreateSession}
                      onCancelSession={handleCancelSession}
                      onStartSession={(s, p) => updateSession(s, p)}
                      onChemistryEnabled={(s, p) => updateSession(s, p)}
                      onCompleteSession={(s, p) => updateSession(s, p)}
                    />
                  ) : (
                    <p style={{ padding: 16 }}>Загрузка…</p>
                  )
                }
              />
              <Route
                path="/payment"
                element={
                  <GuestPaymentPage
                    session={location.state?.session || washInfo?.userSession}
                    payment={location.state?.payment || washInfo?.payment}
                    paymentType={location.state?.paymentType || 'main'}
                    onPaymentComplete={handlePaymentComplete}
                    onPaymentFailed={handlePaymentFailed}
                    onBack={() => navigate(BASE)}
                    basePath={BASE}
                  />
                }
              />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Suspense>
        </div>
      </div>
    </div>
  );
};

export default GuestApp;
