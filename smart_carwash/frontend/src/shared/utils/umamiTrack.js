/**
 * Umami: подтверждённая самооплата (веб / телега / гость). Кассир не трекается.
 * Дедуп по payment.id (или session.id) в sessionStorage — одна оплата = одно событие.
 *
 * ВАЖНО: трекинг всегда откладывается на следующий тик (setTimeout 0) и обёрнут в
 * try/catch. Аналитика не должна находиться на критическом пути «оплата → назначение
 * бокса»: onPaymentComplete/навигация всегда выполняются раньше, а событие Umami
 * уходит уже после, как fire-and-forget. Так трекинг физически не может задержать
 * или сорвать активацию бокса, даже если скрипт Umami тормозит или падает.
 */

export function trackSelfServicePayment({
  channel,
  payment,
  session,
  source,
  paymentType = 'main',
}) {
  // Снимаем нужные значения синхронно, а сам вызов Umami откладываем.
  const amountRub = payment?.amount ? Number(payment.amount) / 100 : undefined;
  const status = payment?.status;
  const paymentId = payment?.id;
  const sessionId = session?.id;
  const serviceType = session?.service_type || 'unknown';

  const run = () => {
    try {
      if (typeof window === 'undefined' || !window.umami?.track) return;

      // Если есть объект платежа — шлём только при succeeded.
      if (status && status !== 'succeeded') return;

      const dedupeKey = paymentId
        ? `umami_payment_${paymentId}`
        : sessionId
          ? `umami_session_${sessionId}_${paymentType}`
          : null;
      if (dedupeKey && sessionStorage.getItem(dedupeKey)) return;
      if (dedupeKey) sessionStorage.setItem(dedupeKey, '1');

      window.umami.track('payment_confirmed', {
        channel,
        source,
        service: serviceType,
        type: paymentType,
        // Umami Revenue tab требует именно revenue + currency (ISO 4217), не amount.
        ...(amountRub != null && amountRub > 0
          ? { revenue: amountRub, currency: 'RUB' }
          : {}),
      });
    } catch {
      // аналитика не должна ломать оплату
    }
  };

  try {
    setTimeout(run, 0);
  } catch {
    // если даже setTimeout недоступен — просто молча пропускаем трекинг
  }
}
