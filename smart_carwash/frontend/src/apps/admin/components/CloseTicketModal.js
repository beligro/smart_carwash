import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import ApiService from '../../../shared/services/ApiService';

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

const Title = styled.h3`
  margin: 0 0 12px 0;
`;

const Muted = styled.span`
  color: #888;
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

const Select = styled.select`
  padding: 6px 10px;
  border: 1px solid #ccc;
  border-radius: 6px;
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

const ErrorMessage = styled.div`
  color: #dc3545;
  background: #f8d7da;
  border: 1px solid #f5c6cb;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 14px;
`;

const TestBlock = styled.div`
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 12px 14px;
  margin-bottom: 16px;
  background: #fafafa;
`;

const TestTitle = styled.div`
  font-weight: 600;
  color: #1565c0;
  margin-bottom: 4px;
`;

const TestRow = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 0;
`;

const TestLabel = styled.span`
  flex: 1;
`;

const OnButton = styled.button`
  padding: 6px 14px;
  border: none;
  border-radius: 6px;
  background: #2e7d32;
  color: #fff;
  cursor: pointer;
  &:disabled { opacity: 0.5; cursor: not-allowed; }
`;

const OffButton = styled.button`
  padding: 6px 14px;
  border: 1px solid #bbb;
  border-radius: 6px;
  background: #f1f3f4;
  color: #333;
  cursor: pointer;
  &:disabled { opacity: 0.5; cursor: not-allowed; }
`;

const Countdown = styled.span`
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  color: #b26a00;
  min-width: 44px;
  text-align: right;
