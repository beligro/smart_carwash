import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import AuthService from '../../../shared/services/AuthService';
import ApiService from '../../../shared/services/ApiService';

const Container = styled.div`
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
`;

const Card = styled.div`
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
`;

const Title = styled.h2`
  margin: 0 0 20px 0;
  color: ${props => props.theme.textColor};
  font-size: 1.5rem;
`;

const Button = styled.button`
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;

  &:hover:not(:disabled) {
    background-color: ${props => props.theme.primaryColorHover};
  }

  &:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  margin-top: 16px;
`;

const Th = styled.th`
  text-align: left;
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  font-weight: 600;
  color: ${props => props.theme.textColor};
`;

const Td = styled.td`
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  color: ${props => props.theme.textColor};
`;

const StatusBadge = styled.span`
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  background-color: ${props => props.active ? '#28a745' : '#dc3545'};
  color: white;
`;

const ActionButton = styled.button`
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  margin-right: 8px;

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &.edit {
    background-color: #007bff;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #0056b3;
    }
  }

  &.delete {
    background-color: #dc3545;
    color: white;
    
    &:hover:not(:disabled) {
      background-color: #c82333;
    }
  }
`;

const Modal = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
`;

const ModalContent = styled.div`
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 24px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
`;

const ModalTitle = styled.h3`
  margin: 0 0 20px 0;
  color: ${props => props.theme.textColor};
`;

const FormGroup = styled.div`
  margin-bottom: 16px;
`;

const Label = styled.label`
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: ${props => props.theme.textColor};
`;

const Input = styled.input`
  width: 100%;
  padding: 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 6px;
  font-size: 1rem;
  background-color: ${props => props.theme.backgroundColor};
  color: ${props => props.theme.textColor};

  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }
`;

const Checkbox = styled.input`
  margin-right: 8px;
`;

const ModalFooter = styled.div`
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
  gap: 10px;
`;

const ErrorMessage = styled.div`
  color: #dc3545;
  padding: 16px;
  background-color: #f8d7da;
  border: 1px solid #f5c6cb;
  border-radius: 6px;
  margin-bottom: 16px;
`;

const SuccessMessage = styled.div`
  color: #155724;
  padding: 16px;
  background-color: #d4edda;
  border: 1px solid #c3e6cb;
  border-radius: 6px;
  margin-bottom: 16px;
