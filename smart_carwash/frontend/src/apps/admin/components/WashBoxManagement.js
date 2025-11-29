import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';
import AuthService from '../../../shared/services/AuthService';
import MobileTable from '../../../shared/components/MobileTable';
import usePolling from '../../../shared/hooks/usePolling';
import axios from 'axios';

// API клиент для Modbus тестирования
const modbusApi = axios.create({
  baseURL: process.env.REACT_APP_API_URL ? process.env.REACT_APP_API_URL.replace('/v1', '') : '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Добавляем токен авторизации
modbusApi.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  return config;
});

// Добавляем перехватчик для обработки ошибок авторизации
modbusApi.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      // Если сервер вернул 401, значит токен истек или недействителен
      // Удаляем токен и перенаправляем на страницу входа
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('expiresAt');
      
      // Определяем, какой тип пользователя был авторизован
      const isAdmin = localStorage.getItem('isAdmin') === 'true';
      localStorage.removeItem('isAdmin');
      
      // Перенаправляем на соответствующую страницу входа
      if (isAdmin) {
        window.location.href = '/admin/login';
      } else {
        window.location.href = '/cashier/login';
      }
    }
    return Promise.reject(error);
  }
);

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

const HighlightedRow = styled.tr`
  background-color: rgba(0, 123, 255, 0.1) !important;
  border-left: 4px solid ${props => props.theme.primaryColor};
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

const CoilStatusIndicator = styled.span`
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  margin-left: 8px;
  background-color: ${props => {
    if (props.status === null || props.status === undefined) return '#f5f5f5';
    return props.status ? '#E8F5E8' : '#FFEBEE';
  }};
  color: ${props => {
    if (props.status === null || props.status === undefined) return '#999';
    return props.status ? '#2E7D32' : '#C62828';
  }};
  border: 1px solid ${props => {
    if (props.status === null || props.status === undefined) return '#e0e0e0';
    return props.status ? '#81C784' : '#EF9A9A';
  }};
