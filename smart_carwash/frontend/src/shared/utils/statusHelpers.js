/**
 * Получает текстовое представление статуса сессии
 * @param {string} status - Статус сессии
 * @returns {string} - Текстовое представление статуса
 */
export const getSessionStatusText = (status) => {
    switch (status) {
      case 'created':
        return 'Создана';
      case 'in_queue':
        return 'В очереди';
      case 'payment_failed':
        return 'Ошибка оплаты';
      case 'assigned':
        return 'Назначена';
      case 'active':
        return 'Активна';
      case 'complete':
        return 'Завершена';
      case 'canceled':
        return 'Отменена';
      case 'expired':
        return 'Истекла';
      default:
        return 'Неизвестно';
    }
};

/**
 * Получает текстовое представление статуса платежа
 * @param {string} status - Статус платежа
 * @returns {string} - Текстовое представление статуса
 */
export const getPaymentStatusText = (status) => {
    switch (status) {
        case 'pending':
            return 'Ожидает оплаты';
        case 'succeeded':
            return 'Оплачен';
        case 'failed':
            return 'Ошибка оплаты';
        case 'refunded':
            return 'Возвращен';
        default:
            return 'Неизвестно';
    }
};

/**
 * Получает отображаемый статус платежа на основе статуса сессии
 * @param {Object} session - Объект сессии
 * @returns {string} - Отображаемый статус платежа
 */
export const getDisplayPaymentStatus = (session) => {
    if (!session || !session.status) {
        return 'Неизвестно';
    }

    switch (session.status) {
        case 'created':
            return 'Ожидает оплаты';
        case 'in_queue':
        case 'assigned':
        case 'active':
        case 'complete':
            return 'Оплачен';
        case 'canceled':
        case 'expired':
            return 'Возвращен';
        case 'payment_failed':
            return 'Ошибка оплаты';
        default:
            return 'Неизвестно';
    }
};

/**
 * Получает цвет для статуса платежа
 * @param {string} status - Статус платежа
 * @returns {string} - CSS класс цвета
 */
export const getPaymentStatusColor = (status) => {
    switch (status) {
        case 'pending':
            return 'warning';
        case 'succeeded':
            return 'success';
        case 'failed':
            return 'danger';
        case 'refunded':
            return 'info';
        default:
            return 'secondary';
    }
};

/**
 * Рассчитывает итоговую сумму с учетом возвратов
 * @param {number} paidAmount - Оплаченная сумма в копейках
 * @param {number} refundedAmount - Возвращенная сумма в копейках
 * @returns {number} - Итоговая сумма в копейках
 */
export const calculateFinalAmount = (paidAmount, refundedAmount = 0) => {
    return Math.max(0, paidAmount - refundedAmount);
};

/**
 * Форматирует информацию о возврате
 * @param {Object} payment - Объект платежа
 * @returns {Object} - Информация о возврате
 */
export const formatRefundInfo = (payment) => {
    if (!payment) return { hasRefund: false };
    
    const refundedAmount = payment.refunded_amount || 0;
    
    if (refundedAmount > 0) {
        const isFullyRefunded = refundedAmount >= payment.amount;
        const refundType = isFullyRefunded ? 'full' : 'partial';
        
        return {
            hasRefund: true,
            refundedAmount: refundedAmount,
            finalAmount: calculateFinalAmount(payment.amount, refundedAmount),
            refundedAt: payment.refunded_at,
            isFullyRefunded: isFullyRefunded,
            refundType: refundType,
            remainingAmount: payment.amount - refundedAmount
        };
    }
    
    return { hasRefund: false };
};

/**
 * Форматирует информацию о возврате для всех платежей сессии
 * @param {Object} sessionPayments - Объект с платежами сессии
 * @returns {Object} - Объект с информацией о возврате
 */
