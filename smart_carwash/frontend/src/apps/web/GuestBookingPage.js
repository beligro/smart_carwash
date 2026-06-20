import React from 'react';
import { lazy, Suspense } from 'react';

const ServiceSelector = lazy(() => import('./components/ServiceSelector'));

/**
 * GuestBookingPage — выбор услуги для гостевой мойки без регистрации.
 * Переиспользует существующий ServiceSelector без изменений.
 */
const GuestBookingPage = ({ onCreateSession }) => {
  return (
    <div>
      <Suspense fallback={<div style={{ padding: 24, textAlign: 'center' }}>Загрузка…</div>}>
        <ServiceSelector
          onSelect={onCreateSession}
          theme="light"
          user={null}
        />
      </Suspense>
    </div>
  );
};

export default GuestBookingPage;
