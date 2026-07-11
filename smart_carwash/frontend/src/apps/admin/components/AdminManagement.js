import React, { useState, useEffect, useCallback } from 'react';
import styled from 'styled-components';
import AuthService from '../../../shared/services/AuthService';

// Разделы админки для галочек доступа (ключи совпадают с роутами/меню AdminApp).
const SECTION_OPTIONS = [
  { key: 'dashboard', label: 'Панель управления' },
  { key: 'washboxes', label: 'Боксы мойки' },
  { key: 'maintenance', label: 'ТО аппаратов' },
  { key: 'service-tickets', label: 'Сервисные наряды' },
  { key: 'sessions', label: 'Сессии мойки' },
  { key: 'queue', label: 'Очередь' },
  { key: 'users', label: 'Клиенты' },
  { key: 'cashiers', label: 'Управление кассирами' },
  { key: 'cleaners', label: 'Управление уборщиками' },
  { key: 'cleaning-logs', label: 'Логи уборки' },
  { key: 'washbox-change-logs', label: 'История боксов' },
  { key: 'payments', label: 'Платежи' },
  { key: 'settings', label: 'Настройки' },
  { key: 'modbus-dashboard', label: 'Modbus мониторинг' },
  { key: 'my-wash', label: 'Моя мойка' },
];

const Container = styled.div`
  padding: 16px;
`;

const Title = styled.h2`
  margin: 0 0 16px 0;
`;

const Card = styled.div`
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
`;

const Row = styled.div`
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  align-items: center;
  margin-bottom: 12px;
`;

const Input = styled.input`
  padding: 8px 10px;
  border: 1px solid #ccc;
  border-radius: 6px;
  font-size: 0.95rem;
`;

const SectionsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 6px 16px;
  margin: 10px 0;
