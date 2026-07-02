import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import GuestApiService, { getGuestToken, setGuestToken } from './GuestApiService';

const SERVICES = { wash: 'Мойка', air_dry: 'Обдув', vacuum: 'Пылесос' };

const formatPrice = (kopecks) =>
  kopecks ? `${Math.round(kopecks / 100)} ₽` : '0 ₽';

/**
 * GuestPaymentPage — страница оплаты для гостевой сессии.
 * Аналог PaymentPage, но использует GuestApiService (без JWT).
 */
const GuestPaymentPage = ({
  session: sessionProp,
  payment: paymentProp,
  paymentType = 'main',
  onPaymentComplete,
  onPaymentFailed,
  onBack,
  basePath,
}) => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [session, setSession] = useState(sessionProp);
  const [payment, setPayment] = useState(paymentProp);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [returnHandled, setReturnHandled] = useState(false);

  // Обработка возврата с Tinkoff (?return=success|fail&gt=<токен>)
  useEffect(() => {
    const returnType = searchParams.get('return');
    if (!returnType || returnHandled) return;

    // Восстанавливаем токен из URL: банк мог открыть возврат в другом
    // браузере, где нет cookie гостевой сессии.
    const gtFromUrl = searchParams.get('gt');
    if (gtFromUrl) setGuestToken(gtFromUrl);

    setReturnHandled(true);
    setSearchParams({}, { replace: true });

    const token = gtFromUrl || getGuestToken();
    if (!token) {
      setError('Сессия не найдена');
      return;
    }

    const fetchAndHandle = async () => {
      setLoading(true);
      try {
        const data = await GuestApiService.getSession(token);
        const sess = data?.session;
        if (!sess) { setError('Сессия не найдена'); return; }
        setSession(sess);
        setPayment(data.payment);
        if (returnType === 'success') {
          onPaymentComplete?.(sess);
        } else {
          onPaymentFailed?.(sess);
        }
      } catch {
        setError('Не удалось загрузить данные сессии');
      } finally {
        setLoading(false);
      }
    };
    fetchAndHandle();
  }, [searchParams, returnHandled, onPaymentComplete, onPaymentFailed, setSearchParams]);

  const handlePay = () => {
    if (payment?.payment_url) {
      window.location.href = payment.payment_url;
    }
  };

  const title = paymentType === 'extension' ? 'Продление сессии' : 'Оплата услуги';
  const serviceName = SERVICES[session?.service_type] || session?.service_type || '';

  if (loading) {
    return <div style={{ padding: 40, textAlign: 'center' }}>Загрузка…</div>;
  }

  if (error) {
    return (
      <div style={{ padding: 24 }}>
        <p style={{ color: 'red' }}>{error}</p>
        <button onClick={onBack} style={btnStyle}>Назад</button>
      </div>
    );
  }

  if (!session || !payment) {
    return (
      <div style={{ padding: 24 }}>
        <p>Нет данных платежа.</p>
        <button onClick={onBack} style={btnStyle}>Назад</button>
      </div>
    );
  }

  return (
    <div style={{ padding: 16, maxWidth: 480, margin: '0 auto' }}>
      <h2 style={{ marginBottom: 16 }}>{title}</h2>

      <div style={cardStyle}>
        <p style={rowStyle}>
          <span>Услуга</span>
          <b>{serviceName}</b>
        </p>
        {session.rental_time_minutes > 0 && (
          <p style={rowStyle}>
            <span>Время</span>
            <b>{session.rental_time_minutes} мин</b>
          </p>
        )}
        {session.car_number && (
          <p style={rowStyle}>
            <span>Автомобиль</span>
            <b>{session.car_number}</b>
          </p>
        )}
        <p style={{ ...rowStyle, fontSize: 18, fontWeight: 700, marginTop: 12 }}>
          <span>Итого</span>
          <span>{formatPrice(payment.amount)}</span>
        </p>
      </div>

      <button onClick={handlePay} style={{ ...btnStyle, background: '#1a73e8', color: '#fff', fontSize: 16, marginTop: 12 }}>
        Перейти к оплате
      </button>
      <button onClick={onBack} style={{ ...btnStyle, marginTop: 8, background: 'transparent', color: '#555' }}>
        Назад
      </button>

      <p style={{ marginTop: 16, fontSize: 12, color: '#888', textAlign: 'center' }}>
        После оплаты вы вернётесь на страницу с таймером вашей мойки.
      </p>
    </div>
  );
};

const cardStyle = {
  background: '#f8f9fa',
  borderRadius: 12,
  padding: '16px 20px',
  marginBottom: 8,
};

const rowStyle = {
  display: 'flex',
  justifyContent: 'space-between',
  margin: '6px 0',
  fontSize: 14,
};

const btnStyle = {
  display: 'block',
  width: '100%',
  padding: '14px 0',
  border: 'none',
  borderRadius: 10,
  cursor: 'pointer',
  fontSize: 15,
};

export default GuestPaymentPage;
