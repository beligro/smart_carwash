/**
 * Umami: подтверждённая самооплата (веб / телега / гость). Кассир не трекается.
 * Дедуп по payment.id (или session.id) в sessionStorage — одна оплата = одно событие.
 */

export function trackSelfServicePayment({
  channel,
  payment,
  session,
  source,
  paymentType = 'main',
}) {
  try {
    if (typeof window === 'undefined' || !window.umami?.track) return;

    // Если есть объект платежа — шлём только при succeeded.
    if (payment?.status && payment.status !== 'succeeded') return;

    const dedupeKey = payment?.id
      ? `umami_payment_${payment.id}`
      : session?.id
        ? `umami_session_${session.id}_${paymentType}`
        : null;
    if (dedupeKey && sessionStorage.getItem(dedupeKey)) return;
    if (dedupeKey) sessionStorage.setItem(dedupeKey, '1');

    const amountRub = payment?.amount ? Math.round(Number(payment.amount) / 100) : undefined;

    window.umami.track('payment_confirmed', {
      channel,
      source,
      service: session?.service_type || 'unknown',
      type: paymentType,
      ...(amountRub != null && amountRub > 0 ? { amount: amountRub } : {}),
    });
  } catch {
    // аналитика не должна ломать оплату
  }
}