`;

const TEST_SECONDS = 120;

const fmtMMSS = (s) => {
  const m = Math.floor(s / 60);
  const sec = s % 60;
  return `${m}:${String(sec).padStart(2, '0')}`;
};

// Модалка закрытия наряда: выбор выполненных работ + комментарий мастера.
// Не-поломки (is_breakdown=false) закрываются без работ.
const CloseTicketModal = ({ ticket, onClose, onClosed }) => {
  const [catalog, setCatalog] = useState([]);
  const [selected, setSelected] = useState({}); // componentId -> { checked, action }
  const [masterComment, setMasterComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  // Тестовое включение коилов из наряда: локальный обратный отсчёт 2:00 на каждый коил.
  const [coilSeconds, setCoilSeconds] = useState({ light: 0, chemistry: 0 });
  const [coilBusy, setCoilBusy] = useState({ light: false, chemistry: false });
  const [coilError, setCoilError] = useState(null);

  const requiresWorks = ticket.is_breakdown !== false;
  const showChemistry = ticket.box_type === 'wash';

  useEffect(() => {
    const active = coilSeconds.light > 0 || coilSeconds.chemistry > 0;
    if (!active) return undefined;
    const timer = setInterval(() => {
      setCoilSeconds(prev => ({
        light: prev.light > 0 ? prev.light - 1 : 0,
        chemistry: prev.chemistry > 0 ? prev.chemistry - 1 : 0,
      }));
    }, 1000);
    return () => clearInterval(timer);
  }, [coilSeconds.light > 0, coilSeconds.chemistry > 0]); // eslint-disable-line react-hooks/exhaustive-deps

  const toggleCoil = async (coil, value) => {
    setCoilBusy(prev => ({ ...prev, [coil]: true }));
    setCoilError(null);
    try {
      await ApiService.testTicketCoil({
        box_id: ticket.box_id,
        coil,
        value,
        ticket_id: ticket.id,
      });
      setCoilSeconds(prev => ({ ...prev, [coil]: value ? TEST_SECONDS : 0 }));
    } catch (e) {
      setCoilError('Ошибка теста коила: ' + (e.response?.data?.error || e.message));
    } finally {
      setCoilBusy(prev => ({ ...prev, [coil]: false }));
    }
  };

  useEffect(() => {
    const load = async () => {
      try {
        const resp = await ApiService.getServiceCatalog(ticket.box_number);
        setCatalog(resp.catalog || []);
      } catch (e) {
        setError('Не удалось загрузить справочник работ');
      }
    };
    load();
  }, [ticket.box_number]);

  const toggle = (id) => {
    setSelected(prev => ({
      ...prev,
      [id]: { checked: !prev[id]?.checked, action: prev[id]?.action || 'replace' },
    }));
  };

  const setAction = (id, action) => {
    setSelected(prev => ({
      ...prev,
      [id]: { checked: prev[id]?.checked ?? true, action },
    }));
  };

  const works = Object.entries(selected)
    .filter(([, v]) => v.checked)
    .map(([componentId, v]) => ({ component_id: Number(componentId), action: v.action }));

  const submit = async () => {
    if (requiresWorks && works.length === 0) {
      setError('Отметьте хотя бы одну выполненную работу');
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      await ApiService.closeServiceTicket(ticket.id, { works, master_comment: masterComment });
      onClosed();
    } catch (e) {
      setError('Ошибка закрытия наряда: ' + (e.response?.data?.error || e.message));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Overlay onClick={submitting ? undefined : onClose}>
      <Modal onClick={(e) => e.stopPropagation()}>
        <Title>Закрыть наряд — бокс #{ticket.box_number}</Title>
        {error && <ErrorMessage>{error}</ErrorMessage>}

        <TestBlock>
          <TestTitle>Проверка коилов (тест)</TestTitle>
          <Muted style={{ fontSize: '0.82rem' }}>
            Кратковременное включение для проверки. Авто-выключение через 2 мин.
          </Muted>
          {[
            { coil: 'light', label: 'Бокс (свет/вода)', show: true },
            { coil: 'chemistry', label: 'Химия', show: showChemistry },
          ].filter(r => r.show).map(({ coil, label }) => {
            const secs = coilSeconds[coil];
            return (
              <TestRow key={coil}>
                <TestLabel>{label}</TestLabel>
                {secs > 0 && <Countdown>{fmtMMSS(secs)}</Countdown>}
                <OnButton
                  onClick={() => toggleCoil(coil, true)}
                  disabled={coilBusy[coil]}
                >ВКЛ</OnButton>
                <OffButton
                  onClick={() => toggleCoil(coil, false)}
                  disabled={coilBusy[coil]}
                >ВЫКЛ</OffButton>
              </TestRow>
            );
          })}
          {coilError && (
            <div style={{ color: '#dc3545', fontSize: '0.82rem', marginTop: 6 }}>{coilError}</div>
          )}
        </TestBlock>

        {!requiresWorks && (
          <div style={{ color: '#1565c0', fontSize: '0.9rem', marginBottom: 10 }}>
            Не поломка — работы не требуются, наряд не идёт в статистику. Можно просто закрыть.
          </div>
        )}
        <Muted>Отметьте выполненные работы (замена по умолчанию):</Muted>
        <div style={{ marginTop: 12 }}>
          {catalog.map(group => (
            <GroupBlock key={group.id}>
              <GroupName>{group.name} <Muted style={{ fontWeight: 400, fontSize: '0.8rem' }}>({group.carrier === 'machine' ? 'аппарат' : group.carrier === 'box' ? 'бокс' : 'общее'})</Muted></GroupName>
              {(group.children || []).map(child => (
                <WorkRow key={child.id}>
                  <input
                    type="checkbox"
                    checked={selected[child.id]?.checked || false}
                    onChange={() => toggle(child.id)}
                  />
                  <span style={{ flex: 1 }}>{child.name}</span>
                  {selected[child.id]?.checked && (
                    (child.repairable || child.cleanable) ? (
                      <Select
                        value={selected[child.id]?.action || 'replace'}
                        onChange={(e) => setAction(child.id, e.target.value)}
                      >
                        <option value="replace">замена</option>
                        {child.repairable && <option value="repair">ремонт</option>}
                        {child.cleanable && <option value="clean">чистка</option>}
                      </Select>
                    ) : (
                      <Muted style={{ fontSize: '0.82rem' }}>только замена</Muted>
                    )
                  )}
                </WorkRow>
              ))}
            </GroupBlock>
          ))}
        </div>

        <label style={{ fontWeight: 500 }}>Комментарий мастера</label>
        <TextArea
          value={masterComment}
          onChange={(e) => setMasterComment(e.target.value)}
          placeholder="Что сделано, детали"
        />

        {requiresWorks && works.length === 0 && (
          <div style={{ color: '#b26a00', fontSize: '0.85rem', marginBottom: 10 }}>
            Отметьте хотя бы одну работу. Если ничего не меняли — выберите «Общее → Другое (комментарий)».
          </div>
        )}
        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end' }}>
          <SecondaryButton onClick={onClose} disabled={submitting}>Отмена</SecondaryButton>
          <PrimaryButton onClick={submit} disabled={submitting || (requiresWorks && works.length === 0)}>
            {submitting ? 'Закрываем...' : 'Закрыть наряд и вернуть в работу'}
          </PrimaryButton>
        </div>
      </Modal>
    </Overlay>
  );
};

export default CloseTicketModal;
