import React, { useCallback, useEffect, useMemo, useState } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';

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

const Filters = styled.div`
  display: flex;
  align-items: flex-end;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
`;

const Field = styled.div`
  display: flex;
  flex-direction: column;
  gap: 4px;
`;

const Label = styled.label`
  font-size: 0.85rem;
  color: ${p => p.theme.textColor};
  opacity: 0.8;
`;

const DateInput = styled.input`
  padding: 8px 10px;
  border: 1px solid ${p => p.theme.borderColor};
  border-radius: 6px;
  background: ${p => p.theme.backgroundColor};
  color: ${p => p.theme.textColor};
  font-size: 0.95rem;
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

const SectionTitle = styled.h3`
  margin: 20px 0 10px;
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  font-size: 0.95rem;
`;

const Th = styled.th`
  text-align: left;
  padding: 8px 10px;
  border-bottom: 2px solid ${p => p.theme.borderColor};
  white-space: nowrap;
`;

const Td = styled.td`
  padding: 8px 10px;
  border-bottom: 1px solid ${p => p.theme.borderColor};
`;

const Empty = styled.div`
  padding: 12px 0;
  opacity: 0.7;
`;

// pad2 форматирует число с ведущим нулём.
const pad2 = (n) => String(n).padStart(2, '0');

// toDateInput возвращает дату в формате YYYY-MM-DD.
const toDateInput = (d) => `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`;

// defaultRange — с начала текущего месяца по сегодня.
const defaultRange = () => {
  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth(), 1);
  return { from: toDateInput(from), to: toDateInput(now) };
};

// formatDateTime форматирует ISO-время в читаемый вид.
const formatDateTime = (iso) => {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return `${pad2(d.getDate())}.${pad2(d.getMonth() + 1)}.${d.getFullYear()} ${pad2(d.getHours())}:${pad2(d.getMinutes())}`;
};

const PersonalWashReport = () => {
  const theme = getTheme('light');
  const initial = useMemo(() => defaultRange(), []);
  const [from, setFrom] = useState(initial.from);
  const [to, setTo] = useState(initial.to);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [summary, setSummary] = useState([]);
  const [items, setItems] = useState([]);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await ApiService.getPersonalWashReport({ from, to });
      setSummary(data?.summary || []);
      setItems(data?.items || []);
    } catch (e) {
      setError(e?.response?.data?.error || e?.message || 'Ошибка загрузки отчёта');
    } finally {
      setLoading(false);
    }
  }, [from, to]);

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <Container theme={theme}>
      <HeaderRow>
        <Title>Личные мойки (отчёт)</Title>
      </HeaderRow>

      <Filters>
        <Field>
          <Label theme={theme}>С даты</Label>
          <DateInput
            theme={theme}
            type="date"
            value={from}
            onChange={(e) => setFrom(e.target.value)}
          />
        </Field>
        <Field>
          <Label theme={theme}>По дату</Label>
          <DateInput
            theme={theme}
            type="date"
            value={to}
            onChange={(e) => setTo(e.target.value)}
          />
        </Field>
        <Button theme={theme} primary onClick={load} disabled={loading}>
          {loading ? 'Загрузка…' : 'Обновить'}
        </Button>
      </Filters>

      {error && <ErrorText>{error}</ErrorText>}

      <SectionTitle>Сводка по сотрудникам</SectionTitle>
      {summary.length === 0 ? (
        <Empty>Нет данных за выбранный период</Empty>
      ) : (
        <Table>
          <thead>
            <tr>
              <Th theme={theme}>Сотрудник</Th>
              <Th theme={theme}>Кол-во</Th>
              <Th theme={theme}>Минут</Th>
            </tr>
          </thead>
          <tbody>
            {summary.map((row, idx) => (
              <tr key={`${row.admin_username}-${idx}`}>
                <Td theme={theme}>{row.admin_username || '—'}</Td>
                <Td theme={theme}>{row.count}</Td>
                <Td theme={theme}>{row.minutes}</Td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}

      <SectionTitle>Детальная история</SectionTitle>
      {items.length === 0 ? (
        <Empty>Нет данных за выбранный период</Empty>
      ) : (
        <Table>
          <thead>
            <tr>
              <Th theme={theme}>Дата</Th>
              <Th theme={theme}>Сотрудник</Th>
              <Th theme={theme}>Бокс</Th>
              <Th theme={theme}>Минут</Th>
            </tr>
          </thead>
          <tbody>
            {items.map((item, idx) => (
              <tr key={`${item.started_at}-${idx}`}>
                <Td theme={theme}>{formatDateTime(item.started_at)}</Td>
                <Td theme={theme}>{item.admin_username || '—'}</Td>
                <Td theme={theme}>№{item.box_number} · {item.box_type_label}</Td>
                <Td theme={theme}>{item.minutes}</Td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}
    </Container>
  );
};

export default PersonalWashReport;