`;

/**
 * Компонент управления уборщиками
 * @returns {React.ReactNode} - Компонент управления уборщиками
 */
const CleanerManagement = () => {
  const theme = getTheme('light');
  const [cleaners, setCleaners] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  
  // Состояние для модального окна создания уборщика
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newCleaner, setNewCleaner] = useState({
    username: '',
    password: '',
  });
  
  // Состояние для модального окна редактирования уборщика
  const [showEditModal, setShowEditModal] = useState(false);
  const [editCleaner, setEditCleaner] = useState({
    id: '',
    username: '',
    password: '',
    is_active: true,
  });
  
  // Загрузка списка уборщиков при монтировании компонента
  useEffect(() => {
    fetchCleaners();
  }, []);
  
  // Функция для загрузки списка уборщиков
  const fetchCleaners = async () => {
    try {
      setLoading(true);
      const response = await AuthService.getCleaners();
      setCleaners(response.cleaners);
      setError('');
    } catch (err) {
      setError('Ошибка при загрузке списка уборщиков');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  // Обработчик изменения полей формы создания уборщика
  const handleCreateChange = (e) => {
    const { name, value } = e.target;
    setNewCleaner(prev => ({
      ...prev,
      [name]: value,
    }));
  };
  
  // Обработчик изменения полей формы редактирования уборщика
  const handleEditChange = (e) => {
    const { name, value, type, checked } = e.target;
    setEditCleaner(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));
  };
  
  // Обработчик создания уборщика
  const handleCreateCleaner = async (e) => {
    e.preventDefault();
    
    if (!newCleaner.username || !newCleaner.password) {
      setError('Все поля обязательны для заполнения');
      return;
    }
    
    try {
      await AuthService.createCleaner(newCleaner);
      setSuccess('Уборщик создан успешно');
      setNewCleaner({ username: '', password: '' });
      setShowCreateModal(false);
      await fetchCleaners();
    } catch (err) {
      setError(err.response?.data?.error || 'Ошибка при создании уборщика');
      console.error(err);
    }
  };
  
  // Обработчик обновления уборщика
  const handleUpdateCleaner = async (e) => {
    e.preventDefault();
    
    if (!editCleaner.username) {
      setError('Имя пользователя обязательно');
      return;
    }
    
    try {
      await AuthService.updateCleaner(editCleaner);
      setSuccess('Уборщик обновлен успешно');
      setShowEditModal(false);
      await fetchCleaners();
    } catch (err) {
      setError(err.response?.data?.error || 'Ошибка при обновлении уборщика');
      console.error(err);
    }
  };
  
  // Обработчик удаления уборщика
  const handleDeleteCleaner = async (cleanerId) => {
    if (!window.confirm('Вы уверены, что хотите удалить этого уборщика?')) {
      return;
    }
    
    try {
      await AuthService.deleteCleaner(cleanerId);
      setSuccess('Уборщик удален успешно');
      await fetchCleaners();
    } catch (err) {
      setError(err.response?.data?.error || 'Ошибка при удалении уборщика');
      console.error(err);
    }
  };
  
  // Функция для открытия модального окна редактирования
  const openEditModal = (cleaner) => {
    setEditCleaner({
      id: cleaner.id,
      username: cleaner.username,
      password: '',
      is_active: cleaner.is_active,
    });
    setShowEditModal(true);
  };
  
  return (
    <Container>
      <Card theme={theme}>
        <Title theme={theme}>Управление уборщиками</Title>
        <Button 
          onClick={() => setShowCreateModal(true)} 
          disabled={loading}
          theme={theme}
        >
          Создать уборщика
        </Button>
        
        {loading && <p>Загрузка...</p>}
        {error && <ErrorMessage theme={theme}>{error}</ErrorMessage>}
        {success && <SuccessMessage theme={theme}>{success}</SuccessMessage>}
        
        {!loading && cleaners.length === 0 ? (
          <p>Нет доступных уборщиков</p>
        ) : (
          <Table>
            <thead>
              <tr>
                <Th theme={theme}>Имя пользователя</Th>
                <Th theme={theme}>Статус</Th>
                <Th theme={theme}>Последний вход</Th>
                <Th theme={theme}>Действия</Th>
              </tr>
            </thead>
            <tbody>
              {cleaners.map(cleaner => (
                <tr key={cleaner.id}>
                  <Td theme={theme}>{cleaner.username}</Td>
                  <Td theme={theme}>
                    <StatusBadge active={cleaner.is_active}>
                      {cleaner.is_active ? 'Активен' : 'Неактивен'}
                    </StatusBadge>
                  </Td>
                  <Td theme={theme}>
                    {cleaner.last_login 
                      ? new Date(cleaner.last_login).toLocaleString('ru-RU')
                      : 'Никогда'
                    }
                  </Td>
                  <Td theme={theme}>
                    <ActionButton
                      className="edit"
                      onClick={() => openEditModal(cleaner)}
                    >
                      Редактировать
                    </ActionButton>
                    <ActionButton
                      className="delete"
                      onClick={() => handleDeleteCleaner(cleaner.id)}
                    >
                      Удалить
                    </ActionButton>
                  </Td>
                </tr>
              ))}
            </tbody>
          </Table>
        )}
      </Card>

      {/* Модальное окно создания уборщика */}
      {showCreateModal && (
        <Modal>
          <ModalContent theme={theme}>
            <ModalTitle theme={theme}>Создать уборщика</ModalTitle>
            <form onSubmit={handleCreateCleaner}>
              <FormGroup>
                <Label theme={theme}>Имя пользователя</Label>
                <Input
                  type="text"
                  name="username"
                  value={newCleaner.username}
                  onChange={handleCreateChange}
                  theme={theme}
                  required
                />
              </FormGroup>
              <FormGroup>
                <Label theme={theme}>Пароль</Label>
                <Input
                  type="password"
                  name="password"
                  value={newCleaner.password}
                  onChange={handleCreateChange}
                  theme={theme}
                  required
                />
              </FormGroup>
              <ModalFooter>
                <Button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  theme={theme}
                  style={{ backgroundColor: '#6c757d' }}
                >
                  Отмена
                </Button>
                <Button type="submit" theme={theme}>
                  Создать
                </Button>
              </ModalFooter>
            </form>
          </ModalContent>
        </Modal>
      )}

      {/* Модальное окно редактирования уборщика */}
      {showEditModal && (
        <Modal>
          <ModalContent theme={theme}>
            <ModalTitle theme={theme}>Редактировать уборщика</ModalTitle>
            <form onSubmit={handleUpdateCleaner}>
              <FormGroup>
                <Label theme={theme}>Имя пользователя</Label>
                <Input
                  type="text"
                  name="username"
                  value={editCleaner.username}
                  onChange={handleEditChange}
                  theme={theme}
                  required
                />
              </FormGroup>
              <FormGroup>
                <Label theme={theme}>Новый пароль (оставьте пустым, чтобы не изменять)</Label>
                <Input
                  type="password"
                  name="password"
                  value={editCleaner.password}
                  onChange={handleEditChange}
                  theme={theme}
                />
              </FormGroup>
              <FormGroup>
                <Label theme={theme}>
                  <Checkbox
                    type="checkbox"
                    name="is_active"
                    checked={editCleaner.is_active}
                    onChange={handleEditChange}
                  />
                  Активен
                </Label>
              </FormGroup>
              <ModalFooter>
                <Button
                  type="button"
                  onClick={() => setShowEditModal(false)}
                  theme={theme}
                  style={{ backgroundColor: '#6c757d' }}
                >
                  Отмена
                </Button>
                <Button type="submit" theme={theme}>
                  Сохранить
                </Button>
              </ModalFooter>
            </form>
          </ModalContent>
        </Modal>
      )}
    </Container>
  );
};

export default CleanerManagement;

