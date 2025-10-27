import { useEffect, useRef, useCallback } from 'react';

/**
 * Хук для автоматического поллинга данных
 * @param {Function} callback - функция для выполнения при каждом поллинге
 * @param {number} interval - интервал в миллисекундах
 * @param {boolean} enabled - включен ли поллинг
 * @param {Array} dependencies - зависимости для перезапуска поллинга
 * @param {boolean} immediate - выполнить callback сразу при монтировании
 */
const usePolling = (callback, interval, enabled = true, dependencies = [], immediate = false) => {
  const intervalRef = useRef(null);
  const callbackRef = useRef(callback);
  const isExecutingRef = useRef(false);

  // Обновляем ссылку на callback
  callbackRef.current = callback;

  // Функция выполнения поллинга
  const executePolling = useCallback(async () => {
    if (isExecutingRef.current) {
      return; // Предотвращаем дублирующие запросы
    }

    isExecutingRef.current = true;
    
    try {
      await callbackRef.current();
    } catch (error) {
      console.error('Ошибка при выполнении поллинга:', error);
      // Продолжаем поллинг даже при ошибках
    } finally {
      isExecutingRef.current = false;
    }
  }, []);

  useEffect(() => {
    // Очищаем предыдущий интервал
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }

    // Если поллинг отключен, не запускаем
    if (!enabled || interval <= 0) {
      return;
    }

    // Выполняем callback сразу, если требуется
    if (immediate) {
      executePolling();
    }

    // Устанавливаем интервал
    intervalRef.current = setInterval(executePolling, interval);

    // Очистка при размонтировании или изменении зависимостей
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [enabled, interval, executePolling, immediate, ...dependencies]);

  // Очистка при размонтировании компонента
  useEffect(() => {
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, []);
};

export default usePolling;

