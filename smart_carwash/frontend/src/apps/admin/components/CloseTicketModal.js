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

// Модалка закрытия наряда: выбор выполненных работ + комментарий мастера.
// Не-поломки (is_breakdown=false) закрываются без работ.
const CloseTicketModal = ({ ticket, onClose, onClosed }) => {
  const [catalog, setCatalog] = useState([]);
  const [selected, setSelected] = useState({}); // componentId -> { checked, action }
  const [masterComment, setMasterComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  const requiresWorks = ticket.is_breakdown !== false;

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
