import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import AuthService from '../../../shared/services/AuthService';

const Container = styled.div`
  display: flex;
  flex-direction: column;
`;

const Card = styled.div`
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
`;

const Title = styled.h2`
  margin-top: 0;
  margin-bottom: 15px;
  color: ${props => props.theme.textColor};
`;

const Button = styled.button`
  padding: 8px 16px;
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  
  &:hover {
    background-color: ${props => props.theme.primaryColorHover};
  }
  
  &:disabled {
    background-color: ${props => props.theme.disabledColor};
    cursor: not-allowed;
  }
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
`;

const Th = styled.th`
  text-align: left;
  padding: 12px;
  border-bottom: 2px solid ${props => props.theme.borderColor};
  color: ${props => props.theme.textColor};
`;

const Td = styled.td`
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  color: ${props => props.theme.textColor};
`;

const ActionButton = styled.button`
  padding: 6px 12px;
  background-color: ${props => props.danger ? props.theme.dangerColor : props.theme.secondaryColor};
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
  font-size: 12px;
  
  &:hover {
    background-color: ${props => props.danger ? props.theme.dangerColorHover : props.theme.secondaryColorHover};
  }
`;

const Form = styled.form`
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-top: 20px;
`;

const FormGroup = styled.div`
  display: flex;
  flex-direction: column;
`;

const Label = styled.label`
  margin-bottom: 5px;
  font-weight: 500;
  color: ${props => props.theme.textColor};
`;

const Input = styled.input`
  padding: 8px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  font-size: 14px;
  
  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }
`;

const Checkbox = styled.input`
  margin-right: 8px;
`;

const CheckboxLabel = styled.label`
  display: flex;
  align-items: center;
  color: ${props => props.theme.textColor};
`;

const ErrorMessage = styled.div`
  color: ${props => props.theme.dangerColor};
  margin-top: 10px;
  font-size: 14px;
`;

const SuccessMessage = styled.div`
  color: ${props => props.theme.successColor};
  margin-top: 10px;
  font-size: 14px;
`;

const Modal = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
`;

const ModalContent = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 20px;
  border-radius: 8px;
  width: 400px;
  max-width: 90%;
`;

const ModalHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
`;

const ModalTitle = styled.h3`
  margin: 0;
  color: ${props => props.theme.textColor};
`;

const CloseButton = styled.button`
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: ${props => props.theme.textColor};
`;

const ModalFooter = styled.div`
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
  gap: 10px;
