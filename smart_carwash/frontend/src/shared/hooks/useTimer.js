import { useState, useEffect, useCallback } from 'react';
import { differenceInSeconds } from 'date-fns';

/**
 * Хук для работы с таймером обратного отсчета
 * @param {Object} session - Объект сессии
 * @returns {Object} - Объект с данными таймера
 */
const useTimer = (session) => {
  const [timeLeft, setTimeLeft] = useState(null);
  
  // Функция для расчета оставшегося времени
  const calculateTimeLeft = useCallback((sessionData) => {
    if (!sessionData) {
      return null;
    }
    
    if (sessionData.status === 'active') {
      const startTime = new Date(sessionData.status_updated_at || sessionData.updated_at);
      
      const now = new Date();
      
      // Получаем выбранное время аренды в минутах (по умолчанию 5 минут)
      const rentalTimeMinutes = sessionData.rental_time_minutes || 5;
      
      // Учитываем время продления, если оно есть
      const extensionTimeMinutes = sessionData.extension_time_minutes || 0;
      
      // Общая продолжительность сессии в секундах
      const totalDuration = (rentalTimeMinutes + extensionTimeMinutes) * 60; // в секундах
      
      // Прошедшее время в секундах
      const elapsedSeconds = differenceInSeconds(now, startTime);
      
      // Оставшееся время в секундах
      const remainingSeconds = Math.max(0, totalDuration - elapsedSeconds);
      
      return remainingSeconds;
    } else if (sessionData.status === 'assigned') {
      const assignedTime = new Date(sessionData.status_updated_at || sessionData.updated_at);
      
      const now = new Date();
      
      // Общая продолжительность резерва - 3 минуты (180 секунд)
      const totalDuration = 180; // в секундах
      
      // Прошедшее время в секундах
      const elapsedSeconds = differenceInSeconds(now, assignedTime);
      
      // Оставшееся время в секундах
      const remainingSeconds = Math.max(0, totalDuration - elapsedSeconds);
      
      return remainingSeconds;
    }
    
    return null;
  }, []);
  
  // Обновление таймера при изменении сессии
  useEffect(() => {
    if (session && (session.status === 'active' || session.status === 'assigned')) {
      // Рассчитываем оставшееся время
      const remaining = calculateTimeLeft(session);
      setTimeLeft(remaining);
      
      // Обновляем таймер каждую секунду
      const timerId = setInterval(() => {
        const remaining = calculateTimeLeft(session);
        setTimeLeft(remaining);
        
        // Если время вышло, останавливаем таймер
        if (remaining <= 0) {
          clearInterval(timerId);
        }
      }, 1000);
      
      // Очищаем интервал при размонтировании компонента
      return () => clearInterval(timerId);
    } else {
      console.log('useTimer: session is not active or assigned, resetting timer');
      // Если сессия не активна и не назначена, сбрасываем таймер
      setTimeLeft(null);
    }
  }, [session, calculateTimeLeft]);
  
  return {
    timeLeft,
    isWarning: timeLeft !== null && timeLeft <= 120 && timeLeft > 60,
    isDanger: timeLeft !== null && timeLeft <= 60,
    isActive: timeLeft !== null && timeLeft > 0
  };
};

export default useTimer;