export const formatSessionRefundInfo = (sessionPayments) => {
    if (!sessionPayments) return { hasRefund: false };
    
    const cost = calculateSessionTotalCost(sessionPayments);
    
    if (cost.totalRefunded > 0) {
        const isFullyRefunded = cost.totalRefunded >= cost.totalAmount;
        const refundType = isFullyRefunded ? 'full' : 'partial';
        
        return {
            hasRefund: true,
            refundedAmount: cost.totalRefunded,
            finalAmount: cost.finalAmount,
            refundedAt: null, // Для сессии с несколькими платежами время возврата может быть разным
            isFullyRefunded: isFullyRefunded,
            refundType: refundType,
            remainingAmount: cost.totalAmount - cost.totalRefunded
        };
    }
    
    return { hasRefund: false };
};

/**
 * Форматирует сумму в рублях
 * @param {number} amount - Сумма в копейках
 * @returns {string} - Отформатированная сумма в рублях
 */
export const formatAmount = (amount) => {
    if (!amount && amount !== 0) return '0 ₽';
    return `${(amount / 100).toFixed(2)} ₽`;
};

/**
 * Форматирует сумму с учетом возврата для отображения
 * @param {Object} payment - Объект платежа
 * @returns {string} - Отформатированная сумма с учетом возврата
 */
export const formatAmountWithRefund = (payment) => {
    if (!payment) return '0 ₽';
    
    const refundedAmount = payment.refunded_amount || 0;
    const originalAmount = payment.amount || 0;
    
    if (refundedAmount === 0) {
        return formatAmount(originalAmount);
    }
    
    const finalAmount = originalAmount - refundedAmount;
    
    if (refundedAmount >= originalAmount) {
        // Полный возврат
        return `${formatAmount(originalAmount)} (возвращено полностью)`;
    } else {
        // Частичный возврат
        return `${formatAmount(finalAmount)} (из ${formatAmount(originalAmount)}, возвращено ${formatAmount(refundedAmount)})`;
    }
};

/**
 * Рассчитывает общую стоимость сессии (основной платеж + все продления)
 * @param {Object} sessionPayments - Объект с платежами сессии
 * @returns {Object} - Объект с общей стоимостью и детализацией
 */
export const calculateSessionTotalCost = (sessionPayments) => {
    if (!sessionPayments) return { totalAmount: 0, totalRefunded: 0, finalAmount: 0, breakdown: [] };
    
    const mainPayment = sessionPayments.main_payment;
    const extensionPayments = sessionPayments.extension_payments || [];
    
    let totalAmount = 0;
    let totalRefunded = 0;
    const breakdown = [];
    
    // Добавляем основной платеж
    if (mainPayment) {
        totalAmount += mainPayment.amount || 0;
        totalRefunded += mainPayment.refunded_amount || 0;
        breakdown.push({
            type: 'main',
            description: 'Основная оплата',
            amount: mainPayment.amount || 0,
            refunded: mainPayment.refunded_amount || 0
        });
    }
    
    // Добавляем платежи продления (только успешные)
    extensionPayments.forEach((payment, index) => {
        // Не учитываем неуспешные платежи продления
        if (payment.status === 'succeeded') {
            totalAmount += payment.amount || 0;
            totalRefunded += payment.refunded_amount || 0;
            breakdown.push({
                type: 'extension',
                description: `Продление ${index + 1}`,
                amount: payment.amount || 0,
                refunded: payment.refunded_amount || 0
            });
        }
    });
    
    const finalAmount = totalAmount - totalRefunded;
    
    return {
        totalAmount,
        totalRefunded,
        finalAmount,
        breakdown
    };
};

/**
 * Форматирует общую стоимость сессии для краткого отображения
 * @param {Object} sessionPayments - Объект с платежами сессии
 * @returns {string} - Отформатированная общая стоимость
 */
export const formatSessionTotalCost = (sessionPayments) => {
    const cost = calculateSessionTotalCost(sessionPayments);
    return formatAmount(cost.finalAmount);
};

/**
 * Форматирует детализированную стоимость сессии для подробного отображения
 * @param {Object} sessionPayments - Объект с платежами сессии
 * @returns {Object} - Объект с детализированной стоимостью для отображения в таблице
 */
