import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useNavigate } from 'react-router-dom';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';

const Container = styled.div`
  padding: 20px;
`;

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
`;

const Title = styled.h2`
  margin: 0;
  color: ${props => props.theme.textColor};
`;

const Button = styled.button`
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  
  &:hover {
    background-color: ${props => props.theme.primaryColorDark};
  }
  
  &:disabled {
    background-color: #ccc;
    cursor: not-allowed;
  }
`;

const Filters = styled.div`
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  flex-wrap: wrap;
`;

const FilterSelect = styled.select`
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 5px;
  background-color: white;
  font-size: 14px;
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
`;

const Th = styled.th`
  background-color: ${props => props.theme.primaryColor};
  color: white;
  padding: 12px;
  text-align: left;
  font-weight: 500;
`;

const Td = styled.td`
  padding: 12px;
  border-bottom: 1px solid #eee;
`;

const StatusBadge = styled.span`
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  text-transform: uppercase;
  
  &.free {
    background-color: #e8f5e8;
    color: #2d5a2d;
  }
  
  &.busy {
    background-color: #ffe8e8;
    color: #5a2d2d;
  }
  
  &.reserved {
    background-color: #fff3e8;
    color: #5a4a2d;
  }
  
  &.maintenance {
    background-color: #e8f0ff;
    color: #2d4a5a;
  }
`;

const ServiceTypeBadge = styled.span`
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  
  &.wash {
    background-color: #e3f2fd;
    color: #1565c0;
  }
  
  &.air_dry {
    background-color: #f3e5f5;
    color: #7b1fa2;
  }
  
  &.vacuum {
    background-color: #e8f5e8;
    color: #2e7d32;
  }
`;

const ActionButton = styled.button`
  background: none;
  border: none;
  color: ${props => props.theme.primaryColor};
  cursor: pointer;
  margin-right: 10px;
  font-size: 14px;
  
  &:hover {
    text-decoration: underline;
  }
  
  &.delete {
    color: #d32f2f;
  }
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
  background-color: white;
  padding: 30px;
  border-radius: 8px;
  width: 400px;
  max-width: 90%;
`;

const ModalTitle = styled.h3`
  margin: 0 0 20px 0;
  color: ${props => props.theme.textColor};
`;

const Form = styled.form`
  display: flex;
  flex-direction: column;
  gap: 15px;
`;

const FormGroup = styled.div`
  display: flex;
  flex-direction: column;
  gap: 5px;
`;

const Label = styled.label`
  font-weight: 500;
  color: ${props => props.theme.textColor};
`;

const Input = styled.input`
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-size: 14px;
`;

const Select = styled.select`
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-size: 14px;
`;

const ButtonGroup = styled.div`
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 20px;
`;

const ErrorMessage = styled.div`
  color: #d32f2f;
  margin-bottom: 15px;
  font-size: 14px;
`;

const SuccessMessage = styled.div`
  color: #2e7d32;
  margin-bottom: 15px;
  font-size: 14px;
`;

const LoadingMessage = styled.div`
  color: ${props => props.theme.textColor};
  text-align: center;
  padding: 20px;
  font-size: 14px;
