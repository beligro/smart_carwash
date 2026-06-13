import React, { useEffect, useState, Suspense, lazy, useRef } from 'react';
import { Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom';
import WebApiService from '../../shared/services/WebApiService';
import styles from './WebApp.module.css';

const Header = lazy(() => import('./components/Header'));
const WashInfo = lazy(() => import('./components/WashInfo/WashInfo'));
const PaymentPage = lazy(() => import('./components/PaymentPage'));
const SessionDetails = lazy(() => import('./components/SessionDetails'));
const SessionHistory = lazy(() => import('./components/SessionHistory'));
const BookingPage = lazy(() => import('./components/BookingPage'));
const PaymentResultPage = lazy(() => import('./PaymentResultPage'));

const BASE = '/web';

const WebApp = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [user, setUser] = useState(null);
  const [washInfo, setWashInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [carwashStatus, setCarwashStatus] = useState(null);
  const [carwashStatusLoading, setCarwashStatusLoading] = useState(true);
  const queuePollingInterval = useRef(null);
  const sessionPollingInterval = useRef(null);
  const sessionResetTimer = useRef(null);

  const token = typeof window !== 'undefined' ? localStorage.getItem('web_token') : null;

  useEffect(() => {
    if (!token) {
      navigate(`${BASE}/login`, { replace: true });
      return;
    }
    const loadUser = async () => {
      try {
        const me = await WebApiService.getMe();
        setUser(me);
      } catch (e) {
        navigate(`${BASE}/login`, { replace: true });
      } finally {
        setLoading(false);
      }
    };
    loadUser();
  }, [token, navigate]);

  useEffect(() => {
    const load = async () => {
      try {
        const st = await WebApiService.getCarwashStatus();
        setCarwashStatus(st);
      } catch {
        setCarwashStatus({ is_closed: false });
      } finally {
        setCarwashStatusLoading(false);
      }
    };
    load();
  }, []);

  const clearPollingIntervals = () => {
    if (queuePollingInterval.current) clearInterval(queuePollingInterval.current);
    queuePollingInterval.current = null;
    if (sessionPollingInterval.current) clearInterval(sessionPollingInterval.current);
    sessionPollingInterval.current = null;
    if (sessionResetTimer.current) clearTimeout(sessionResetTimer.current);
    sessionResetTimer.current = null;
  };

  const checkAndResetTerminalSession = (session) => {
    if (!session || !['complete', 'canceled', 'expired', 'payment_failed'].includes(session.status)) return;
    if (location.pathname !== BASE && location.pathname !== `${BASE}/`) return;
    if (sessionResetTimer.current) return;
    sessionResetTimer.current = setTimeout(() => {
      setWashInfo((prev) => (prev ? { ...prev, userSession: null, payment: null } : null));
      sessionResetTimer.current = null;
    }, 5000);
  };

  const startSessionPolling = (sessionId) => {
    if (sessionPollingInterval.current) clearInterval(sessionPollingInterval.current);
    sessionPollingInterval.current = setInterval(async () => {
      try {
        const data = await WebApiService.getSessionById(sessionId);
        if (data?.session) {
          setWashInfo((prev) => (prev ? { ...prev, userSession: data.session, payment: data.payment } : null));
          if (['complete', 'canceled', 'expired', 'payment_failed'].includes(data.session.status)) {
            clearInterval(sessionPollingInterval.current);
            sessionPollingInterval.current = null;
            checkAndResetTerminalSession(data.session);
          }
        }
      } catch {}
    }, 2000);
  };

  const fetchQueueStatus = async (isInitial = false) => {
    if (!user) return;
    try {
      if (isInitial) setLoading(true);
      const response = await WebApiService.getQueueStatus();
      setWashInfo((prev) => ({
        ...prev,
        allBoxes: response?.boxes || [],
        washQueue: response?.wash_queue || { queue_size: 0, has_queue: false },
        airDryQueue: response?.air_dry_queue || { queue_size: 0, has_queue: false },
        vacuumQueue: response?.vacuum_queue || { queue_size: 0, has_queue: false },
      }));
      if (isInitial) {
        try {
          const sessionResp = await WebApiService.getUserSession();
          if (sessionResp?.session) {
            setWashInfo((prev) => ({ ...prev, userSession: sessionResp.session, payment: sessionResp.payment }));
            const s = sessionResp.session;
            if (['complete', 'canceled', 'expired', 'payment_failed'].includes(s.status)) {
              checkAndResetTerminalSession(s);
            } else {
              startSessionPolling(s.id);
            }
          }
        } catch {}
      }
      setError(null);
    } catch (e) {
      if (isInitial) setError('Ошибка загрузки');
      setWashInfo((prev) => prev || { allBoxes: [], washQueue: { queue_size: 0, has_queue: false }, airDryQueue: { queue_size: 0, has_queue: false }, vacuumQueue: { queue_size: 0, has_queue: false }, userSession: null, payment: null });
    } finally {
      if (isInitial) setLoading(false);
    }
  };

  const startQueuePolling = () => {
    if (queuePollingInterval.current) clearInterval(queuePollingInterval.current);
    queuePollingInterval.current = setInterval(() => fetchQueueStatus(false), 10000);
  };

  useEffect(() => {
    if (!user || carwashStatus?.is_closed) return;
    fetchQueueStatus(true).then(() => startQueuePolling());
  }, [user, carwashStatus]);

  useEffect(() => () => clearPollingIntervals(), []);

  const handleCreateSessionWithPayment = async (serviceData) => {
    if (carwashStatus?.is_closed) {
      setError('Мойка временно закрыта');
      return;
    }
    if (!user) return;
    setLoading(true);
    setError(null);
    try {
      const requestData = {
        userId: user.id,
        serviceType: serviceData.serviceType,
        withChemistry: serviceData.withChemistry,
        chemistryTimeMinutes: serviceData.chemistryTimeMinutes || 0,
        carNumber: serviceData.carNumber,
        carNumberCountry: serviceData.carNumberCountry || 'RUS',
        rentalTimeMinutes: serviceData.rentalTimeMinutes,
      };
      const response = await WebApiService.createSessionWithPayment(requestData);
      if (response?.session && response?.payment) {
        if (sessionPollingInterval.current) clearInterval(sessionPollingInterval.current);
        setWashInfo((prev) => ({ ...prev, userSession: response.session, payment: response.payment }));
        startSessionPolling(response.session.id);
        navigate(`${BASE}/payment`, { state: { session: response.session, payment: response.payment } });
      } else {
        setError('Ошибка создания сессии');
      }
    } catch (err) {
      const msg = err.response?.data?.error || err.message || 'Ошибка создания сессии';
      if (msg.includes('активн')) {
        const sessionResp = await WebApiService.getUserSession();
        if (sessionResp?.session) {
          setWashInfo((prev) => ({ ...prev, userSession: sessionResp.session, payment: sessionResp.payment }));
          navigate(BASE);
        }
      } else {
        setError(msg);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCancelSession = async (sessionId) => {
    try {
      await WebApiService.cancelSession(sessionId);
      const data = await WebApiService.getSessionById(sessionId);
      setWashInfo((prev) => (prev ? { ...prev, userSession: data?.session, payment: data?.payment } : null));
      if (sessionPollingInterval.current) clearInterval(sessionPollingInterval.current);
      startQueuePolling();
    } catch (e) {
      throw e;
    }
  };

  const handlePaymentComplete = (updatedSession) => {
    if (sessionPollingInterval.current) clearInterval(sessionPollingInterval.current);
    setWashInfo((prev) => ({ ...prev, userSession: updatedSession, payment: updatedSession?.payment }));
    startSessionPolling(updatedSession?.id);
    navigate(BASE);
  };

  const handlePaymentFailed = (updatedSession) => {
    if (sessionPollingInterval.current) clearInterval(sessionPollingInterval.current);
    setWashInfo((prev) => ({ ...prev, userSession: updatedSession, payment: updatedSession?.payment }));
    startSessionPolling(updatedSession?.id);
    navigate(BASE);
  };

  const showBack = location.pathname !== BASE && location.pathname !== `${BASE}/`;

  const handleLogout = async () => {
    try {
      await WebApiService.logout();
    } finally {
      navigate(`${BASE}/login`, { replace: true });
      window.location.reload();
    }
  };

  const handleLinkTelegram = async () => {
    try {
      const res = await WebApiService.linkTelegramRequest();
      const link = res?.link;
      if (link) {
        window.open(link, '_blank', 'noopener,noreferrer');
        const me = await WebApiService.getMe();
        setUser(me);
      }
    } catch (e) {
      setError(e.response?.data?.error || e.message || 'Не удалось получить ссылку');
    }
  };

  if (!token) return null;
  if (loading && !user) return <div style={{ padding: 24, textAlign: 'center' }}>Загрузка…</div>;

  return (
    <div className={styles.root}>
      <div className={styles.appContainer}>
        <Suspense fallback={<div className={styles.content}>Загрузка…</div>}>
          <Header
            theme="light"
            onBack={showBack ? () => navigate(BASE) : undefined}
            onLogout={handleLogout}
          />
        </Suspense>
        <div className={styles.content}>
        <Routes>
          <Route
            path="/"
            element={
              carwashStatusLoading ? (
                <p>Загрузка…</p>
              ) : carwashStatus?.is_closed ? (
                <div style={{ padding: 40, textAlign: 'center' }}>Мойка временно закрыта. Справки: 287-03-78</div>
              ) : loading ? (
                <p>Загрузка…</p>
              ) : error ? (
                <p style={{ color: 'red' }}>{error}</p>
              ) : washInfo ? (
                <Suspense fallback={<div>Загрузка…</div>}>
                  <WashInfo
                    washInfo={washInfo}
                    theme="light"
                    onCreateSession={handleCreateSessionWithPayment}
                    onViewHistory={() => navigate(`${BASE}/history`)}
                    onLinkTelegram={handleLinkTelegram}
                    onCancelSession={handleCancelSession}
                    onCompleteSession={(s, p) => setWashInfo((prev) => ({ ...prev, userSession: s, payment: p }))}
                    onStartSession={(s, p) => setWashInfo((prev) => ({ ...prev, userSession: s, payment: p }))}
                    onChemistryEnabled={(s, p) => setWashInfo((prev) => ({ ...prev, userSession: s, payment: p }))}
                    user={user}
                    basePath={BASE}
                  />
                </Suspense>
              ) : (
                <p>Нет данных</p>
              )
            }
          />
          <Route path="/session/:sessionId" element={<Suspense fallback={<div>Загрузка…</div>}><SessionDetails theme="light" user={user} basePath={BASE} /></Suspense>} />
          <Route path="/history" element={<Suspense fallback={<div>Загрузка…</div>}><SessionHistory theme="light" user={user} basePath={BASE} /></Suspense>} />
          <Route
            path="/payment"
            element={
              <Suspense fallback={<div>Загрузка…</div>}>
                <PaymentPage
                  session={location?.state?.session}
                  payment={location?.state?.payment}
                  onPaymentComplete={handlePaymentComplete}
                  onPaymentFailed={handlePaymentFailed}
                  onBack={() => navigate(BASE)}
                  theme="light"
                  paymentType={location?.state?.paymentType || 'main'}
                  basePath={BASE}
                />
              </Suspense>
            }
          />
          <Route path="/booking" element={<Suspense fallback={<div>Загрузка…</div>}><BookingPage theme="light" user={user} onCreateSession={handleCreateSessionWithPayment} /></Suspense>} />
          <Route path="/payment/success" element={<PaymentResultPage success />} />
          <Route path="/payment/fail" element={<PaymentResultPage success={false} />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
        </div>
      </div>
    </div>
  );
};

export default WebApp;
