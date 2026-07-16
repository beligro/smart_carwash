import React, { useEffect, useState, useRef, Suspense, lazy } from 'react';
import { Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom';
import GuestApiService, { getGuestToken, setGuestToken, clearGuestToken } from './GuestApiService';
import { ReservationWindow } from '../../shared/components/UI';
import styles from './WebApp.module.css';

const Header = lazy(() => import('./components/Header'));
const WashInfo = lazy(() => import('./components/WashInfo/WashInfo'));
const BookingPage = lazy(() => import('./components/BookingPage'));
const GuestPaymentPage = lazy(() => import('./GuestPaymentPage'));

const BASE = '/web/guest';

// Экран после завершения мойки: предложение зарегистрироваться + преимущества
const GuestCompletionOffer = ({ onFinish, session, onPay }) => (
  <div style={{ padding: 16, maxWidth: 480, margin: '0 auto' }}>
    <div style={{ background: '#f0fdf4', borderRadius: 12, padding: '20px', textAlign: 'center' }}>
      <p style={{ fontSize: 20, fontWeight: 700, margin: '0 0 8px', color: '#2e7d32' }}>
        Мойка завершена!
      </p>
      <p style={{ fontSize: 14, color: '#555', margin: '0 0 4px' }}>
        Спасибо, что воспользовались нашей мойкой.
      </p>
    </div>

    {/* Окно приоритетной брони бокса после завершения мойки */}
    {session && (session.cooldown_until || session.cooldown_minutes) && (
      <ReservationWindow
        boxNumber={session.box_number}
        cooldownUntil={session.cooldown_until}
        cooldownMinutes={session.cooldown_minutes}
        completedAt={session.active_ended_at || session.status_updated_at}
        serviceType={session.service_type}
        onPay={onPay}
        theme="light"
      />
    )}

    <div style={{ background: '#f8f9fa', borderRadius: 12, padding: '16px 20px', marginTop: 12 }}>
      <p style={{ fontSize: 15, fontWeight: 600, margin: '0 0 10px' }}>
        Зарегистрируйтесь и получите больше:
      </p>
      <ul style={{ margin: 0, paddingLeft: 18, fontSize: 14, color: '#444', lineHeight: 1.7 }}>
        <li>История моек и платежей</li>
        <li>Уведомления о статусе и завершении</li>
        <li>Программа лояльности — каждая N-я мойка в подарок</li>
        <li>Быстрый повторный заказ без ввода данных</li>
      </ul>
    </div>

    <a
      href="/web/login"
      style={{ display: 'block', textAlign: 'center', textDecoration: 'none', background: '#1a73e8', color: '#fff', borderRadius: 10, padding: '14px 0', fontSize: 15, fontWeight: 600, marginTop: 16 }}
    >
      Зарегистрироваться
    </a>
    <button
      onClick={onFinish}
      style={{ width: '100%', background: 'transparent', color: '#888', border: 'none', padding: '14px 0', fontSize: 14, cursor: 'pointer', marginTop: 4 }}
    >
      Помыть ещё раз без регистрации
    </button>
  </div>
);

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
    // Для завершённой мойки не сбрасываем автоматически — показываем экран
    // с предложением регистрации; пользователь сам закрывает его.
    if (session.status === 'complete') return;
    if (sessionResetTimer.current) return;
    sessionResetTimer.current = setTimeout(() => {
      clearGuestToken();
      setWashInfo((prev) => (prev ? { ...prev, userSession: null, payment: null } : prev));
      sessionResetTimer.current = null;
    }, 5000);
  };

  // Завершение гостем экрана после мойки: чистим cookie и возвращаем к форме
  const handleGuestFinish = () => {
    if (sessionResetTimer.current) {
      clearTimeout(sessionResetTimer.current);
      sessionResetTimer.current = null;
    }
    clearGuestToken();
    setWashInfo((prev) => (prev ? { ...prev, userSession: null, payment: null } : prev));
    navigate(BASE, { replace: true });
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
      // Восстановление сессии после возврата из банка: если возврат открылся
      // в другом браузере (нет cookie), токен приходит в URL (?gt=...).
      try {
        const gt = new URLSearchParams(window.location.search).get('gt');
        if (gt) setGuestToken(gt);
      } catch {}

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

            // Резюме оплаты: если у клиента в этом же браузере осталась незавершённая
            // неоплаченная сессия (created) с живой ссылкой Tinkoff — ведём сразу на оплату,
            // чтобы он продолжил, а не упирался в «уже есть активная сессия» при пересоздании.
            // Не вмешиваемся в возврат из банка (?return=...) и когда уже на странице оплаты.
            const params = new URLSearchParams(window.location.search);
            const isReturnFlow = !!params.get('return');
            const onPaymentPage = window.location.pathname.includes('/payment');
            if (
              data.session.status === 'created' &&
              data.payment?.payment_url &&
              (!data.payment.status || data.payment.status === 'pending') &&
              !isReturnFlow &&
              !onPaymentPage
            ) {
              navigate(`${BASE}/payment`, { state: { session: data.session, payment: data.payment } });
            }
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

  // Переход в обычный флоу создания новой гостевой сессии/оплаты
  // (в течение кулдауна тот же бокс достанется по приоритету по номеру машины)
  const handlePayToStay = () => {
    if (sessionResetTimer.current) {
      clearTimeout(sessionResetTimer.current);
      sessionResetTimer.current = null;
    }
    const preselectServiceType = washInfo?.userSession?.service_type;
    navigate(`${BASE}/booking`, { state: { preselectServiceType } });
  };

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
                  ) : washInfo?.userSession?.status === 'complete' ? (
                    <GuestCompletionOffer
                      onFinish={handleGuestFinish}
                      session={washInfo.userSession}
                      onPay={handlePayToStay}
                    />
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
                path="/booking"
                element={
                  <BookingPage
                    theme="light"
                    user={null}
                    onCreateSession={handleCreateSession}
                  />
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
              <Route path="*" element={<Navigate to={BASE} replace />} />
            </Routes>
          </Suspense>
        </div>
      </div>
    </div>
  );
};

export default GuestApp;
