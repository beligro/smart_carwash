import React, { useEffect, useState, useRef, Suspense, lazy } from 'react';
import { Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom';
import GuestApiService, { getGuestToken, clearGuestToken } from './GuestApiService';
import styles from './WebApp.module.css';

const Header = lazy(() => import('./components/Header'));
const GuestBookingPage = lazy(() => import('./GuestBookingPage'));
const GuestPaymentPage = lazy(() => import('./GuestPaymentPage'));
const GuestSessionPage = lazy(() => import('./GuestSessionPage'));

const BASE = '/web/guest';

const GuestApp = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [session, setSession] = useState(null);
  const [payment, setPayment] = useState(null);
  const [loading, setLoading] = useState(true);
  const [carwashStatus, setCarwashStatus] = useState(null);
  const sessionPollingRef = useRef(null);

  const isTerminal = (s) =>
    s && ['complete', 'canceled', 'expired', 'payment_failed'].includes(s.status);

  const startSessionPolling = (token) => {
    if (sessionPollingRef.current) clearInterval(sessionPollingRef.current);
    sessionPollingRef.current = setInterval(async () => {
      try {
        const data = await GuestApiService.getSession(token);
        if (data?.session) {
          setSession(data.session);
          setPayment(data.payment);
          if (isTerminal(data.session)) {
            clearInterval(sessionPollingRef.current);
            sessionPollingRef.current = null;
          }
        }
      } catch {}
    }, 3000);
  };

  // При загрузке: проверяем наличие активной сессии в cookie
  useEffect(() => {
    const load = async () => {
      try {
        const status = await GuestApiService.getCarwashStatus();
        setCarwashStatus(status);
      } catch {
        setCarwashStatus({ is_closed: false });
      }

      const token = getGuestToken();
      if (token) {
        try {
          const data = await GuestApiService.getSession(token);
          if (data?.session && !isTerminal(data.session)) {
            setSession(data.session);
            setPayment(data.payment);
            startSessionPolling(token);
          } else {
            // Сессия завершена — чистим cookie
            clearGuestToken();
          }
        } catch {
          clearGuestToken();
        }
      }
      setLoading(false);
    };
    load();
    return () => {
      if (sessionPollingRef.current) clearInterval(sessionPollingRef.current);
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // Редирект: если есть активная сессия и мы на странице букинга — идём на сессию
  useEffect(() => {
    if (!loading && session && !isTerminal(session)) {
      const isOnBooking =
        location.pathname === BASE ||
        location.pathname === `${BASE}/` ||
        location.pathname === `${BASE}/booking`;
      if (isOnBooking) {
        navigate(`${BASE}/session`, { replace: true });
      }
    }
  }, [loading, session, location.pathname, navigate]);

  const handleCreateSession = async (serviceData) => {
    if (carwashStatus?.is_closed) return;
    try {
      const resp = await GuestApiService.createSession(serviceData);
      setSession(resp.session);
      setPayment(resp.payment);
      navigate(`${BASE}/payment`, {
        state: { session: resp.session, payment: resp.payment },
      });
    } catch (err) {
      alert(err.response?.data?.error || err.message || 'Ошибка создания сессии');
    }
  };

  const handlePaymentComplete = (updatedSession) => {
    setSession(updatedSession);
    const token = getGuestToken();
    if (token) startSessionPolling(token);
    navigate(`${BASE}/session`, { replace: true });
  };

  const handlePaymentFailed = (updatedSession) => {
    setSession(updatedSession);
    navigate(`${BASE}/session`, { replace: true });
  };

  const handleSessionExtended = (updatedSession, newPayment) => {
    setSession(updatedSession);
    setPayment(newPayment);
    navigate(`${BASE}/payment`, {
      state: { session: updatedSession, payment: newPayment, paymentType: 'extension' },
    });
  };

  const handleSessionComplete = () => {
    clearGuestToken();
    setSession(null);
    setPayment(null);
    navigate(BASE, { replace: true });
  };

  const showBack =
    location.pathname !== BASE && location.pathname !== `${BASE}/`;

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
          <Header
            theme="light"
            onBack={showBack ? () => navigate(BASE) : undefined}
          />
        </Suspense>
        <div className={styles.content}>
          <Suspense fallback={<div style={{ padding: 24 }}>Загрузка…</div>}>
            <Routes>
              <Route
                path="/"
                element={<GuestBookingPage onCreateSession={handleCreateSession} />}
              />
              <Route path="/booking" element={<GuestBookingPage onCreateSession={handleCreateSession} />} />
              <Route
                path="/payment"
                element={
                  <GuestPaymentPage
                    session={location.state?.session || session}
                    payment={location.state?.payment || payment}
                    paymentType={location.state?.paymentType || 'main'}
                    onPaymentComplete={handlePaymentComplete}
                    onPaymentFailed={handlePaymentFailed}
                    onBack={() => navigate(BASE)}
                    basePath={BASE}
                  />
                }
              />
              <Route
                path="/session"
                element={
                  session ? (
                    <GuestSessionPage
                      session={session}
                      payment={payment}
                      onExtend={handleSessionExtended}
                      onComplete={handleSessionComplete}
                    />
                  ) : (
                    <Navigate to={BASE} replace />
                  )
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
