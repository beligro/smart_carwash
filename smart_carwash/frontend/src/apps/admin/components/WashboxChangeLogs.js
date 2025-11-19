import React, { useEffect, useMemo, useState } from 'react';
import styled from 'styled-components';
import ApiService from '../../../shared/services/ApiService';

const Container = styled.div`
  background: ${p => p.theme.cardBackground};
  border: 1px solid ${p => p.theme.borderColor};
  border-radius: 8px;
  padding: 16px;
`;

const Title = styled.h2`
  margin: 0 0 16px 0;
`;

const Filters = styled.div`
  display: grid;
  grid-template-columns: repeat(6, minmax(140px, 1fr));
  gap: 12px;
  margin-bottom: 12px;
  @media (max-width: 1000px) {
    grid-template-columns: repeat(2, minmax(140px, 1fr));
  }
`;

const Input = styled.input`
  width: 100%;
  padding: 8px 10px;
  border: 1px solid ${p => p.theme.borderColor};
  background: ${p => p.theme.backgroundColor};
  color: ${p => p.theme.textColor};
  border-radius: 6px;
`;

const Select = styled.select`
  width: 100%;
  padding: 8px 10px;
  border: 1px solid ${p => p.theme.borderColor};
  background: ${p => p.theme.backgroundColor};
  color: ${p => p.theme.textColor};
  border-radius: 6px;
`;

const Actions = styled.div`
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
`;

const Button = styled.button`
  padding: 8px 12px;
  background: ${p => p.primary ? p.theme.primaryColor : p.theme.cardBackground};
  color: ${p => p.primary ? 'white' : p.theme.textColor};
  border: 1px solid ${p => p.primary ? p.theme.primaryColor : p.theme.borderColor};
  border-radius: 6px;
  cursor: pointer;
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
`;

const Th = styled.th`
  text-align: left;
  padding: 10px;
  border-bottom: 1px solid ${p => p.theme.borderColor};
  font-weight: 600;
`;

const Td = styled.td`
  padding: 10px;
  border-bottom: 1px solid ${p => p.theme.borderColor};
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 13px;
`;

const Footer = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
`;

const Pagination = styled.div`
  display: flex;
  gap: 8px;
  align-items: center;
`;

const ErrorText = styled.div`
  color: #d33;
  margin-bottom: 8px;
