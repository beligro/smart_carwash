import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';

const BOX_TYPE_LABEL = { wash: 'Мойка', vacuum: 'Пылесос', air: 'Воздух' };
const TYPE_ORDER = ['wash', 'vacuum', 'air'];

const Container = styled.div`
  background: ${p => p.theme.cardBackground};
  border: 1px solid ${p => p.theme.borderColor};
  border-radius: 8px;
  padding: 16px;
`;

const HeaderRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 12px;
`;

const Title = styled.h2`
  margin: 0;
`;

const Button = styled.button`
  padding: 8px 14px;
  background: ${p => (p.primary ? p.theme.primaryColor : p.theme.cardBackground)};
  color: ${p => (p.primary ? 'white' : p.theme.textColor)};
  border: 1px solid ${p => (p.primary ? p.theme.primaryColor : p.theme.borderColor)};
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.95rem;
  &:disabled { opacity: 0.6; cursor: not-allowed; }
`;

const ErrorText = styled.div`
  color: #d33;
  margin-bottom: 8px;
`;

const SuccessText = styled.div`
  color: #2e7d32;
  margin-bottom: 8px;
`;

const SectionTitle = styled.h3`
  margin: 20px 0 10px;
`;

const Grid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
`;

const Card = styled.div`
  border: 1px solid ${p => p.theme.borderColor};
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: ${p => (p.active ? 'rgba(22, 163, 74, 0.08)' : p.theme.backgroundColor)};
`;

const BoxNumber = styled.div`
  font-size: 1.2rem;
  font-weight: 700;
`;

const Muted = styled.div`
  font-size: 0.85rem;
  opacity: 0.75;
`;

const Timer = styled.div`
  font-size: 1.4rem;
  font-weight: 800;
  color: ${p => (p.warn ? '#dc2626' : p.theme.primaryColor)};
  font-variant-numeric: tabular-nums;
`;

const StartButton = styled(Button)`
  width: 100%;
`;

const StopButton = styled(Button)`
  width: 100%;
  background: #dc2626;
  color: #fff;
  border-color: #dc2626;
`;

const pad2 = (n) => String(n).padStart(2, '0');

// Обратный отсчёт до expiresAt в формате MM:SS (не уходит в минус).
const countdown = (expiresIso, nowMs) => {
  if (!expiresIso) return '—';
  const end = new Date(expiresIso).getTime();
  if (Number.isNaN(end)) return '—';
  let s = Math.max(0, Math.floor((end - nowMs) / 1000));
  const m = Math.floor(s / 60);
  s -= m * 60;
  return `${pad2(m)}:${pad2(s)}`;
};

