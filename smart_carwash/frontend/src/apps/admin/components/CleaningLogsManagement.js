import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';
import { 
  convertDateTimeLocalToUTC, 
  convertUTCToDateTimeLocal, 
  getQuickFilterDates,
  formatDateForDisplay 
} from '../../../shared/utils/dateUtils';

const Container = styled.div`
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
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
`;

const FiltersContainer = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
  padding: 20px;
  background-color: ${props => props.theme.backgroundColor};
  border-radius: 8px;
  border: 1px solid ${props => props.theme.borderColor};
`;

const FilterGroup = styled.div`
  display: flex;
  flex-direction: column;
`;

const Label = styled.label`
  margin-bottom: 5px;
  font-weight: 500;
  color: ${props => props.theme.textColor};
`;

const Input = styled.input`
  padding: 8px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  font-size: 14px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  
  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }
`;

const Select = styled.select`
  padding: 8px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  font-size: 14px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  
  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }
`;

const Button = styled.button`
  padding: 10px 20px;
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.2s;
  
  &:hover {
    background-color: ${props => props.theme.primaryColorHover};
  }
  
  &:disabled {
    background-color: ${props => props.theme.disabledColor};
    cursor: not-allowed;
  }
`;

const TableContainer = styled.div`
  overflow-x: auto;
  margin-top: 20px;
  border-radius: 8px;
  border: 1px solid ${props => props.theme.borderColor};
`;

const Table = styled.table`
  width: 100%;
  min-width: 1000px;
  border-collapse: collapse;
`;

const TableHeader = styled.th`
  background-color: ${props => props.theme.tableHeaderBackground};
  color: ${props => props.theme.textColor};
  padding: 12px;
  text-align: left;
  font-weight: 600;
  border-bottom: 2px solid ${props => props.theme.borderColor};
`;

const TableCell = styled.td`
  padding: 12px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  color: ${props => props.theme.textColor};
`;

const TableRow = styled.tr`
  &:hover {
    background-color: ${props => props.theme.tableRowHover};
  }
`;

const StatusBadge = styled.span`
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  text-transform: uppercase;
  
  ${props => {
    switch (props.status) {
      case 'in_progress':
        return `
          background-color: #fef3c7;
          color: #92400e;
        `;
      case 'completed':
        return `
          background-color: #d1fae5;
          color: #065f46;
        `;
      case 'cancelled':
        return `
          background-color: #fee2e2;
          color: #991b1b;
        `;
      default:
        return `
          background-color: #f3f4f6;
          color: #374151;
        `;
    }
  }}
`;

const PaginationContainer = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding: 15px;
  background-color: ${props => props.theme.backgroundColor};
  border-radius: 8px;
`;

const PaginationInfo = styled.span`
  color: ${props => props.theme.textColor};
  font-size: 14px;
`;

const PaginationButtons = styled.div`
  display: flex;
  gap: 10px;
`;

const PaginationButton = styled.button`
  padding: 8px 12px;
  background-color: ${props => props.theme.cardBackground};
  color: ${props => props.theme.textColor};
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  
  &:hover:not(:disabled) {
    background-color: ${props => props.theme.primaryColor};
    color: white;
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

const LoadingSpinner = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px;
  color: ${props => props.theme.textColor};
`;

const ErrorMessage = styled.div`
  color: #dc3545;
  padding: 16px;
  background-color: #f8d7da;
  border: 1px solid #f5c6cb;
  border-radius: 6px;
  margin-bottom: 16px;
`;

const EmptyMessage = styled.div`
  text-align: center;
  padding: 40px;
  color: ${props => props.theme.textColor};
  font-style: italic;
`;

// Модальные окна для фильтров времени
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
  width: 500px;
  max-width: 90%;
  max-height: 80vh;
  overflow-y: auto;
`;

const ModalHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
`;

const ModalTitle = styled.h3`
  margin: 0;
  color: ${props => props.theme.textColor};
`;

