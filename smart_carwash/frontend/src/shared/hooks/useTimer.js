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
    
    console.log('calculateTimeLeft: sessionData =', sessionData);
    
    if (sessionData.status === 'active') {
      // Для активной сессии - используем выбранное время аренды (по умолчанию 5 минут)
      // Время начала сессии - это время последнего обновления статуса на active
      const startTimeStr = sessionData.status_updated_at || sessionData.updated_at;
      console.log('calculateTimeLeft: startTimeStr =', startTimeStr);
      
      if (!startTimeStr) {
        console.warn('Отсутствует время обновления статуса для активной сессии');
        return null;
      }
      
      const startTime = new Date(startTimeStr);
      if (isNaN(startTime.getTime())) {
        console.warn('Некорректное время обновления статуса для активной сессии:', startTimeStr);
        return null;
      }
      
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
      
      console.log('calculateTimeLeft: active session, remainingSeconds =', remainingSeconds);
      return remainingSeconds;
    } else if (sessionData.status === 'assigned') {
      // Для назначенной сессии - 3 минуты с момента назначения
      // Время назначения сессии - это время последнего обновления статуса на assigned
      const assignedTimeStr = sessionData.status_updated_at || sessionData.updated_at;
      console.log('calculateTimeLeft: assignedTimeStr =', assignedTimeStr);
      
      if (!assignedTimeStr) {
        console.warn('Отсутствует время обновления статуса для назначенной сессии');
        return null;
      }
      
      const assignedTime = new Date(assignedTimeStr);
      if (isNaN(assignedTime.getTime())) {
        console.warn('Некорректное время обновления статуса для назначенной сессии:', assignedTimeStr);
        return null;
      }
      
      const now = new Date();
      
      // Общая продолжительность резерва - 3 минуты (180 секунд)
      const totalDuration = 180; // в секундах
      
      // Прошедшее время в секундах
      const elapsedSeconds = differenceInSeconds(now, assignedTime);
      
      // Оставшееся время в секундах
      const remainingSeconds = Math.max(0, totalDuration - elapsedSeconds);
      
      console.log('calculateTimeLeft: assigned session, remainingSeconds =', remainingSeconds);
      return remainingSeconds;
    }
    
    console.log('calculateTimeLeft: unknown status =', sessionData.status);
    return null;
  }, []);
  
  // Обновление таймера при изменении сессии
  useEffect(() => {
    console.log('useTimer useEffect: session =', session);
    
    if (session && (session.status === 'active' || session.status === 'assigned')) {
      console.log('useTimer: session status is active or assigned, calculating time...');
      
      // Рассчитываем оставшееся время
      const remaining = calculateTimeLeft(session);
      console.log('useTimer: calculated remaining time =', remaining);
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