`;

const actorOptions = [
  { value: '', label: 'Любой источник' },
  { value: 'super_admin', label: 'Админ' },
  { value: 'limited_admin', label: 'Второй админ' },
  { value: 'cashier', label: 'Кассир' },
  { value: 'cleaner', label: 'Уборщик' },
  { value: 'user', label: 'Пользователь' },
  { value: 'anpr', label: 'ANPR камера' },
  { value: 'system', label: 'Система' },
];

const actionOptions = [
  { value: '', label: 'Любое действие' },
  { value: 'status_change', label: 'Смена статуса' },
  { value: 'light_on', label: 'Свет включен' },
  { value: 'light_off', label: 'Свет выключен' },
  { value: 'chemistry_on', label: 'Химия включена' },
  { value: 'chemistry_off', label: 'Химия выключена' },
];

const formatDateTime = (iso) => {
  try {
    const d = new Date(iso);
    const dd = d.toLocaleDateString();
    const tt = d.toLocaleTimeString();
    return `${dd} ${tt}`;
  } catch {
    return iso;
  }
};

const translateStatus = (status) => {
  switch ((status || '').toLowerCase()) {
    case 'free': return 'Свободен';
    case 'reserved': return 'Зарезервирован';
    case 'busy': return 'Занят';
    case 'maintenance': return 'На обслуживании';
    case 'cleaning': return 'На уборке';
    default: return status || '-';
  }
};

const translateAction = (action) => {
  switch (action) {
    case 'status_change': return 'Изменение статуса';
    case 'light_on': return 'Свет включен';
    case 'light_off': return 'Свет выключен';
    case 'chemistry_on': return 'Химия включена';
    case 'chemistry_off': return 'Химия выключена';
    default: return action || '-';
  }
};

const translateActor = (actor) => {
  switch ((actor || '').toLowerCase()) {
    case 'super_admin': return 'Админ';
    case 'limited_admin': return 'Второй админ';
    case 'cashier': return 'Кассир';
    case 'cleaner': return 'Уборщик';
    case 'user': return 'Пользователь';
    case 'anpr': return 'ANPR камера';
    case 'system': return 'Система';
    default: return actor || '-';
  }
};

const getStr = (obj, snake, pascal) => obj[snake] ?? obj[pascal] ?? '';
const getNum = (obj, snake, pascal) => obj[snake] ?? obj[pascal] ?? '';
const getBool = (obj, snake, pascal) => {
  const v = obj[snake];
  if (v === true || v === false) return v;
  const p = obj[pascal];
  if (p === true || p === false) return p;
  return undefined;
};

const WashboxChangeLogs = ({ theme }) => {
  const [filters, setFilters] = useState({
    boxNumber: '',
    actorType: '',
    action: '',
    dateFrom: '',
    dateTo: '',
    limit: 50,
    offset: 0,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [rows, setRows] = useState([]);
  const [total, setTotal] = useState(0);

  const canPrev = useMemo(() => filters.offset > 0, [filters.offset]);
  const canNext = useMemo(() => (filters.offset + filters.limit) < total, [filters.offset, filters.limit, total]);

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const params = {
        boxNumber: filters.boxNumber ? Number(filters.boxNumber) : undefined,
        actorType: filters.actorType || undefined,
        action: filters.action || undefined,
        dateFrom: filters.dateFrom ? new Date(filters.dateFrom).toISOString() : undefined,
        dateTo: filters.dateTo ? new Date(filters.dateTo).toISOString() : undefined,
        limit: filters.limit,
        offset: filters.offset,
      };
      const data = await ApiService.getWashboxChangeLogs(params);
      setRows(data.logs || []);
      setTotal(data.total || 0);
    } catch (e) {
      setError('Не удалось загрузить логи. Попробуйте позже.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onApply = () => {
    setFilters((f) => ({ ...f, offset: 0 }));
    setTimeout(load, 0);
  };

  const onReset = () => {
    setFilters({
      boxNumber: '',
      actorType: '',
      action: '',
      dateFrom: '',
      dateTo: '',
      limit: 50,
      offset: 0,
    });
    setTimeout(load, 0);
  };

  const goPrev = () => {
    if (!canPrev) return;
    setFilters((f) => ({ ...f, offset: Math.max(0, f.offset - f.limit) }));
    setTimeout(load, 0);
  };
  const goNext = () => {
    if (!canNext) return;
    setFilters((f) => ({ ...f, offset: f.offset + f.limit }));
    setTimeout(load, 0);
  };

  return (
    <Container theme={theme}>
      <Title>История изменений боксов</Title>
      {error && <ErrorText>{error}</ErrorText>}

      <Filters>
        <Input
          theme={theme}
          type="number"
          min="1"
          placeholder="Номер бокса"
          value={filters.boxNumber}
          onChange={(e) => setFilters({ ...filters, boxNumber: e.target.value })}
        />
        <Select
          theme={theme}
          value={filters.actorType}
          onChange={(e) => setFilters({ ...filters, actorType: e.target.value })}
        >
          {actorOptions.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
        </Select>
        <Select
          theme={theme}
          value={filters.action}
          onChange={(e) => setFilters({ ...filters, action: e.target.value })}
        >
          {actionOptions.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
        </Select>
        <Input
          theme={theme}
          type="datetime-local"
          placeholder="Дата с"
          value={filters.dateFrom}
          onChange={(e) => setFilters({ ...filters, dateFrom: e.target.value })}
        />
        <Input
          theme={theme}
          type="datetime-local"
          placeholder="Дата по"
          value={filters.dateTo}
          onChange={(e) => setFilters({ ...filters, dateTo: e.target.value })}
        />
        <Select
          theme={theme}
          value={filters.limit}
          onChange={(e) => setFilters({ ...filters, limit: Number(e.target.value), offset: 0 })}
        >
          {[20, 50, 100, 200].map(n => <option key={n} value={n}>{n} на странице</option>)}
        </Select>
      </Filters>

      <Actions>
        <Button primary onClick={onApply}>Применить</Button>
        <Button onClick={onReset}>Сбросить</Button>
      </Actions>

      <div style={{ overflowX: 'auto' }}>
        <Table>
          <thead>
            <tr>
              <Th>Время</Th>
              <Th>Бокс</Th>
              <Th>Действие</Th>
              <Th>Источник</Th>
              <Th>Было</Th>
              <Th>Стало</Th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><Td colSpan="6">Загрузка...</Td></tr>
            ) : rows.length === 0 ? (
              <tr><Td colSpan="6">Нет данных</Td></tr>
            ) : (
              rows.map((r) => {
                const isStatus = getStr(r, 'action', 'Action') === 'status_change';
                const createdAtStr = getStr(r, 'created_at', 'CreatedAt');
                const boxNumber = getNum(r, 'box_number', 'BoxNumber');
                const action = getStr(r, 'action', 'Action');
                const actor = getStr(r, 'actor_type', 'ActorType');
                const prevStatus = translateStatus(getStr(r, 'prev_status', 'PrevStatus'));
                const newStatus = translateStatus(getStr(r, 'new_status', 'NewStatus'));
                const prevValue = getBool(r, 'prev_value', 'PrevValue');
                const newValue = getBool(r, 'new_value', 'NewValue');
                return (
                  <tr key={getStr(r, 'id', 'ID')}>
                    <Td>{createdAtStr ? formatDateTime(createdAtStr) : '-'}</Td>
                    <Td>{boxNumber}</Td>
                    <Td>{translateAction(action)}</Td>
                    <Td>{translateActor(actor)}</Td>
                    <Td>{isStatus ? prevStatus : (prevValue === undefined || prevValue === null ? '-' : (prevValue ? 'on' : 'off'))}</Td>
                    <Td>{isStatus ? newStatus : (newValue === undefined || newValue === null ? '-' : (newValue ? 'on' : 'off'))}</Td>
                  </tr>
                );
              })
            )}
          </tbody>
        </Table>
      </div>

      <Footer>
        <div>Всего: {total}</div>
        <Pagination>
          <Button onClick={goPrev} disabled={!canPrev}>Назад</Button>
          <div>{filters.offset + 1} - {Math.min(filters.offset + filters.limit, total)} из {total}</div>
          <Button onClick={goNext} disabled={!canNext}>Вперед</Button>
        </Pagination>
      </Footer>
    </Container>
  );
};

export default WashboxChangeLogs;