`;

const ChemistryBadge = styled.span`
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  
  &.enabled {
    background-color: #e8f5e8;
    color: #2d5a2d;
  }
  
  &.disabled {
    background-color: #ffe8e8;
    color: #5a2d2d;
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
  max-height: 90vh;
  overflow-y: auto;
`;

const ModalContentWide = styled.div`
  background-color: white;
  padding: 30px;
  border-radius: 8px;
  width: 800px;
  max-width: 95%;
  max-height: 90vh;
  overflow-y: auto;
`;

const TwoColumnGrid = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  
  @media (max-width: 768px) {
    grid-template-columns: 1fr;
  }
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

const FullWidthFormGroup = styled(FormGroup)`
  grid-column: 1 / -1;
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

const TestButtonGroup = styled.div`
  display: flex;
  gap: 8px;
  margin-top: 8px;
`;

const TestButton = styled.button`
  background: #FF9800;
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  
  &:hover {
    opacity: 0.9;
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  &.success {
    background: #4CAF50;
  }
  
  &.error {
    background: #F44336;
  }
`;

const TestResult = styled.div`
  margin-top: 8px;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 12px;
  
  &.success {
    background: #E8F5E8;
    color: #2E7D32;
  }
  
  &.error {
    background: #FFEBEE;
    color: #C62828;
  }
  
  &.testing {
    background: #E3F2FD;
    color: #1565C0;
  }
`;

const ControlButtonsGroup = styled.div`
  display: flex;
  gap: 4px;
  align-items: center;
`;

const ControlButton = styled.button`
  background: ${props => props.$isOn ? '#4CAF50' : '#F44336'};
  color: white;
  border: none;
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
  font-weight: 500;
  min-width: 45px;
  transition: opacity 0.2s;
  
  &:hover:not(:disabled) {
    opacity: 0.85;
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

const ControlStatus = styled.div`
  font-size: 10px;
  margin-top: 2px;
  
  &.success {
    color: #2E7D32;
  }
  
  &.error {
    color: #C62828;
  }
  
  &.testing {
    color: #1565C0;
  }
`;

const LoadingMessage = styled.div`
  color: ${props => props.theme.textColor};
  text-align: center;
  padding: 20px;
  font-size: 14px;
`;

const BoxNumber = styled.span`
  font-size: 24px;
  font-weight: bold;
  color: ${props => props.theme.textColor};
`;

const MobileBoxNumber = styled.span`
  font-size: 28px;
  font-weight: bold;
  color: ${props => props.theme.textColor};
`;

const UUIDContainer = styled.div`
  display: flex;
  gap: 10px;
  align-items: center;
  margin-bottom: 15px;
  padding: 12px;
  background-color: #f8f9fa;
  border-radius: 6px;
`;

const UUIDField = styled.div`
  flex: 1;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  padding: 8px;
  background-color: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  color: #666;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const CopyButton = styled.button`
  background-color: #6c757d;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  white-space: nowrap;
  
  &:hover {
    background-color: #5a6268;
  }
`;

const CommentCell = styled.div`
  max-width: 250px;
  word-wrap: break-word;
  white-space: pre-wrap;
  line-height: 1.4;
  font-size: 13px;
`;

const CharCounter = styled.small`
  color: ${props => props.$isNearLimit ? '#d32f2f' : '#666'};
  font-size: 12px;
  margin-top: 4px;
  display: block;
`;

const Textarea = styled.textarea`
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  width: 100%;
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
    serviceType: 'wash',
    chemistryEnabled: true,
    priority: 'A',
    lightCoilRegister: '',
    chemistryCoilRegister: '',
    comment: ''
  });
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const highlightedBoxNumber = searchParams.get('highlight');
  const isLimitedAdmin = AuthService.isLimitedAdmin();
  
  // Состояния для тестирования Modbus
  const [testingBox, setTestingBox] = useState(null);
  const [testResults, setTestResults] = useState({});
  
  // Состояния для управления регистрами в таблице
  const [controlOperations, setControlOperations] = useState({});
  const [controlResults, setControlResults] = useState({});

  // Загрузка боксов
  const fetchWashBoxes = async () => {
    try {
      setLoading(true);
      setError('');
      
      const response = await ApiService.getWashBoxes(filters);
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

  // Поллинг для автоматического обновления данных каждые 3 секунды
  usePolling(fetchWashBoxes, 3000, true, [filters]);

  // Автоматическое открытие модального окна при переходе с выделенным боксом
  useEffect(() => {
    if (highlightedBoxNumber && washBoxes.length > 0) {
      const washBox = washBoxes.find(box => box.number.toString() === highlightedBoxNumber);
      if (washBox) {
        openEditModal(washBox);
        // Убираем параметр из URL
        navigate('/admin/washboxes', { replace: true });
      }
    }
  }, [highlightedBoxNumber, washBoxes]);

  // Функции тестирования Modbus
  const testCoil = async (boxId, register, value) => {
    try {
      setTestingBox(boxId);
      setTestResults(prev => ({ 
        ...prev, 
        [`${boxId}_${register}`]: { 
          status: 'testing', 
          message: `Тестирование ${register}...` 
        } 
      }));
      
      const response = await modbusApi.post('/admin/modbus/test-coil', { 
        box_id: boxId, 
        register, 
        value 
      });
      
      setTestResults(prev => ({ 
        ...prev, 
        [`${boxId}_${register}`]: { 
          status: response.data.success ? 'success' : 'error', 
          message: response.data.message 
        } 
      }));
    } catch (err) {
      setTestResults(prev => ({ 
        ...prev, 
        [`${boxId}_${register}`]: { 
          status: 'error', 
          message: err.response?.data?.error || err.message 
        } 
      }));
    } finally {
      setTestingBox(null);
    }
  };

  // Функция копирования UUID
  const handleCopyUUID = (uuid) => {
    navigator.clipboard.writeText(uuid);
    setSuccess('UUID скопирован в буфер обмена');
    setTimeout(() => setSuccess(''), 2000);
  };

  // Функция управления регистром из таблицы
  const handleControlRegister = async (boxId, register, value, type) => {
    const key = `${boxId}_${type}`;
    
    try {
      setControlOperations(prev => ({ ...prev, [key]: true }));
      setControlResults(prev => ({ 
        ...prev, 
        [key]: { 
          status: 'testing', 
          message: value ? 'Включение...' : 'Выключение...' 
        } 
      }));
      
      const response = await modbusApi.post('/admin/modbus/test-coil', { 
        box_id: boxId, 
        register, 
        value 
      });
      
      setControlResults(prev => ({ 
        ...prev, 
        [key]: { 
          status: response.data.success ? 'success' : 'error', 
          message: response.data.success 
            ? (value ? '✓ Включено' : '✓ Выключено') 
            : response.data.message 
        } 
      }));
      
      // Если операция успешна, обновляем список боксов для получения актуальных статусов
      if (response.data.success) {
        fetchWashBoxes();
        
        // Автоматически скрываем сообщение об успехе через 3 секунды
        setTimeout(() => {
          setControlResults(prev => {
            const newResults = { ...prev };
            delete newResults[key];
            return newResults;
          });
        }, 3000);
      }
    } catch (err) {
      setControlResults(prev => ({ 
        ...prev, 
        [key]: { 
          status: 'error', 
          message: err.response?.data?.error || err.message 
        } 
      }));
    } finally {
      setControlOperations(prev => ({ ...prev, [key]: false }));
    }
  };

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
        service_type: formData.serviceType,
        chemistry_enabled: formData.chemistryEnabled,
        priority: formData.priority.toUpperCase() || 'A',
        light_coil_register: formData.lightCoilRegister || null,
        chemistry_coil_register: formData.chemistryCoilRegister || null,
        comment: formData.comment || null
      };
      
      await ApiService.createWashBox(washBoxData);
      setSuccess('Бокс успешно создан');
      setShowCreateModal(false);
      setFormData({ number: '', status: 'free', serviceType: 'wash', chemistryEnabled: true, priority: 'A', lightCoilRegister: '', chemistryCoilRegister: '', comment: '' });
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
        service_type: formData.serviceType,
        chemistry_enabled: formData.chemistryEnabled,
        priority: formData.priority.toUpperCase() || 'A',
        light_coil_register: formData.lightCoilRegister || null,
        chemistry_coil_register: formData.chemistryCoilRegister || null,
        comment: formData.comment || null
      };
      
      await ApiService.updateWashBox(editingWashBox.id, washBoxData);
      setSuccess('Бокс успешно обновлен');
      setShowEditModal(false);
      setEditingWashBox(null);
      setFormData({ number: '', status: 'free', serviceType: 'wash', chemistryEnabled: true, priority: 'A', lightCoilRegister: '', chemistryCoilRegister: '', comment: '' });
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
      serviceType: washBox.service_type,
      chemistryEnabled: washBox.chemistry_enabled,
      priority: washBox.priority || 'A',
      lightCoilRegister: washBox.light_coil_register || '',
      chemistryCoilRegister: washBox.chemistry_coil_register || '',
      comment: washBox.comment || ''
    });
    setShowEditModal(true);
  };

  // Закрытие модальных окон
  const closeModals = () => {
    setShowCreateModal(false);
    setShowEditModal(false);
    setEditingWashBox(null);
    setFormData({ number: '', status: 'free', serviceType: 'wash', chemistryEnabled: true, priority: 'A', lightCoilRegister: '', chemistryCoilRegister: '', comment: '' });
    setError('');
  };

  const getStatusText = (status) => {
    const statusMap = {
      free: 'Свободен',
      busy: 'Занят',
      reserved: 'Зарезервирован',
      maintenance: 'Сервис'
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
        <div style={{ display: 'flex', gap: '10px' }}>
          {!isLimitedAdmin && (
          <Button theme={theme} onClick={() => setShowCreateModal(true)}>
            Добавить бокс
          </Button>
          )}
        </div>
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
          <option value="maintenance">Сервис</option>
          <option value="cleaning">На уборке</option>
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

      {/* Десктопная таблица */}
      <Table theme={theme}>
        <thead>
          <tr>
            <Th theme={theme}>Номер</Th>
            <Th theme={theme}>Статус</Th>
            <Th theme={theme}>Тип услуги</Th>
            <Th theme={theme}>Приоритет</Th>
            <Th theme={theme}>Химия</Th>
            <Th theme={theme}>Регистр света</Th>
            <Th theme={theme}>Регистр химии</Th>
            <Th theme={theme}>Комментарий</Th>
            <Th theme={theme}>Дата создания</Th>
            <Th theme={theme}>Действия</Th>
          </tr>
        </thead>
        <tbody>
          {washBoxes.map((washBox) => {
            const isHighlighted = washBox.number.toString() === highlightedBoxNumber;
            const RowComponent = isHighlighted ? HighlightedRow : 'tr';
            
            return (
              <RowComponent key={washBox.id} theme={theme}>
                <Td>
                  <BoxNumber theme={theme}>{washBox.number}</BoxNumber>
                </Td>
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
                <Td>
                  <span style={{ 
                    fontWeight: 'bold', 
                    fontSize: '16px',
                    color: washBox.priority === 'A' ? '#28a745' : ['B', 'C'].includes(washBox.priority) ? '#ffc107' : '#dc3545' 
                  }}>
                    {washBox.priority || 'A'}
                  </span>
                </Td>
                <Td>
                  {washBox.service_type === 'wash' ? (
                    <ChemistryBadge className={washBox.chemistry_enabled ? 'enabled' : 'disabled'}>
                      {washBox.chemistry_enabled ? 'Включена' : 'Отключена'}
                    </ChemistryBadge>
                  ) : (
                    <span style={{ color: '#999' }}>Недоступна</span>
                  )}
                </Td>
                <Td>
                  {washBox.light_coil_register ? (
                    <div>
                      <div style={{ marginBottom: '6px', display: 'flex', alignItems: 'center', flexWrap: 'wrap' }}>
                        <span>{washBox.light_coil_register}</span>
                        <CoilStatusIndicator status={washBox.light_status}>
                          {washBox.light_status === null || washBox.light_status === undefined ? '?' : washBox.light_status ? '💡 ВКЛ' : '💡 ВЫКЛ'}
                        </CoilStatusIndicator>
                      </div>
                      {(!isLimitedAdmin || (isLimitedAdmin && washBox.status === 'maintenance')) && (
                      <ControlButtonsGroup>
                        <ControlButton
                          $isOn={true}
                          onClick={() => handleControlRegister(washBox.id, washBox.light_coil_register, true, 'light')}
                          disabled={controlOperations[`${washBox.id}_light`]}
                        >
                          ВКЛ
                        </ControlButton>
                        <ControlButton
                          $isOn={false}
                          onClick={() => handleControlRegister(washBox.id, washBox.light_coil_register, false, 'light')}
                          disabled={controlOperations[`${washBox.id}_light`]}
                        >
                          ВЫКЛ
                        </ControlButton>
                      </ControlButtonsGroup>
                      )}
                      {controlResults[`${washBox.id}_light`] && (
                        <ControlStatus className={controlResults[`${washBox.id}_light`].status}>
                          {controlResults[`${washBox.id}_light`].message}
                        </ControlStatus>
                      )}
                    </div>
                  ) : (
                    <span style={{ color: '#999' }}>Не задан</span>
                  )}
                </Td>
                <Td>
                  {washBox.service_type === 'wash' ? (
                    washBox.chemistry_coil_register ? (
                      <div>
                        <div style={{ marginBottom: '6px', display: 'flex', alignItems: 'center', flexWrap: 'wrap' }}>
                          <span>{washBox.chemistry_coil_register}</span>
                          <CoilStatusIndicator status={washBox.chemistry_status}>
                            {washBox.chemistry_status === null || washBox.chemistry_status === undefined ? '?' : washBox.chemistry_status ? '🧪 ВКЛ' : '🧪 ВЫКЛ'}
                          </CoilStatusIndicator>
                        </div>
                        {(!isLimitedAdmin || (isLimitedAdmin && washBox.status === 'maintenance')) && (
                        <ControlButtonsGroup>
                          <ControlButton
                            $isOn={true}
                            onClick={() => handleControlRegister(washBox.id, washBox.chemistry_coil_register, true, 'chemistry')}
                            disabled={controlOperations[`${washBox.id}_chemistry`]}
                          >
                            ВКЛ
                          </ControlButton>
                          <ControlButton
                            $isOn={false}
                            onClick={() => handleControlRegister(washBox.id, washBox.chemistry_coil_register, false, 'chemistry')}
                            disabled={controlOperations[`${washBox.id}_chemistry`]}
                          >
                            ВЫКЛ
                          </ControlButton>
                        </ControlButtonsGroup>
                        )}
                        {controlResults[`${washBox.id}_chemistry`] && (
                          <ControlStatus className={controlResults[`${washBox.id}_chemistry`].status}>
                            {controlResults[`${washBox.id}_chemistry`].message}
                          </ControlStatus>
                        )}
                      </div>
                    ) : (
                      <span style={{ color: '#999' }}>Не задан</span>
                    )
                  ) : (
                    <span style={{ color: '#999' }}>Недоступна</span>
                  )}
                </Td>
                <Td>
                  {washBox.comment ? (
                    <CommentCell>{washBox.comment}</CommentCell>
                  ) : (
                    <span style={{ color: '#999' }}>—</span>
                  )}
                </Td>
                <Td>{new Date(washBox.created_at).toLocaleDateString('ru-RU')}</Td>
                <Td>
                  <ActionButton theme={theme} onClick={() => openEditModal(washBox)}>
                    Редактировать
                  </ActionButton>
                  {!isLimitedAdmin && (
                    <ActionButton 
                      theme={theme} 
                      className="delete"
                      onClick={() => handleDelete(washBox.id)}
                    >
                      Удалить
                    </ActionButton>
                  )}
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
              </RowComponent>
            );
          })}
        </tbody>
      </Table>

      {/* Мобильные карточки */}
      <MobileTable
        data={washBoxes}
        columns={[
          { key: 'number', label: 'Номер', accessor: (item) => (
            <MobileBoxNumber theme={theme}>{item.number}</MobileBoxNumber>
          )},
          { key: 'status', label: 'Статус', accessor: (item) => (
            <StatusBadge className={item.status}>{getStatusText(item.status)}</StatusBadge>
          )},
          { key: 'service_type', label: 'Тип услуги', accessor: (item) => (
            <ServiceTypeBadge className={item.service_type}>{getServiceTypeText(item.service_type)}</ServiceTypeBadge>
          )},
          { key: 'priority', label: 'Приоритет', accessor: (item) => (
            <span style={{ 
              fontWeight: 'bold',
              fontSize: '16px',
              color: item.priority === 'A' ? '#28a745' : ['B', 'C'].includes(item.priority) ? '#ffc107' : '#dc3545' 
            }}>
              {item.priority || 'A'}
            </span>
          )},
          { key: 'chemistry', label: 'Химия', accessor: (item) => (
            item.service_type === 'wash' ? (
              <ChemistryBadge className={item.chemistry_enabled ? 'enabled' : 'disabled'}>
                {item.chemistry_enabled ? 'Включена' : 'Отключена'}
              </ChemistryBadge>
            ) : (
              <span style={{ color: '#999' }}>Недоступна</span>
            )
          )},
          { key: 'light_register', label: 'Регистр света', accessor: (item) => (
            item.light_coil_register ? (
              <div>
                <div style={{ marginBottom: '6px', fontSize: '12px', display: 'flex', alignItems: 'center', flexWrap: 'wrap' }}>
                  <span>{item.light_coil_register}</span>
                  <CoilStatusIndicator status={item.light_status}>
                    {item.light_status === null || item.light_status === undefined ? '?' : item.light_status ? '💡 ВКЛ' : '💡 ВЫКЛ'}
                  </CoilStatusIndicator>
                </div>
                <ControlButtonsGroup>
                  <ControlButton
                    $isOn={true}
                    onClick={() => handleControlRegister(item.id, item.light_coil_register, true, 'light')}
                    disabled={controlOperations[`${item.id}_light`]}
                  >
                    ВКЛ
                  </ControlButton>
                  <ControlButton
                    $isOn={false}
                    onClick={() => handleControlRegister(item.id, item.light_coil_register, false, 'light')}
                    disabled={controlOperations[`${item.id}_light`]}
                  >
                    ВЫКЛ
                  </ControlButton>
                </ControlButtonsGroup>
                {controlResults[`${item.id}_light`] && (
                  <ControlStatus className={controlResults[`${item.id}_light`].status}>
                    {controlResults[`${item.id}_light`].message}
                  </ControlStatus>
                )}
              </div>
            ) : (
              <span style={{ color: '#999' }}>Не задан</span>
            )
          )},
          { key: 'chemistry_register', label: 'Регистр химии', accessor: (item) => (
            item.service_type === 'wash' ? (
              item.chemistry_coil_register ? (
                <div>
                  <div style={{ marginBottom: '6px', fontSize: '12px', display: 'flex', alignItems: 'center', flexWrap: 'wrap' }}>
                    <span>{item.chemistry_coil_register}</span>
                    <CoilStatusIndicator status={item.chemistry_status}>
                      {item.chemistry_status === null || item.chemistry_status === undefined ? '?' : item.chemistry_status ? '🧪 ВКЛ' : '🧪 ВЫКЛ'}
                    </CoilStatusIndicator>
                  </div>
                  <ControlButtonsGroup>
                    <ControlButton
                      $isOn={true}
                      onClick={() => handleControlRegister(item.id, item.chemistry_coil_register, true, 'chemistry')}
                      disabled={controlOperations[`${item.id}_chemistry`]}
                    >
                      ВКЛ
                    </ControlButton>
                    <ControlButton
                      $isOn={false}
                      onClick={() => handleControlRegister(item.id, item.chemistry_coil_register, false, 'chemistry')}
                      disabled={controlOperations[`${item.id}_chemistry`]}
                    >
                      ВЫКЛ
                    </ControlButton>
                  </ControlButtonsGroup>
                  {controlResults[`${item.id}_chemistry`] && (
                    <ControlStatus className={controlResults[`${item.id}_chemistry`].status}>
                      {controlResults[`${item.id}_chemistry`].message}
                    </ControlStatus>
                  )}
                </div>
              ) : (
                <span style={{ color: '#999' }}>Не задан</span>
              )
            ) : (
              <span style={{ color: '#999' }}>Недоступна</span>
            )
          )},
          { key: 'comment', label: 'Комментарий', accessor: (item) => (
            item.comment ? (
              <CommentCell>{item.comment}</CommentCell>
            ) : (
              <span style={{ color: '#999' }}>—</span>
            )
          )},
          { key: 'created_at', label: 'Создан', accessor: (item) => new Date(item.created_at).toLocaleDateString('ru-RU') }
        ]}
        getBorderColor={(washBox) => {
          switch (washBox.status) {
            case 'free': return '#28a745';
            case 'busy': return '#dc3545';
            case 'reserved': return '#ffc107';
            case 'maintenance': return '#17a2b8';
            default: return '#6c757d';
          }
        }}
        renderActions={(washBox) => [
          <ActionButton key="edit" theme={theme} onClick={() => openEditModal(washBox)}>
            Редактировать
          </ActionButton>,
          !isLimitedAdmin ? (
            <ActionButton key="delete" theme={theme} className="delete" onClick={() => handleDelete(washBox.id)}>
              Удалить
            </ActionButton>
          ) : null,
          <ActionButton key="sessions" theme={theme} onClick={() => {
            navigate('/admin/sessions', { 
              state: { 
                filters: { boxNumber: washBox.number },
                showBoxFilter: true 
              } 
            });
          }}>
            Сессии
          </ActionButton>
        ]}
        theme={theme}
      />

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
                  <option value="maintenance">Сервис</option>
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
              
              <FormGroup>
                <Label theme={theme}>Приоритет</Label>
                <Input
                  type="text"
                  value={formData.priority}
                  onChange={(e) => setFormData({ ...formData, priority: e.target.value.toUpperCase() })}
                  maxLength="1"
                  pattern="[A-Z]"
                  style={{ textTransform: 'uppercase' }}
                  required
                />
                <small style={{ color: '#666', fontSize: '12px' }}>
                  A - наивысший приоритет, затем B, C, D и так далее (только заглавные латинские буквы)
                </small>
              </FormGroup>
              
              {formData.serviceType === 'wash' && (
                <FormGroup>
                  <Label theme={theme}>
                    <input
                      type="checkbox"
                      checked={formData.chemistryEnabled}
                      onChange={(e) => setFormData({ ...formData, chemistryEnabled: e.target.checked })}
                      style={{ marginRight: '8px' }}
                    />
                    Химия включена
                  </Label>
                </FormGroup>
              )}
              
              <FormGroup>
                <Label theme={theme}>Комментарий</Label>
                <Textarea
                  value={formData.comment}
                  onChange={(e) => setFormData({ ...formData, comment: e.target.value })}
                  maxLength={1000}
                  rows={4}
                  placeholder="Дополнительная информация о боксе..."
                />
                <CharCounter $isNearLimit={(formData.comment?.length || 0) > 900}>
                  {formData.comment?.length || 0}/1000
                </CharCounter>
              </FormGroup>
              
              <FormGroup>
                <Label theme={theme}>Регистр света (0x0001)</Label>
                <Input
                  type="text"
                  value={formData.lightCoilRegister}
                  onChange={(e) => setFormData({ ...formData, lightCoilRegister: e.target.value })}
                  placeholder="0x0001"
                  pattern="0x[0-9a-fA-F]{1,4}"
                />
                <small style={{ color: '#666', fontSize: '12px' }}>
                  Hex формат регистра для управления светом
                </small>
              </FormGroup>
              
              {formData.serviceType === 'wash' && (
                <FormGroup>
                  <Label theme={theme}>Регистр химии (0x0002)</Label>
                  <Input
                    type="text"
                    value={formData.chemistryCoilRegister}
                    onChange={(e) => setFormData({ ...formData, chemistryCoilRegister: e.target.value })}
                    placeholder="0x00002"
                    pattern="0x[0-9a-fA-F]{1,4}"
                  />
                  <small style={{ color: '#666', fontSize: '12px' }}>
                    Hex формат регистра для управления химией (только для мойки)
                  </small>
                </FormGroup>
              )}
              
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
          <ModalContentWide>
            <ModalTitle theme={theme}>Редактировать бокс</ModalTitle>
            <Form onSubmit={handleUpdate}>
              {editingWashBox && (
                <FullWidthFormGroup>
                  <UUIDContainer>
                    <UUIDField>{editingWashBox.id}</UUIDField>
                    <CopyButton type="button" onClick={() => handleCopyUUID(editingWashBox.id)}>
                      📋 Копировать ID
                    </CopyButton>
                  </UUIDContainer>
                </FullWidthFormGroup>
              )}
              
              <FullWidthFormGroup>
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
              </FullWidthFormGroup>
              
              <TwoColumnGrid>
                <FormGroup>
                  <Label theme={theme}>Статус</Label>
                <Select
                    value={formData.status}
                  onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                  >
                  <option value="free" disabled={isLimitedAdmin && editingWashBox && editingWashBox.status !== 'maintenance'}>Свободен</option>
                  {!isLimitedAdmin && <option value="busy">Занят</option>}
                  {!isLimitedAdmin && <option value="reserved">Зарезервирован</option>}
                  <option value="maintenance">Сервис</option>
                  {!isLimitedAdmin && <option value="cleaning">На уборке</option>}
                  </Select>
                </FormGroup>
                
                {!isLimitedAdmin && (
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
                )}
                
                {!isLimitedAdmin && (
                <FormGroup>
                  <Label theme={theme}>Приоритет</Label>
                  <Input
                    type="text"
                    value={formData.priority}
                    onChange={(e) => setFormData({ ...formData, priority: e.target.value.toUpperCase() })}
                    maxLength="1"
                    pattern="[A-Z]"
                    style={{ textTransform: 'uppercase' }}
                    required
                  />
                  <small style={{ color: '#666', fontSize: '12px' }}>
                    A - высокий, B, C... - низкий
                  </small>
                </FormGroup>
                )}
                
                {!isLimitedAdmin && formData.serviceType === 'wash' && (
                  <FormGroup>
                    <Label theme={theme}>
                      <input
                        type="checkbox"
                        checked={formData.chemistryEnabled}
                        onChange={(e) => setFormData({ ...formData, chemistryEnabled: e.target.checked })}
                        style={{ marginRight: '8px' }}
                      />
                      Химия включена
                    </Label>
                  </FormGroup>
              )}
              </TwoColumnGrid>
              
              {!isLimitedAdmin && (
              <FullWidthFormGroup>
                <Label theme={theme}>Комментарий</Label>
                <Textarea
                  value={formData.comment}
                  onChange={(e) => setFormData({ ...formData, comment: e.target.value })}
                  maxLength={1000}
                  rows={3}
                  placeholder="Дополнительная информация о боксе..."
                />
                <CharCounter $isNearLimit={(formData.comment?.length || 0) > 900}>
                  {formData.comment?.length || 0}/1000
                </CharCounter>
              </FullWidthFormGroup>
              )}
              
              <TwoColumnGrid>
                {!isLimitedAdmin && (
                <FormGroup>
                  <Label theme={theme}>Регистр света (0x0001)</Label>
                  <Input
                    type="text"
                    value={formData.lightCoilRegister}
                    onChange={(e) => setFormData({ ...formData, lightCoilRegister: e.target.value })}
                    placeholder="0x0001"
                    pattern="0x[0-9a-fA-F]{1,4}"
                  />
                  <small style={{ color: '#666', fontSize: '12px' }}>
                    Hex формат
                  </small>
                  {formData.lightCoilRegister && editingWashBox && (
                    <TestButtonGroup>
                      <TestButton
                        type="button"
                        onClick={() => testCoil(editingWashBox.id, formData.lightCoilRegister, true)}
                        disabled={testingBox === editingWashBox.id}
                      >
                        Тест ВКЛ
                      </TestButton>
                      <TestButton
                        type="button"
                        onClick={() => testCoil(editingWashBox.id, formData.lightCoilRegister, false)}
                        disabled={testingBox === editingWashBox.id}
                      >
                        Тест ВЫКЛ
                      </TestButton>
                    </TestButtonGroup>
                  )}
                  {testResults[`${editingWashBox?.id}_${formData.lightCoilRegister}`] && (
                    <TestResult className={testResults[`${editingWashBox?.id}_${formData.lightCoilRegister}`].status}>
                      {testResults[`${editingWashBox?.id}_${formData.lightCoilRegister}`].message}
                    </TestResult>
                  )}
                </FormGroup>
                )}
                
                {!isLimitedAdmin && formData.serviceType === 'wash' && (
                  <FormGroup>
                    <Label theme={theme}>Регистр химии (0x0002)</Label>
                    <Input
                      type="text"
                      value={formData.chemistryCoilRegister}
                      onChange={(e) => setFormData({ ...formData, chemistryCoilRegister: e.target.value })}
                      placeholder="0x00002"
                      pattern="0x[0-9a-fA-F]{1,4}"
                    />
                    <small style={{ color: '#666', fontSize: '12px' }}>
                      Hex формат
                    </small>
                    {formData.chemistryCoilRegister && editingWashBox && (
                      <TestButtonGroup>
                        <TestButton
                          type="button"
                          onClick={() => testCoil(editingWashBox.id, formData.chemistryCoilRegister, true)}
                          disabled={testingBox === editingWashBox.id}
                        >
                          Тест ВКЛ
                        </TestButton>
                        <TestButton
                          type="button"
                          onClick={() => testCoil(editingWashBox.id, formData.chemistryCoilRegister, false)}
                          disabled={testingBox === editingWashBox.id}
                        >
                          Тест ВЫКЛ
                        </TestButton>
                      </TestButtonGroup>
                    )}
                    {testResults[`${editingWashBox?.id}_${formData.chemistryCoilRegister}`] && (
                      <TestResult className={testResults[`${editingWashBox?.id}_${formData.chemistryCoilRegister}`].status}>
                        {testResults[`${editingWashBox?.id}_${formData.chemistryCoilRegister}`].message}
                      </TestResult>
                    )}
                  </FormGroup>
                )}
              </TwoColumnGrid>
              
              <FullWidthFormGroup>
                <ButtonGroup>
                  <Button theme={theme} type="button" onClick={closeModals}>
                    Отмена
                  </Button>
                  <Button theme={theme} type="submit" disabled={loading}>
                    {loading ? 'Сохранение...' : 'Сохранить'}
                  </Button>
                </ButtonGroup>
              </FullWidthFormGroup>
            </Form>
          </ModalContentWide>
        </Modal>
      )}
    </Container>
  );
};

export default WashBoxManagement; 