const CloseButton = styled.button`
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: ${props => props.theme.textColor};
  
  &:hover {
    color: ${props => props.theme.primaryColor};
  }
`;

const CleaningLogsManagement = () => {
  const theme = getTheme('light');
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(20);

  // Фильтры
  const [filters, setFilters] = useState({
    status: ''
  });
  const [appliedFilters, setAppliedFilters] = useState({
    status: ''
  });
  const [localDateFilters, setLocalDateFilters] = useState({
    date_from: '',
    date_to: ''
  });

  // Состояние для модальных окон фильтров времени
  const [showDateFromModal, setShowDateFromModal] = useState(false);
  const [showDateToModal, setShowDateToModal] = useState(false);
  const [tempDateFrom, setTempDateFrom] = useState('');
  const [tempDateTo, setTempDateTo] = useState('');

  const loadLogs = async () => {
    setLoading(true);
    setError('');

    try {
      // Формируем объект с фильтрами и пагинацией
      const requestData = {
        limit: pageSize,
        offset: (currentPage - 1) * pageSize
      };
      
      // Добавляем примененные фильтры только если они не пустые
      if (appliedFilters.status && appliedFilters.status.trim() !== '') {
        requestData.status = appliedFilters.status;
      }
      
      // Конвертируем локальные даты в UTC перед отправкой
      if (localDateFilters.date_from && localDateFilters.date_from.trim() !== '') {
        requestData.date_from = convertDateTimeLocalToUTC(localDateFilters.date_from);
      }
      if (localDateFilters.date_to && localDateFilters.date_to.trim() !== '') {
        requestData.date_to = convertDateTimeLocalToUTC(localDateFilters.date_to);
      }

      const response = await ApiService.getCleaningLogs(requestData);
      setLogs(response.cleaning_logs || []);
      setTotal(response.total || 0);
    } catch (err) {
      setError('Ошибка при загрузке логов уборки: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadLogs();
  }, [currentPage, appliedFilters, localDateFilters]);

  const handleFilterChange = (field, value) => {
    setFilters(prev => ({
      ...prev,
      [field]: value
    }));
    setCurrentPage(1); // Сбрасываем на первую страницу при изменении фильтров
  };

  const handleDateFilterChange = (field, value) => {
    setLocalDateFilters(prev => ({
      ...prev,
      [field]: value
    }));
    setCurrentPage(1); // Сбрасываем на первую страницу при изменении фильтров
  };

  const handleSearch = () => {
    setAppliedFilters(filters);
    setCurrentPage(1);
  };

  const handleClearFilters = () => {
    const emptyFilters = {
      status: ''
    };
    const emptyLocalDateFilters = {
      date_from: '',
      date_to: ''
    };
    setFilters(emptyFilters);
    setAppliedFilters(emptyFilters);
    setLocalDateFilters(emptyLocalDateFilters);
    setCurrentPage(1);
  };

  // Функции для модальных окон фильтров времени
  const openDateFromModal = () => {
    setTempDateFrom(localDateFilters.date_from);
    setShowDateFromModal(true);
  };

  const openDateToModal = () => {
    setTempDateTo(localDateFilters.date_to);
    setShowDateToModal(true);
  };

  const applyDateFromFilter = () => {
    setLocalDateFilters(prev => ({ ...prev, date_from: tempDateFrom }));
    setShowDateFromModal(false);
    setCurrentPage(1);
  };

  const applyDateToFilter = () => {
    setLocalDateFilters(prev => ({ ...prev, date_to: tempDateTo }));
    setShowDateToModal(false);
    setCurrentPage(1);
  };

  const cancelDateFromFilter = () => {
    setTempDateFrom(localDateFilters.date_from);
    setShowDateFromModal(false);
  };

  const cancelDateToFilter = () => {
    setTempDateTo(localDateFilters.date_to);
    setShowDateToModal(false);
  };

  const formatDateTime = (dateString) => {
    if (!dateString) return '-';
    return formatDateForDisplay(dateString);
  };

  const formatDuration = (minutes) => {
    if (!minutes) return '-';
    return `${minutes} мин`;
  };

  const getStatusText = (status) => {
    switch (status) {
      case 'in_progress':
        return 'В процессе';
      case 'completed':
        return 'Завершена';
      case 'cancelled':
        return 'Отменена';
      default:
        return status;
    }
  };

  const totalPages = Math.ceil(total / pageSize);

  return (
    <Container>
      <Card theme={theme}>
        <Title theme={theme}>Логи уборки</Title>
        
        {error && <ErrorMessage>{error}</ErrorMessage>}

        <FiltersContainer theme={theme}>
          <FilterGroup>
            <Label theme={theme}>Статус</Label>
            <Select
              theme={theme}
              value={filters.status}
              onChange={(e) => handleFilterChange('status', e.target.value)}
            >
              <option value="">Все статусы</option>
              <option value="in_progress">В процессе</option>
              <option value="completed">Завершена</option>
              <option value="cancelled">Отменена</option>
            </Select>
          </FilterGroup>

          <FilterGroup>
            <Label theme={theme}>Дата с</Label>
            <Button 
              theme={theme} 
              onClick={openDateFromModal}
              style={{ 
                padding: '10px 20px', 
                backgroundColor: localDateFilters.date_from ? theme.primaryColor : '#f5f5f5',
                color: localDateFilters.date_from ? 'white' : theme.textColor,
                border: '1px solid #ddd',
                borderRadius: '4px',
                cursor: 'pointer',
                width: '100%',
                textAlign: 'left'
              }}
            >
              {localDateFilters.date_from ? `Дата с: ${localDateFilters.date_from}` : 'Выбрать дату с'}
            </Button>
          </FilterGroup>

          <FilterGroup>
            <Label theme={theme}>Дата по</Label>
            <Button 
              theme={theme} 
              onClick={openDateToModal}
              style={{ 
                padding: '10px 20px', 
                backgroundColor: localDateFilters.date_to ? theme.primaryColor : '#f5f5f5',
                color: localDateFilters.date_to ? 'white' : theme.textColor,
                border: '1px solid #ddd',
                borderRadius: '4px',
                cursor: 'pointer',
                width: '100%',
                textAlign: 'left'
              }}
            >
              {localDateFilters.date_to ? `Дата по: ${localDateFilters.date_to}` : 'Выбрать дату по'}
            </Button>
          </FilterGroup>

          <FilterGroup>
            <Label theme={theme}>&nbsp;</Label>
            <div style={{ display: 'flex', gap: '10px' }}>
              <Button theme={theme} onClick={handleSearch}>
                Поиск
              </Button>
              <Button theme={theme} onClick={handleClearFilters}>
                Очистить
              </Button>
            </div>
          </FilterGroup>
        </FiltersContainer>

        {loading ? (
          <LoadingSpinner theme={theme}>Загрузка логов уборки...</LoadingSpinner>
        ) : logs.length === 0 ? (
          <EmptyMessage theme={theme}>Логи уборки не найдены</EmptyMessage>
        ) : (
          <>
            <TableContainer theme={theme}>
              <Table>
                <thead>
                  <tr>
                    <TableHeader theme={theme}>ID</TableHeader>
                    <TableHeader theme={theme}>Уборщик</TableHeader>
                    <TableHeader theme={theme}>Бокс</TableHeader>
                    <TableHeader theme={theme}>Тип бокса</TableHeader>
                    <TableHeader theme={theme}>Начало</TableHeader>
                    <TableHeader theme={theme}>Завершение</TableHeader>
                    <TableHeader theme={theme}>Длительность</TableHeader>
                    <TableHeader theme={theme}>Статус</TableHeader>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log) => (
                    <TableRow key={log.id} theme={theme}>
                      <TableCell theme={theme}>{log.id}</TableCell>
                      <TableCell theme={theme}>{log.cleaner_username || 'Неизвестно'}</TableCell>
                      <TableCell theme={theme}>Бокс {log.wash_box_number}</TableCell>
                      <TableCell theme={theme}>
                        {log.wash_box_type === 'wash' ? 'Мойка' :
                         log.wash_box_type === 'air_dry' ? 'Обдув' :
                         log.wash_box_type === 'vacuum' ? 'Пылесос' : log.wash_box_type}
                      </TableCell>
                      <TableCell theme={theme}>{formatDateTime(log.started_at)}</TableCell>
                      <TableCell theme={theme}>{formatDateTime(log.completed_at)}</TableCell>
                      <TableCell theme={theme}>{formatDuration(log.duration_minutes)}</TableCell>
                      <TableCell theme={theme}>
                        <StatusBadge status={log.status}>
                          {getStatusText(log.status)}
                        </StatusBadge>
                      </TableCell>
                    </TableRow>
                  ))}
                </tbody>
              </Table>
            </TableContainer>

            <PaginationContainer theme={theme}>
              <PaginationInfo theme={theme}>
                Показано {((currentPage - 1) * pageSize) + 1}-{Math.min(currentPage * pageSize, total)} из {total} записей
              </PaginationInfo>
              
              <PaginationButtons>
                <PaginationButton
                  theme={theme}
                  onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                  disabled={currentPage === 1}
                >
                  Назад
                </PaginationButton>
                
                <PaginationButton
                  theme={theme}
                  onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                  disabled={currentPage === totalPages}
                >
                  Вперед
                </PaginationButton>
              </PaginationButtons>
            </PaginationContainer>
          </>
        )}
      </Card>

      {/* Модальное окно для выбора даты "с" */}
      {showDateFromModal && (
        <Modal>
          <ModalContent>
            <ModalHeader>
              <ModalTitle theme={theme}>Выберите дату "с"</ModalTitle>
              <CloseButton onClick={cancelDateFromFilter} theme={theme}>×</CloseButton>
            </ModalHeader>
            
            <div style={{ marginBottom: '20px' }}>
              <label style={{ display: 'block', marginBottom: '10px', fontWeight: '500' }}>
                Дата и время:
              </label>
              <input
                type="datetime-local"
                value={tempDateFrom}
                onChange={(e) => setTempDateFrom(e.target.value)}
                style={{
                  width: '100%',
                  padding: '10px',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                  fontSize: '16px'
                }}
              />
            </div>

            <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
              <Button theme={theme} onClick={cancelDateFromFilter} style={{ backgroundColor: '#6c757d' }}>
                Отмена
              </Button>
              <Button theme={theme} onClick={applyDateFromFilter}>
                Применить
              </Button>
            </div>
          </ModalContent>
        </Modal>
      )}

      {/* Модальное окно для выбора даты "по" */}
      {showDateToModal && (
        <Modal>
          <ModalContent>
            <ModalHeader>
              <ModalTitle theme={theme}>Выберите дату "по"</ModalTitle>
              <CloseButton onClick={cancelDateToFilter} theme={theme}>×</CloseButton>
            </ModalHeader>
            
            <div style={{ marginBottom: '20px' }}>
              <label style={{ display: 'block', marginBottom: '10px', fontWeight: '500' }}>
                Дата и время:
              </label>
              <input
                type="datetime-local"
                value={tempDateTo}
                onChange={(e) => setTempDateTo(e.target.value)}
                style={{
                  width: '100%',
                  padding: '10px',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                  fontSize: '16px'
                }}
              />
            </div>

            <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
              <Button theme={theme} onClick={cancelDateToFilter} style={{ backgroundColor: '#6c757d' }}>
                Отмена
              </Button>
              <Button theme={theme} onClick={applyDateToFilter}>
                Применить
              </Button>
            </div>
          </ModalContent>
        </Modal>
      )}
    </Container>
  );
};

export default CleaningLogsManagement;
