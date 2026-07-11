import React, { useEffect, useMemo, useState } from 'react';
import styled, { keyframes } from 'styled-components';

/**
 * ReservationWindow — окно/панель приоритетной брони бокса после завершения мойки.
 *
 * После перехода сессии в статус `complete` бокс на время кулдауна держится за
 * клиентом (приоритетный возврат в тот же бокс). Компонент показывает обратный
 * таймер до конца кулдауна и кнопку «Оплатить и остаться», ведущую в обычный
 * флоу создания новой сессии/оплаты.
 *
 * @param {Object} props
 * @param {number|string} [props.boxNumber] - Номер бокса.
 * @param {string} [props.cooldownUntil] - ISO-дата дедлайна кулдауна (приоритетно).
 * @param {number} [props.cooldownMinutes] - Длительность кулдауна в минутах (fallback).
 * @param {string} [props.completedAt] - Момент завершения мойки (для fallback-расчёта дедлайна).
 * @param {string} [props.serviceType] - Тип услуги завершённой сессии (для предвыбора в новом флоу).
 * @param {Function} props.onPay - Колбэк перехода в флоу создания новой сессии/оплаты.
 * @param {string} [props.theme] - Тема оформления ('light' | 'dark').
 */
const ReservationWindow = ({
  boxNumber,
  cooldownUntil,
  cooldownMinutes,
  completedAt,
  serviceType,
  onPay,
  theme = 'light',
}) => {
  // Вычисляем дедлайн брони: приоритетно из cooldown_until, иначе из completedAt + cooldownMinutes.
  const deadline = useMemo(() => {
    if (cooldownUntil) {
      const d = new Date(cooldownUntil);
      if (!isNaN(d.getTime())) {
        return d.getTime();
      }
    }
    if (cooldownMinutes && completedAt) {
      const base = new Date(completedAt);
      if (!isNaN(base.getTime())) {
        return base.getTime() + cooldownMinutes * 60 * 1000;
      }
    }
    return null;
  }, [cooldownUntil, cooldownMinutes, completedAt]);

  const [secondsLeft, setSecondsLeft] = useState(() => {
    if (deadline == null) return null;
    return Math.max(0, Math.floor((deadline - Date.now()) / 1000));
  });

  useEffect(() => {
    if (deadline == null) {
      setSecondsLeft(null);
      return undefined;
    }

    const tick = () => {
      const remaining = Math.max(0, Math.floor((deadline - Date.now()) / 1000));
      setSecondsLeft(remaining);
      return remaining;
    };

    // Считаем сразу, затем каждую секунду
    if (tick() <= 0) {
      return undefined;
    }
    const intervalId = setInterval(() => {
      if (tick() <= 0) {
        clearInterval(intervalId);
      }
    }, 1000);

    return () => clearInterval(intervalId);
  }, [deadline]);

  // Если дедлайна нет вовсе — ничего не показываем.
  if (deadline == null) {
    return null;
  }

  const expired = secondsLeft == null || secondsLeft <= 0;
  const minutes = expired ? 0 : Math.floor(secondsLeft / 60);
  const seconds = expired ? 0 : secondsLeft % 60;
  const formatted = `${minutes}:${seconds.toString().padStart(2, '0')}`;

  const boxLabel = boxNumber != null && boxNumber !== '' ? `№${boxNumber}` : '';

  return (
    <Panel $theme={theme} $expired={expired}>
      <IconRow>
        <Icon>{expired ? '⌛' : '🅿️'}</Icon>
      </IconRow>

      {expired ? (
        <>
          <Title $theme={theme}>Бронь истекла</Title>
          <Text $theme={theme}>
            Время приоритетного удержания бокса {boxLabel} закончилось. Вы можете
            записаться на новую мойку в общем порядке.
          </Text>
        </>
      ) : (
        <>
          <Title $theme={theme}>
            Бокс {boxLabel} забронирован за вами ещё
          </Title>
          <Countdown $theme={theme}>{formatted}</Countdown>
          <Text $theme={theme}>Оплатите, чтобы остаться в боксе.</Text>
          {typeof onPay === 'function' && (
            <PayButton
              type="button"
              onClick={() => onPay({ serviceType })}
            >
              Оплатить и остаться
            </PayButton>
          )}
        </>
      )}
    </Panel>
  );
};

const pulse = keyframes`
  0% { box-shadow: 0 0 0 0 rgba(255, 152, 0, 0.35); }
  70% { box-shadow: 0 0 0 10px rgba(255, 152, 0, 0); }
  100% { box-shadow: 0 0 0 0 rgba(255, 152, 0, 0); }
`;

const Panel = styled.div`
  margin: 16px 0;
  padding: 20px 18px;
  border-radius: 14px;
  text-align: center;
  border: 2px solid ${({ $expired }) => ($expired ? '#bdbdbd' : '#ff9800')};
  background: ${({ $theme, $expired }) => {
    if ($expired) {
      return $theme === 'dark' ? '#2a2a2a' : '#f5f5f5';
    }
    return $theme === 'dark' ? '#3a2e10' : '#fff8e1';
  }};
  animation: ${({ $expired }) => ($expired ? 'none' : pulse)} 2s infinite;
`;

const IconRow = styled.div`
  display: flex;
  justify-content: center;
`;

const Icon = styled.div`
  font-size: 32px;
  line-height: 1;
  margin-bottom: 8px;
`;

const Title = styled.p`
  margin: 0 0 6px 0;
  font-size: 16px;
  font-weight: 700;
  color: ${({ $theme }) => ($theme === 'dark' ? '#ffe0b2' : '#e65100')};
`;

const Countdown = styled.div`
  font-size: 40px;
  font-weight: 800;
  letter-spacing: 1px;
  margin: 6px 0;
  font-variant-numeric: tabular-nums;
  color: ${({ $theme }) => ($theme === 'dark' ? '#ffb74d' : '#ef6c00')};
`;

const Text = styled.p`
  margin: 4px 0 0 0;
  font-size: 14px;
  color: ${({ $theme }) => ($theme === 'dark' ? '#cfcfcf' : '#5d4037')};
`;

const PayButton = styled.button`
  margin-top: 16px;
  width: 100%;
  padding: 14px 16px;
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  background: #ff9800;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.2s ease;

  &:hover {
    background: #fb8c00;
  }

  &:active {
    background: #ef6c00;
  }
`;

export default ReservationWindow;