`;

/**
 * Компонент управления кассирами
 * @returns {React.ReactNode} - Компонент управления кассирами
 */
const CashierManagement = () => {
  const theme = getTheme('light');
  const [cashiers, setCashiers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  
  // Состояние для модального окна создания кассира
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newCashier, setNewCashier] = useState({
    username: '',
    password: '',
  });
  
  // Состояние для модального окна редактирования кассира
  const [showEditModal, setShowEditModal] = useState(false);
  const [editCashier, setEditCashier] = useState({
    id: '',
    username: '',
    password: '',
    is_active: true,
  });
  
  // Загрузка списка кассиров при монтировании компонента
  useEffect(() => {
    fetchCashiers();
  }, []);
  
  // Функция для загрузки списка кассиров
  const fetchCashiers = async () => {
    try {
      setLoading(true);
      const response = await AuthService.getCashiers();
      setCashiers(response.cashiers);
      setError('');
    } catch (err) {
      setError('Ошибка при загрузке списка кассиров');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  // Обработчик изменения полей формы создания кассира
  const handleCreateChange = (e) => {
    const { name, value } = e.target;
    setNewCashier(prev => ({
      ...prev,
      [name]: value,
    }));
  };
  
  // Обработчик изменения полей формы редактирования кассира
  const handleEditChange = (e) => {
    const { name, value, type, checked } = e.target;
    setEditCashier(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));
  };
  
  // Обработчик отправки формы создания кассира
  const handleCreateSubmit = async (e) => {
    e.preventDefault();
    
    if (!newCashier.username || !newCashier.password) {
      setError('Пожалуйста, заполните все поля');
      return;
    }
    
    try {
      setLoading(true);
      await AuthService.createCashier(newCashier.username, newCashier.password);
      setSuccess('Кассир успешно создан');
      setNewCashier({ username: '', password: '' });
      setShowCreateModal(false);
      fetchCashiers();
    } catch (err) {
      if (err.response && err.response.data && err.response.data.error) {
        setError(err.response.data.error);
      } else {
        setError('Ошибка при создании кассира');
      }
    } finally {
      setLoading(false);
    }
  };
  
  // Обработчик отправки формы редактирования кассира
  const handleEditSubmit = async (e) => {
    e.preventDefault();
    
    if (!editCashier.username) {
      setError('Имя пользователя не может быть пустым');
      return;
    }
    
    try {
      setLoading(true);
      await AuthService.updateCashier(editCashier.id, {
        username: editCashier.username,
        password: editCashier.password, // Если пароль пустой, он не будет изменен на бэкенде
        is_active: editCashier.is_active,
      });
      setSuccess('Кассир успешно обновлен');
      setShowEditModal(false);
      fetchCashiers();
    } catch (err) {
      if (err.response && err.response.data && err.response.data.error) {
        setError(err.response.data.error);
      } else {
        setError('Ошибка при обновлении кассира');
      }
    } finally {
      setLoading(false);
    }
  };
  
  // Обработчик удаления кассира
  const handleDelete = async (id) => {
    if (!window.confirm('Вы уверены, что хотите удалить этого кассира?')) {
      return;
    }
    
    try {
      setLoading(true);
      await AuthService.deleteCashier(id);
      setSuccess('Кассир успешно удален');
      fetchCashiers();
    } catch (err) {
      if (err.response && err.response.data && err.response.data.error) {
        setError(err.response.data.error);
      } else {
        setError('Ошибка при удалении кассира');
      }
    } finally {
      setLoading(false);
    }
  };
  
  // Открытие модального окна редактирования кассира
  const openEditModal = (cashier) => {
    setEditCashier({
      id: cashier.id,
      username: cashier.username,
      password: '',
      is_active: cashier.is_active,
    });
    setShowEditModal(true);
  };
  
  return (
    <Container>
      <Card theme={theme}>
        <Title theme={theme}>Управление кассирами</Title>
        <Button 
          onClick={() => setShowCreateModal(true)} 
          disabled={loading}
          theme={theme}
        >
          Создать кассира
        </Button>
        
        {loading && <p>Загрузка...</p>}
        {error && <ErrorMessage theme={theme}>{error}</ErrorMessage>}
        {success && <SuccessMessage theme={theme}>{success}</SuccessMessage>}
        
        {!loading && cashiers.length === 0 ? (
          <p>Нет доступных кассиров</p>
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
              {cashiers.map(cashier => (
                <tr key={cashier.id}>
                  <Td theme={theme}>{cashier.username}</Td>
                  <Td theme={theme}>
                    {cashier.is_active ? 'Активен' : 'Неактивен'}
                  </Td>
                  <Td theme={theme}>
                    {cashier.last_login 
                      ? new Date(cashier.last_login).toLocaleString() 
                      : 'Никогда'}
                  </Td>
                  <Td theme={theme}>
                    <ActionButton 
                      onClick={() => openEditModal(cashier)}
                      theme={theme}
                    >
                      Редактировать
                    </ActionButton>
                    <ActionButton 
                      onClick={() => handleDelete(cashier.id)}
                      danger
                      theme={theme}
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
      
      {/* Модальное окно создания кассира */}
      {showCreateModal && (
        <Modal>
          <ModalContent theme={theme}>
            <ModalHeader>
              <ModalTitle theme={theme}>Создать кассира</ModalTitle>
              <CloseButton 
                onClick={() => setShowCreateModal(false)}
                theme={theme}
              >
                &times;
              </CloseButton>
            </ModalHeader>
            <Form onSubmit={handleCreateSubmit}>
              <FormGroup>
                <Label theme={theme}>Имя пользователя</Label>
                <Input
                  type="text"
                  name="username"
                  value={newCashier.username}
                  onChange={handleCreateChange}
                  theme={theme}
                />
              </FormGroup>
              <FormGroup>
                <Label theme={theme}>Пароль</Label>
                <Input
                  type="password"
                  name="password"
                  value={newCashier.password}
                  onChange={handleCreateChange}
                  theme={theme}
                />
              </FormGroup>
              {error && <ErrorMessage theme={theme}>{error}</ErrorMessage>}
              <ModalFooter>
                <Button 
                  type="button" 
                  onClick={() => setShowCreateModal(false)}
                  theme={theme}
                >
                  Отмена
                </Button>
                <Button 
                  type="submit" 
                  disabled={loading}
                  theme={theme}
                >
                  {loading ? 'Создание...' : 'Создать'}
                </Button>
              </ModalFooter>
            </Form>
          </ModalContent>
        </Modal>
      )}
      
      {/* Модальное окно редактирования кассира */}
      {showEditModal && (
        <Modal>
          <ModalContent theme={theme}>
            <ModalHeader>
              <ModalTitle theme={theme}>Редактировать кассира</ModalTitle>
              <CloseButton 
                onClick={() => setShowEditModal(false)}
                theme={theme}
              >
                &times;
              </CloseButton>
            </ModalHeader>
            <Form onSubmit={handleEditSubmit}>
              <FormGroup>
                <Label theme={theme}>Имя пользователя</Label>
                <Input
                  type="text"
                  name="username"
                  value={editCashier.username}
                  onChange={handleEditChange}
                  theme={theme}
                />
              </FormGroup>
              <FormGroup>
                <Label theme={theme}>Новый пароль (оставьте пустым, чтобы не менять)</Label>
                <Input
                  type="password"
                  name="password"
                  value={editCashier.password}
                  onChange={handleEditChange}
                  theme={theme}
                />
              </FormGroup>
              <FormGroup>
                <CheckboxLabel theme={theme}>
                  <Checkbox
                    type="checkbox"
                    name="is_active"
                    checked={editCashier.is_active}
                    onChange={handleEditChange}
                  />
                  Активен
                </CheckboxLabel>
              </FormGroup>
              {error && <ErrorMessage theme={theme}>{error}</ErrorMessage>}
              <ModalFooter>
                <Button 
                  type="button" 
                  onClick={() => setShowEditModal(false)}
                  theme={theme}
                >
                  Отмена
                </Button>
                <Button 
                  type="submit" 
                  disabled={loading}
                  theme={theme}
                >
                  {loading ? 'Сохранение...' : 'Сохранить'}
                </Button>
              </ModalFooter>
            </Form>
          </ModalContent>
        </Modal>
      )}
    </Container>
  );
};

export default CashierManagement;
