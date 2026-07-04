import React, { useState, useEffect, useCallback } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';
import CloseTicketModal from './CloseTicketModal';

const theme = getTheme('light');

const Container = styled.div`
  padding: 20px;
`;

const Title = styled.h2`
  margin: 0 0 16px 0;
`;

const Tabs = styled.div`
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  border-bottom: 1px solid #e0e0e0;
`;

const Tab = styled.button`
  padding: 10px 18px;
  border: none;
  background: none;
  font-size: 1rem;
  cursor: pointer;
  color: ${p => (p.active ? '#1565c0' : '#666')};
  border-bottom: 3px solid ${p => (p.active ? '#1565c0' : 'transparent')};
  font-weight: ${p => (p.active ? 600 : 400)};
`;

const Filters = styled.div`
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  align-items: center;
`;

const Select = styled.select`
  padding: 8px 12px;
  border: 1px solid #ccc;
  border-radius: 6px;
  font-size: 0.9rem;
`;

const TicketCard = styled.div`
  background: #fff;
  border: 1px solid #e6e6e6;
  border-left: 5px solid ${p => (p.open ? '#dc3545' : '#28a745')};
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 14px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
`;

const TicketHead = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  flex-wrap: wrap;
`;

const BoxTitle = styled.div`
  font-size: 1.15rem;
  font-weight: 700;
`;

const Badge = styled.span`
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 0.78rem;
  font-weight: 600;
  color: #fff;
  background: ${p => (p.open ? '#dc3545' : '#28a745')};
`;

const Row = styled.div`
  margin-top: 6px;
  font-size: 0.9rem;
  color: #333;
`;

const Muted = styled.span`
  color: #888;
`;

const WorksList = styled.ul`
  margin: 8px 0 0 0;
  padding-left: 18px;
  font-size: 0.88rem;
  color: #333;
`;

const PrimaryButton = styled.button`
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  background: #1565c0;
  color: #fff;
  font-weight: 500;
  cursor: pointer;
  &:disabled { opacity: 0.5; cursor: not-allowed; }
`;

const SecondaryButton = styled.button`
  padding: 8px 16px;
  border: 1px solid #ccc;
  border-radius: 6px;
  background: #fff;
  cursor: pointer;
  &:disabled { opacity: 0.5; cursor: not-allowed; }
`;

const Overlay = styled.div`
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.5);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 16px;
`;

const Modal = styled.div`
  background: #fff;
  border-radius: 10px;
  padding: 24px;
  width: 100%;
  max-width: 640px;
  max-height: 85vh;
  overflow-y: auto;
`;

const GroupBlock = styled.div`
  margin-bottom: 14px;
`;

const GroupName = styled.div`
  font-weight: 600;
  margin-bottom: 6px;
  color: #1565c0;
`;

const WorkRow = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 0;
`;

const TextArea = styled.textarea`
  width: 100%;
  min-height: 70px;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 6px;
  box-sizing: border-box;
  margin: 10px 0 16px 0;
`;

const Input = styled.input`
  padding: 8px 12px;
  border: 1px solid #ccc;
  border-radius: 6px;
  font-size: 0.9rem;
`;

const ErrorMessage = styled.div`
  color: #dc3545;
  background: #f8d7da;
  border: 1px solid #f5c6cb;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 14px;
`;