`;

const WashBoxManagement = () => {
  const theme = getTheme('light');
  const [washBoxes, setWashBoxes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [editingWashBox, setEditingWashBox] = useState(null);
  const [filters, setFilters] = useState({
    status: '',
    serviceType: ''
  });
  const [formData, setFormData] = useState({
    number: '',
    status: 'free',
    serviceType: 'wash'
  });
  const navigate = useNavigate();

  // Загрузка боксов
  const fetchWashBoxes = async () => {
    try {
      setLoading(true);
      setError('');
      
      const response = await ApiService.getWashBoxes();
      setWashBoxes(response.wash_boxes || []);
    } catch (err) {
      setError('Ошибка при загрузке боксов');
      console.error('Error fetching wash boxes:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWashBoxes();
  }, [filters]);

  // Создание бокса
  const handleCreate = async (e) => {
    e.preventDefault();
    
    if (!formData.number) {
      setError('Номер бокса обязателен');
      return;
    }
    
    try {
      setLoading(true);
      setError('');
      
      // Исправляем названия полей для JSON
      const washBoxData = {
        number: parseInt(formData.number, 10),
        status: formData.status,
        service_type: formData.serviceType
      };
      
      await ApiService.createWashBox(washBoxData);
      setSuccess('Бокс успешно создан');
      setShowCreateModal(false);
      setFormData({ number: '', status: 'free', serviceType: 'wash' });
      fetchWashBoxes();
    } catch (err) {
      if (err.response?.data?.error) {
        setError(err.response.data.error);
      } else {
        setError('Ошибка при создании бокса');
      }
    } finally {
      setLoading(false);
    }
  };

  // Обновление бокса
  const handleUpdate = async (e) => {
    e.preventDefault();
    
    try {
      setLoading(true);
      setError('');
      
      // Исправляем названия полей для JSON
      const washBoxData = {
        number: parseInt(formData.number, 10),
        status: formData.status,
        service_type: formData.serviceType
      };
      
      await ApiService.updateWashBox(editingWashBox.id, washBoxData);
      setSuccess('Бокс успешно обновлен');
      setShowEditModal(false);
      setEditingWashBox(null);
      setFormData({ number: '', status: 'free', serviceType: 'wash' });
      fetchWashBoxes();
    } catch (err) {
      if (err.response?.data?.error) {
        setError(err.response.data.error);
      } else {
        setError('Ошибка при обновлении бокса');
      }
    } finally {
      setLoading(false);
    }
  };

  // Удаление бокса
  const handleDelete = async (id) => {
    if (!window.confirm('Вы уверены, что хотите удалить этот бокс?')) {
      return;
    }
    
    try {
      setLoading(true);
      setError('');
      
      await ApiService.deleteWashBox(id);
      setSuccess('Бокс успешно удален');
      fetchWashBoxes();
    } catch (err) {
      if (err.response?.data?.error) {
        setError(err.response.data.error);
      } else {
        setError('Ошибка при удалении бокса');
      }
    } finally {
      setLoading(false);
    }
  };

  // Открытие модального окна редактирования
  const openEditModal = (washBox) => {
    setEditingWashBox(washBox);
    setFormData({
      number: washBox.number.toString(),
      status: washBox.status,
      serviceType: washBox.service_type
    });
    setShowEditModal(true);
  };

  // Закрытие модальных окон
  const closeModals = () => {
    setShowCreateModal(false);
    setShowEditModal(false);
    setEditingWashBox(null);
    setFormData({ number: '', status: 'free', serviceType: 'wash' });
    setError('');
  };

  const getStatusText = (status) => {
    const statusMap = {
      free: 'Свободен',
      busy: 'Занят',
      reserved: 'Зарезервирован',
      maintenance: 'Обслуживание'
    };
    return statusMap[status] || status;
  };

  const getServiceTypeText = (serviceType) => {
    const serviceMap = {
      wash: 'Мойка',
      air_dry: 'Обдув',
      vacuum: 'Пылесос'
    };
    return serviceMap[serviceType] || serviceType;
  };

  if (loading && washBoxes.length === 0) {
    return <LoadingMessage theme={theme}>Загрузка боксов...</LoadingMessage>;
  }

  return (
    <Container>
      <Header>
        <Title theme={theme}>Управление боксами мойки</Title>
        <Button theme={theme} onClick={() => setShowCreateModal(true)}>
          Добавить бокс
        </Button>
      </Header>

      {error && <ErrorMessage>{error}</ErrorMessage>}
      {success && <SuccessMessage>{success}</SuccessMessage>}

      <Filters>
        <FilterSelect
          value={filters.status}
          onChange={(e) => setFilters({ ...filters, status: e.target.value })}
        >
          <option value="">Все статусы</option>
          <option value="free">Свободен</option>
          <option value="busy">Занят</option>
          <option value="reserved">Зарезервирован</option>
          <option value="maintenance">Обслуживание</option>
        </FilterSelect>

        <FilterSelect
          value={filters.serviceType}
          onChange={(e) => setFilters({ ...filters, serviceType: e.target.value })}
        >
          <option value="">Все типы услуг</option>
          <option value="wash">Мойка</option>
          <option value="air_dry">Обдув</option>
          <option value="vacuum">Пылесос</option>
        </FilterSelect>
      </Filters>

      <Table theme={theme}>
        <thead>
          <tr>
            <Th theme={theme}>ID</Th>
            <Th theme={theme}>Номер</Th>
            <Th theme={theme}>Статус</Th>
            <Th theme={theme}>Тип услуги</Th>
            <Th theme={theme}>Дата создания</Th>
            <Th theme={theme}>Действия</Th>
          </tr>
        </thead>
        <tbody>
          {washBoxes.map((washBox) => (
            <tr key={washBox.id}>
              <Td>{washBox.id}</Td>
              <Td>{washBox.number}</Td>
              <Td>
                <StatusBadge className={washBox.status}>
                  {getStatusText(washBox.status)}
                </StatusBadge>
              </Td>
              <Td>
                <ServiceTypeBadge className={washBox.service_type}>
                  {getServiceTypeText(washBox.service_type)}
                </ServiceTypeBadge>
              </Td>
              <Td>{new Date(washBox.created_at).toLocaleDateString('ru-RU')}</Td>
              <Td>
                <ActionButton theme={theme} onClick={() => openEditModal(washBox)}>
                  Редактировать
                </ActionButton>
                <ActionButton 
                  theme={theme} 
                  className="delete"
                  onClick={() => handleDelete(washBox.id)}
                >
                  Удалить
                </ActionButton>
                <ActionButton 
                  theme={theme}
                  onClick={() => {
                    navigate('/admin/sessions', { 
                      state: { 
                        filters: { boxNumber: washBox.number },
                        showBoxFilter: true 
                      } 
                    });
                  }}
                >
                  Сессии
                </ActionButton>
              </Td>
            </tr>
          ))}
        </tbody>
      </Table>

      {/* Модальное окно создания */}
      {showCreateModal && (
        <Modal>
          <ModalContent>
            <ModalTitle theme={theme}>Создать новый бокс</ModalTitle>
            <Form onSubmit={handleCreate}>
              <FormGroup>
                <Label theme={theme}>Номер бокса</Label>
                <Input
                  type="number"
                  value={formData.number}
                  onChange={(e) => setFormData({ ...formData, number: e.target.value })}
                  required
                />
              </FormGroup>
              
              <FormGroup>
                <Label theme={theme}>Статус</Label>
                <Select
                  value={formData.status}
                  onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                >
                  <option value="free">Свободен</option>
                  <option value="busy">Занят</option>
                  <option value="reserved">Зарезервирован</option>
                  <option value="maintenance">Обслуживание</option>
                </Select>
              </FormGroup>
              
              <FormGroup>
                <Label theme={theme}>Тип услуги</Label>
                <Select
                  value={formData.serviceType}
                  onChange={(e) => setFormData({ ...formData, serviceType: e.target.value })}
                >
                  <option value="wash">Мойка</option>
                  <option value="air_dry">Обдув</option>
                  <option value="vacuum">Пылесос</option>
                </Select>
              </FormGroup>
              
              <ButtonGroup>
                <Button theme={theme} type="button" onClick={closeModals}>
                  Отмена
                </Button>
                <Button theme={theme} type="submit" disabled={loading}>
                  {loading ? 'Создание...' : 'Создать'}
                </Button>
              </ButtonGroup>
            </Form>
          </ModalContent>
        </Modal>
      )}

      {/* Модальное окно редактирования */}
      {showEditModal && (
        <Modal>
          <ModalContent>
            <ModalTitle theme={theme}>Редактировать бокс</ModalTitle>
            <Form onSubmit={handleUpdate}>
              <FormGroup>
                <Label theme={theme}>Номер бокса</Label>
                <div style={{ 
                  padding: '8px 12px', 
                  border: '1px solid #ddd', 
                  borderRadius: '4px', 
                  backgroundColor: '#f5f5f5',
                  color: '#666'
                }}>
                  {formData.number}
                </div>
              </FormGroup>
              
              <FormGroup>
                <Label theme={theme}>Статус</Label>
                <Select
                  value={formData.status}
                  onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                >
                  <option value="free">Свободен</option>
                  <option value="busy">Занят</option>
                  <option value="reserved">Зарезервирован</option>
                  <option value="maintenance">Обслуживание</option>
                </Select>
              </FormGroup>
              
              <FormGroup>
                <Label theme={theme}>Тип услуги</Label>
                <Select
                  value={formData.serviceType}
                  onChange={(e) => setFormData({ ...formData, serviceType: e.target.value })}
                >
                  <option value="wash">Мойка</option>
                  <option value="air_dry">Обдув</option>
                  <option value="vacuum">Пылесос</option>
                </Select>
              </FormGroup>
              
              <ButtonGroup>
                <Button theme={theme} type="button" onClick={closeModals}>
                  Отмена
                </Button>
                <Button theme={theme} type="submit" disabled={loading}>
                  {loading ? 'Сохранение...' : 'Сохранить'}
                </Button>
              </ButtonGroup>
            </Form>
          </ModalContent>
        </Modal>
      )}
    </Container>
  );
};

export default WashBoxManagement; 