`;

const CheckLabel = styled.label`
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  cursor: pointer;
`;

const Button = styled.button`
  padding: 9px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 600;
  background: ${(p) => (p.$secondary ? '#eceff1' : '#1976d2')};
  color: ${(p) => (p.$secondary ? '#333' : '#fff')};
  &:disabled { opacity: 0.6; cursor: not-allowed; }
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  th, td { text-align: left; padding: 8px 10px; border-bottom: 1px solid #eee; font-size: 0.9rem; vertical-align: top; }
  th { color: #666; font-weight: 600; }
`;

const Badge = styled.span`
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 0.78rem;
  background: ${(p) => (p.$active ? '#E8F5E8' : '#FFEBEE')};
  color: ${(p) => (p.$active ? '#2E7D32' : '#C62828')};
`;

const ErrorMsg = styled.div`
  background: #FFEBEE; border: 1px solid #EF9A9A; color: #C62828;
  padding: 10px 12px; border-radius: 6px; margin-bottom: 12px; font-size: 0.9rem;
`;

const emptyForm = { username: '', password: '', display_name: '', sections: [] };

const AdminManagement = () => {
  const [admins, setAdmins] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [editId, setEditId] = useState(null);

  const load = useCallback(async () => {
    try {
      setError('');
      const list = await AuthService.getAdmins();
      setAdmins(list);
    } catch (e) {
      setError('Не удалось загрузить список: ' + (e.response?.data?.error || e.message));
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  const toggleSection = (key) => {
    setForm((f) => ({
      ...f,
      sections: f.sections.includes(key)
        ? f.sections.filter((s) => s !== key)
        : [...f.sections, key],
    }));
  };

  const resetForm = () => { setForm(emptyForm); setEditId(null); };

  const handleCreate = async () => {
    if (!form.username || !form.password) {
      setError('Укажите логин и пароль');
      return;
    }
    setLoading(true);
    setError('');
    try {
      await AuthService.createAdmin({
        username: form.username.trim(),
        password: form.password,
        display_name: form.display_name.trim(),
        allowed_sections: form.sections,
      });
      resetForm();
      await load();
    } catch (e) {
      setError(e.response?.data?.error || e.message);
    } finally {
      setLoading(false);
    }
  };

  const startEdit = (a) => {
    setEditId(a.id);
    setForm({ username: a.username, password: '', display_name: a.display_name || '', sections: a.allowed_sections || [] });
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleSaveEdit = async () => {
    setLoading(true);
    setError('');
    try {
      await AuthService.updateAdmin(editId, {
        display_name: form.display_name.trim(),
        allowed_sections: form.sections,
        ...(form.password ? { password: form.password } : {}),
      });
      resetForm();
      await load();
    } catch (e) {
      setError(e.response?.data?.error || e.message);
    } finally {
      setLoading(false);
    }
  };

  const toggleActive = async (a) => {
    setError('');
    try {
      await AuthService.updateAdmin(a.id, { is_active: !a.is_active });
      await load();
    } catch (e) {
      setError(e.response?.data?.error || e.message);
    }
  };

  return (
    <Container>
      <Title>Управление администраторами</Title>
      {error && <ErrorMsg>{error}</ErrorMsg>}

      <Card>
        <h3 style={{ marginTop: 0 }}>{editId ? 'Редактировать учётку' : 'Новая учётка'}</h3>
        <Row>
          <Input
            placeholder="Логин"
            value={form.username}
            disabled={!!editId}
            onChange={(e) => setForm({ ...form, username: e.target.value })}
          />
          <Input
            placeholder={editId ? 'Новый пароль (если менять)' : 'Пароль'}
            type="text"
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
          />
          <Input
            placeholder="Имя (напр. Костя)"
            value={form.display_name}
            onChange={(e) => setForm({ ...form, display_name: e.target.value })}
          />
        </Row>
        <div style={{ fontWeight: 600, fontSize: '0.9rem', marginBottom: 6 }}>Доступные разделы:</div>
        <SectionsGrid>
          {SECTION_OPTIONS.map((s) => (
            <CheckLabel key={s.key}>
              <input
                type="checkbox"
                checked={form.sections.includes(s.key)}
                onChange={() => toggleSection(s.key)}
              />
              {s.label}
            </CheckLabel>
          ))}
        </SectionsGrid>
        <Row>
          {editId ? (
            <>
              <Button onClick={handleSaveEdit} disabled={loading}>Сохранить</Button>
              <Button $secondary onClick={resetForm} disabled={loading}>Отмена</Button>
            </>
          ) : (
            <Button onClick={handleCreate} disabled={loading}>Создать</Button>
          )}
        </Row>
      </Card>

      <Card>
        <h3 style={{ marginTop: 0 }}>Учётки</h3>
        <Table>
          <thead>
            <tr>
              <th>Логин</th>
              <th>Имя</th>
              <th>Разделы</th>
              <th>Статус</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {admins.length === 0 ? (
              <tr><td colSpan={5} style={{ color: '#999' }}>Пока нет учёток</td></tr>
            ) : (
              admins.map((a) => (
                <tr key={a.id}>
                  <td>{a.username}</td>
                  <td>{a.display_name || '—'}</td>
                  <td style={{ maxWidth: 320, fontSize: '0.82rem', color: '#555' }}>
                    {(a.allowed_sections && a.allowed_sections.length)
                      ? a.allowed_sections.map((k) => (SECTION_OPTIONS.find((o) => o.key === k) || {}).label || k).join(', ')
                      : '—'}
                  </td>
                  <td><Badge $active={a.is_active}>{a.is_active ? 'активен' : 'выключен'}</Badge></td>
                  <td>
                    <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                      <Button $secondary onClick={() => startEdit(a)}>Изменить</Button>
                      <Button $secondary onClick={() => toggleActive(a)}>
                        {a.is_active ? 'Выключить' : 'Включить'}
                      </Button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </Table>
      </Card>
    </Container>
  );
};

export default AdminManagement;