export const formatSessionDetailedCost = (sessionPayments) => {
    const cost = calculateSessionTotalCost(sessionPayments);
    
    if (cost.breakdown.length === 0) {
        return {
            totalCost: '0 ₽',
            details: []
        };
    }
    
    if (cost.breakdown.length === 1) {
        // Только основной платеж
        const payment = cost.breakdown[0];
        if (payment.refunded === 0) {
            return {
                totalCost: formatAmount(payment.amount),
                details: []
            };
        } else {
            return {
                totalCost: `${formatAmount(payment.amount - payment.refunded)} (из ${formatAmount(payment.amount)}, возвращено ${formatAmount(payment.refunded)})`,
                details: []
            };
        }
    }
    
    // Несколько платежей - создаем структуру для отображения в таблице
    const details = [];
    
    // Основной платеж
    const mainPayment = cost.breakdown.find(p => p.type === 'main');
    if (mainPayment) {
        details.push({
            label: 'Основная оплата',
            value: formatAmount(mainPayment.amount),
            refunded: mainPayment.refunded > 0 ? formatAmount(mainPayment.refunded) : null
        });
    }
    
    // Продления (суммируем все)
    const extensionPayments = cost.breakdown.filter(p => p.type === 'extension');
    if (extensionPayments.length > 0) {
        const totalExtensionAmount = extensionPayments.reduce((sum, p) => sum + p.amount, 0);
        const totalExtensionRefunded = extensionPayments.reduce((sum, p) => sum + p.refunded, 0);
        
        details.push({
            label: 'Продление',
            value: formatAmount(totalExtensionAmount),
            refunded: totalExtensionRefunded > 0 ? formatAmount(totalExtensionRefunded) : null
        });
    }
    
    // Общая стоимость
    let totalCost = formatAmount(cost.totalAmount);
    if (cost.totalRefunded > 0) {
        totalCost = `${formatAmount(cost.finalAmount)} (из ${formatAmount(cost.totalAmount)}, возвращено ${formatAmount(cost.totalRefunded)})`;
    }
    
    return {
        totalCost,
        details
    };
};
  
  /**
   * Получает текстовое представление статуса бокса
   * @param {string} status - Статус бокса
   * @returns {string} - Текстовое представление статуса
   */
  export const getBoxStatusText = (status) => {
    switch (status) {
      case 'free':
        return 'Свободен';
      case 'reserved':
        return 'Зарезервирован';
      case 'busy':
        return 'Занят';
      case 'maintenance':
        return 'На обслуживании';
      default:
        return 'Неизвестно';
    }
  };
  
  /**
 * Получает описание статуса сессии для отображения пользователю
 * @param {string} status - Статус сессии
 * @returns {string} - Описание статуса
 */
export const getSessionStatusDescription = (status) => {
    switch (status) {
      case 'created':
        return 'Сессия создана, ожидание оплаты...';
      case 'in_queue':
        return 'Оплачено, ожидание назначения бокса...';
      case 'payment_failed':
        return 'Ошибка оплаты, попробуйте еще раз';
      case 'assigned':
        return 'Бокс назначен, ожидание клиента';
      case 'active':
        return 'Мойка в процессе';
      case 'complete':
        return 'Мойка завершена';
      case 'canceled':
        return 'Сессия отменена';
      case 'expired':
        return 'Время сессии истекло';
      default:
        return 'Неизвестный статус';
    }
  };
  
  /**
   * Получает описание типа услуги для отображения пользователю
   * @param {string} serviceType - Тип услуги
   * @returns {string} - Описание типа услуги
   */
  export const getServiceTypeDescription = (serviceType) => {
  // Добавляем логирование для отладки
  if (!serviceType) {
    console.warn('getServiceTypeDescription: serviceType is null or undefined');
    return 'Тип услуги не указан';
  }
  
  console.log('getServiceTypeDescription: serviceType =', serviceType);
  
  switch (serviceType) {
    case 'wash':
      return 'Мойка';
    case 'air_dry':
      return 'Воздух для продувки';
    case 'vacuum':
      return 'Пылеводосос';
    default:
      console.warn('getServiceTypeDescription: unknown serviceType =', serviceType);
      // Возвращаем исходное значение вместо "Неизвестная услуга"
      return serviceType;
  }
};