const MyWash = () => {
  const theme = getTheme('light');
  const [boxes, setBoxes] = useState([]);
  const [active, setActive] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [feedback, setFeedback] = useState('');
  const [busyId, setBusyId] = useState(null);
  const [nowMs, setNowMs] = useState(Date.now());
  const isMounted = useRef(true);

  const load = useCallback(async (silent = false) => {
    if (!silent) setLoading(true);
    try {
      const data = await ApiService.getPersonalWashBoxes();
      if (!isMounted.current) return;
      setBoxes(Array.isArray(data?.boxes) ? data.boxes : []);
      setActive(Array.isArray(data?.active) ? data.active : []);
      setError('');
    } catch (e) {
      if (!isMounted.current) return;
      setError('Не удалось загрузить боксы. Попробуйте обновить.');
    } finally {
      if (isMounted.current && !silent) setLoading(false);
    }
  }, []);

  useEffect(() => {
    isMounted.current = true;
    load();
    const poll = setInterval(() => load(true), 7000);
    return () => {
      isMounted.current = false;
      clearInterval(poll);
    };
  }, [load]);

  // Локальный тик раз в секунду для живых таймеров.
  useEffect(() => {
    const id = setInterval(() => setNowMs(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  const start = async (boxId) => {
    setBusyId(boxId);
    setFeedback('');
    try {
      const resp = await ApiService.startPersonalWash(boxId);
      setFeedback(resp?.message || 'Бокс включён');
      await load(true);
    } catch (e) {
      setError(e?.response?.data?.error || 'Не удалось включить бокс');
    } finally {
      setBusyId(null);
    }
  };

  const stop = async (boxId) => {
    setBusyId(boxId);
    setFeedback('');
    try {
      await ApiService.returnPersonalWash(boxId);
      setFeedback('Бокс возвращён в работу');
      await load(true);
    } catch (e) {
      setError(e?.response?.data?.error || 'Не удалось вернуть бокс');
    } finally {
      setBusyId(null);
    }
  };

  // Свободные боксы, сгруппированные по типу.
  const freeByType = useMemo(() => {
    const groups = {};
    for (const b of boxes) {
      if (b.status !== 'free') continue;
      const t = b.box_type || 'wash';
      if (!groups[t]) groups[t] = [];
      groups[t].push(b);
    }
    for (const t of Object.keys(groups)) {
      groups[t].sort((a, b) => (a.number ?? 0) - (b.number ?? 0));
    }
    return groups;
  }, [boxes]);

  const sortedActive = useMemo(
    () => active.slice().sort((a, b) => (a.number ?? 0) - (b.number ?? 0)),
    [active],
  );

  return (
    <Container theme={theme}>
      <HeaderRow>
        <Title>Моя мойка</Title>
        <Button theme={theme} onClick={() => load()} disabled={loading}>
          {loading ? 'Обновление…' : 'Обновить'}
        </Button>
      </HeaderRow>

      {error && <ErrorText onClick={() => setError('')}>{error}</ErrorText>}
      {feedback && <SuccessText onClick={() => setFeedback('')}>{feedback}</SuccessText>}

      {sortedActive.length > 0 && (
        <>
          <SectionTitle>Активные личные включения</SectionTitle>
          <Grid>
            {sortedActive.map((a) => {
              const secLeft = a.expires_at
                ? Math.max(0, Math.floor((new Date(a.expires_at).getTime() - nowMs) / 1000))
                : 0;
              return (
                <Card key={a.action_id || a.id || a.number} theme={theme} active>
                  <BoxNumber>Бокс №{a.number}</BoxNumber>
                  <Muted>{BOX_TYPE_LABEL[a.box_type] || a.box_type || '—'}</Muted>
                  {a.admin_username && <Muted>Кто: {a.admin_username}</Muted>}
                  <Timer theme={theme} warn={secLeft <= 60}>{countdown(a.expires_at, nowMs)}</Timer>
                  <StopButton
                    theme={theme}
                    onClick={() => stop(a.id)}
                    disabled={!a.id || busyId === a.id}
                  >
                    {busyId === a.id ? 'Выключение…' : 'Выключить сейчас'}
                  </StopButton>
                </Card>
              );
            })}
          </Grid>
        </>
      )}

      {TYPE_ORDER.filter((t) => (freeByType[t] || []).length > 0).map((t) => (
        <React.Fragment key={t}>
          <SectionTitle>{BOX_TYPE_LABEL[t]}</SectionTitle>
          <Grid>
            {freeByType[t].map((b) => {
              const remaining = b.remaining ?? 0;
              const canStart = remaining > 0 && busyId !== b.id;
              return (
                <Card key={b.id} theme={theme}>
                  <BoxNumber>Бокс №{b.number}</BoxNumber>
                  <Muted>Остаток лимита: {remaining} / 2 в сутки</Muted>
                  <StartButton
                    primary
                    theme={theme}
                    onClick={() => start(b.id)}
                    disabled={!canStart}
                  >
                    {busyId === b.id
                      ? 'Включение…'
                      : remaining > 0
                      ? 'Личное включение (15 мин)'
                      : 'Лимит исчерпан'}
                  </StartButton>
                </Card>
              );
            })}
          </Grid>
        </React.Fragment>
      ))}

      {!loading && sortedActive.length === 0 && TYPE_ORDER.every((t) => (freeByType[t] || []).length === 0) && (
        <Muted style={{ marginTop: 16 }}>Свободных боксов нет.</Muted>
      )}
    </Container>
  );
};

export default MyWash;