const RecipientRow = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid #eee;
`;

const fmtNsk = (iso) => {
  if (!iso) return '—';
  try {
    return new Date(iso).toLocaleString('ru-RU', { timeZone: 'Asia/Novosibirsk' });
  } catch {
    return iso;
  }
};

const fmtDowntime = (minutes) => {
  if (minutes == null) return '—';
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h > 0) return `${h} ч ${m} мин`;
  return `${m} мин`;
};

const actionLabel = (a) => (a === 'repair' ? 'ремонт' : a === 'clean' ? 'чистка' : 'замена');

// Модалка закрытия наряда вынесена в общий компонент ./CloseTicketModal
// (переиспользуется на странице «ТО аппаратов»).

// ==================== Журнал нарядов ====================
const TicketsJournal = () => {
  const [tickets, setTickets] = useState([]);
  const [statusFilter, setStatusFilter] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [closingTicket, setClosingTicket] = useState(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const filters = {};
      if (statusFilter) filters.status = statusFilter;
      const resp = await ApiService.getServiceTickets(filters);
      setTickets(resp.tickets || []);
    } catch (e) {
      setError('Ошибка загрузки нарядов');
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => { load(); }, [load]);

  return (
    <div>
      <Filters>
        <span>Статус:</span>
        <Select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="">Все</option>
          <option value="open">Открытые</option>
          <option value="closed">Закрытые</option>
        </Select>
        <SecondaryButton onClick={load}>Обновить</SecondaryButton>
      </Filters>

      {error && <ErrorMessage>{error}</ErrorMessage>}
      {loading ? (
        <div>Загрузка...</div>
      ) : tickets.length === 0 ? (
        <Muted>Нарядов нет</Muted>
      ) : (
        tickets.map(t => {
          const isOpen = t.status === 'open';
          return (
            <TicketCard key={t.id} open={isOpen}>
              <TicketHead>
                <div>
                  <BoxTitle>Бокс #{t.box_number} <Badge open={isOpen}>{isOpen ? 'В сервисе' : 'Закрыт'}</Badge></BoxTitle>
                  <Row><Muted>Поставил:</Muted> {t.opened_by || '—'} · {fmtNsk(t.opened_at)}</Row>
                  <Row><Muted>Симптом:</Muted> {t.symptom_name ? `${t.symptom_group}: ${t.symptom_name}` : '—'}</Row>
                  {t.cashier_comment && <Row><Muted>Коммент кассира:</Muted> {t.cashier_comment}</Row>}
                  {!isOpen && (
                    <>
                      <Row><Muted>Закрыл:</Muted> {t.closed_by || '—'} · {fmtNsk(t.closed_at)} · <Muted>простой</Muted> {fmtDowntime(t.downtime_minutes)}</Row>
                      {t.master_comment && <Row><Muted>Коммент мастера:</Muted> {t.master_comment}</Row>}
                      {t.works && t.works.length > 0 && (
                        <WorksList>
                          {t.works.map((w, i) => (
                            <li key={i}>{w.group_name}: {w.component_name} — {actionLabel(w.action)}{w.motor_hours != null ? ` (${w.motor_hours} мч)` : ''}</li>
                          ))}
                        </WorksList>
                      )}
                    </>
                  )}
                </div>
                {isOpen && (
                  <PrimaryButton onClick={() => setClosingTicket(t)}>Закрыть наряд</PrimaryButton>
                )}
              </TicketHead>
            </TicketCard>
          );
        })
      )}

      {closingTicket && (
        <CloseTicketModal
          ticket={closingTicket}
          onClose={() => setClosingTicket(null)}
          onClosed={() => { setClosingTicket(null); load(); }}
        />
      )}
    </div>
  );
};

// ==================== Получатели уведомлений ====================
const RecipientsManager = () => {
  const [recipients, setRecipients] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [newName, setNewName] = useState('');
  const [newChatId, setNewChatId] = useState('');
  const [adding, setAdding] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await ApiService.getNotificationRecipients();
      setRecipients(resp.recipients || []);
    } catch (e) {
      setError('Ошибка загрузки получателей');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  const toggle = async (rec) => {
    try {
      await ApiService.setNotificationRecipientActive(rec.id, !rec.is_active);
      load();
    } catch (e) {
      setError('Не удалось изменить получателя');
    }
  };

  const add = async () => {
    if (!newName.trim() || !newChatId.trim()) return;
    setAdding(true);
    setError(null);
    try {
      await ApiService.createNotificationRecipient({ name: newName.trim(), chat_id: newChatId.trim() });
      setNewName('');
      setNewChatId('');
      load();
    } catch (e) {
      setError('Не удалось добавить получателя');
    } finally {
      setAdding(false);
    }
  };

  return (
    <div>
      <Muted>Получатели push-уведомлений при постановке бокса в сервис.</Muted>
      {error && <ErrorMessage style={{ marginTop: 12 }}>{error}</ErrorMessage>}

      <div style={{ marginTop: 16, marginBottom: 20 }}>
        {loading ? (
          <div>Загрузка...</div>
        ) : recipients.length === 0 ? (
          <Muted>Получателей нет</Muted>
        ) : (
          recipients.map(r => (
            <RecipientRow key={r.id}>
              <input type="checkbox" checked={r.is_active} onChange={() => toggle(r)} />
              <span style={{ flex: 1 }}>
                <strong>{r.name}</strong> <Muted>· chat_id {r.chat_id}</Muted>
              </span>
              <Badge open={!r.is_active} style={{ background: r.is_active ? '#28a745' : '#999' }}>
                {r.is_active ? 'включён' : 'выключен'}
              </Badge>
            </RecipientRow>
          ))
        )}
      </div>

      <div style={{ display: 'flex', gap: 10, alignItems: 'center', flexWrap: 'wrap' }}>
        <Input placeholder="Имя" value={newName} onChange={(e) => setNewName(e.target.value)} />
        <Input placeholder="chat_id (число)" value={newChatId} onChange={(e) => setNewChatId(e.target.value)} />
        <PrimaryButton onClick={add} disabled={adding}>{adding ? 'Добавляем...' : 'Добавить'}</PrimaryButton>
      </div>
    </div>
  );
};

// ==================== Корневой компонент ====================
const ServiceTicketsManagement = () => {
  const [tab, setTab] = useState('journal');

  return (
    <Container theme={theme}>
      <Title>Сервисные наряды</Title>
      <Tabs>
        <Tab active={tab === 'journal'} onClick={() => setTab('journal')}>Журнал нарядов</Tab>
        <Tab active={tab === 'recipients'} onClick={() => setTab('recipients')}>Получатели уведомлений</Tab>
      </Tabs>

      {tab === 'journal' ? <TicketsJournal /> : <RecipientsManager />}
    </Container>
  );
};

export default ServiceTicketsManagement;
