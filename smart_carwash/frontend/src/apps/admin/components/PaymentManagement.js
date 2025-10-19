import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { useSearchParams, useLocation } from 'react-router-dom';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';
import PaymentStatistics from './PaymentStatistics';
import { 
  convertDateTimeLocalToUTC, 
  convertUTCToDateTimeLocal, 
  getQuickFilterDates,
  formatDateForDisplay 
} from '../../../shared/utils/dateUtils';

const Container = styled.div`
  padding: 20px;
`;

const Title = styled.h1`
  margin: 0 0 30px 0;
  color: ${props => props.theme.textColor};
  font-size: 2rem;
`;

const FiltersContainer = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
`;

const FiltersTitle = styled.h3`
  margin: 0 0 15px 0;
  color: ${props => props.theme.textColor};
  font-size: 1.1rem;
`;

const FiltersGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
  margin-bottom: 15px;
`;

const FilterGroup = styled.div`
  display: flex;
  flex-direction: column;
`;

const FilterLabel = styled.label`
  margin-bottom: 5px;
  color: ${props => props.theme.textColor};
  font-size: 0.9rem;
  font-weight: 500;
`;

const FilterInput = styled.input`
  padding: 8px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  font-size: 0.9rem;

  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }
`;

const FilterSelect = styled.select`
  padding: 8px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  font-size: 0.9rem;

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
  font-size: 0.9rem;
  font-weight: 500;
  transition: background-color 0.2s ease;

  &:hover {
    background-color: ${props => props.theme.primaryColorHover};
  }

  &:disabled {
    background-color: ${props => props.theme.disabledColor};
    cursor: not-allowed;
  }
`;

const ButtonGroup = styled.div`
  display: flex;
  gap: 10px;
  margin-top: 15px;
`;

const TableContainer = styled.div`
  background-color: ${props => props.theme.cardBackground};
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
`;

const TableHeader = styled.th`
  padding: 15px;
  text-align: left;
  background-color: ${props => props.theme.tableHeaderBackground};
  color: ${props => props.theme.textColor};
  font-weight: 600;
  font-size: 0.9rem;
  border-bottom: 1px solid ${props => props.theme.borderColor};
`;

const TableCell = styled.td`
  padding: 12px 15px;
  border-bottom: 1px solid ${props => props.theme.borderColor};
  color: ${props => props.theme.textColor};
  font-size: 0.9rem;
`;

const StatusBadge = styled.span`
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  text-transform: uppercase;
  background-color: ${props => {
    switch (props.status) {
      case 'succeeded': return '#10b981';
      case 'pending': return '#f59e0b';
      case 'failed': return '#ef4444';
      case 'refunded': return '#6b7280';
      default: return '#6b7280';
    }
  }};
  color: white;
`;

const RefundButton = styled.button`
  padding: 6px 12px;
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  font-weight: 500;
  transition: background-color 0.2s ease;

  &:hover {
    background-color: #dc2626;
  }

  &:disabled {
    background-color: #9ca3af;
    cursor: not-allowed;
  }
`;

const PaginationContainer = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background-color: ${props => props.theme.cardBackground};
  border-top: 1px solid ${props => props.theme.borderColor};
`;

const PaginationInfo = styled.span`
  color: ${props => props.theme.textColor};
  font-size: 0.9rem;
`;

const PaginationButton = styled.button`
  padding: 8px 12px;
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  font-weight: 500;
  transition: background-color 0.2s ease;

  &:hover {
    background-color: ${props => props.theme.primaryColorHover};
  }

  &:disabled {
    background-color: ${props => props.theme.disabledColor};
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
  padding: 15px;
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 4px;
  color: #dc2626;
  margin-bottom: 20px;
`;

const PaymentManagement = () => {
  const theme = getTheme('light');
  const [searchParams, setSearchParams] = useSearchParams();
  const location = useLocation();
  
  const [payments, setPayments] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showStatistics, setShowStatistics] = useState(false);
  const [filters, setFilters] = useState({
    payment_id: '',
    session_id: '',
    user_id: '',
    status: '',
    payment_type: '',
    payment_method: ''
  });
  const [localFilters, setLocalFilters] = useState({
    date_from: '',
    date_to: ''
  });
  const [pagination, setPagination] = useState({
    total: 0,
    limit: 20,
    offset: 0
  });
  const [isInitialized, setIsInitialized] = useState(false);

  // Инициализация фильтров из URL параметров
  useEffect(() => {
    const sessionId = searchParams.get('session_id');
    const userId = searchParams.get('user_id');
    
    if (sessionId || userId) {
      const newFilters = {
        payment_id: '',
        session_id: sessionId || '',
        user_id: userId || '',
        status: '',
        payment_type: '',
        payment_method: ''
      };
      
      setFilters(newFilters);
      setPagination(prev => ({ ...prev, offset: 0 }));
      setIsInitialized(true);
    } else {
      setIsInitialized(true);
    }
  }, [searchParams]);

  const statusOptions = [
    { value: '', label: 'Все статусы' },
    { value: 'pending', label: 'Ожидает оплаты' },
    { value: 'succeeded', label: 'Оплачен' },
    { value: 'failed', label: 'Ошибка' },
    { value: 'refunded', label: 'Возвращен' }
  ];

  const paymentTypeOptions = [
    { value: '', label: 'Все типы' },
    { value: 'main', label: 'Основной' },
    { value: 'extension', label: 'Продление' }
  ];

  const paymentMethodOptions = [
    { value: '', label: 'Все методы' },
    { value: 'tinkoff', label: 'Tinkoff' },
    { value: 'cashier', label: 'Кассир' }
  ];

  const loadPayments = async () => {
    setLoading(true);
    setError('');

    try {
      const params = {
        limit: pagination.limit,
        offset: pagination.offset
      };

      // Добавляем только непустые фильтры (исключая date_from и date_to)
      Object.keys(filters).forEach(key => {
        if (filters[key] && key !== 'date_from' && key !== 'date_to') {
          params[key] = filters[key];
        }
      });

      // Конвертируем локальные даты в UTC перед отправкой
      if (localFilters.date_from) {
        params.date_from = convertDateTimeLocalToUTC(localFilters.date_from);
      }
      if (localFilters.date_to) {
        params.date_to = convertDateTimeLocalToUTC(localFilters.date_to);
      }

      const response = await ApiService.getPayments(params);
      setPayments(response.payments || []);
      setPagination(prev => ({
        ...prev,
        total: response.total || 0
      }));
    } catch (err) {
      setError('Ошибка при загрузке платежей: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isInitialized) {
      loadPayments();
    }
  }, [pagination.offset, filters, localFilters, isInitialized]);

  const handleFilterChange = (key, value) => {
    setFilters(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const handleDateFilterChange = (key, value) => {
    setLocalFilters(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const handleApplyFilters = () => {
    setPagination(prev => ({ ...prev, offset: 0 }));
    loadPayments();
  };

  const handleClearFilters = () => {
    const clearedFilters = {
      payment_id: '',
      session_id: '',
      user_id: '',
      status: '',
      payment_type: '',
      payment_method: ''
    };
    const clearedLocalFilters = {
      date_from: '',
      date_to: ''
    };
    setFilters(clearedFilters);
    setLocalFilters(clearedLocalFilters);
    setPagination(prev => ({ ...prev, offset: 0 }));
    loadPayments();
  };

  const handleRefund = async (paymentId, amount) => {
    if (!window.confirm(`Вы уверены, что хотите вернуть ${amount / 100} руб. за этот платеж?`)) {
      return;
    }

    try {
      await ApiService.refundPayment({
        paymentId: paymentId,
        amount: amount
      });
      
      // Перезагружаем платежи
      loadPayments();
      window.alert('Возврат выполнен успешно');
    } catch (err) {
      window.alert('Ошибка при выполнении возврата: ' + (err.message || 'Неизвестная ошибка'));
    }
  };

  const formatAmount = (amount) => {
    return `${(amount / 100).toFixed(2)} руб.`;
  };

  const formatDate = (dateString) => {
    return formatDateForDisplay(dateString);
  };

  const canRefund = (payment) => {
    return payment.status === 'succeeded' && 
           payment.refunded_amount < payment.amount && 
           payment.payment_method !== 'cashier';
  };

  const getRefundAmount = (payment) => {
    return payment.amount - payment.refunded_amount;
  };

  return (
    <Container>
      <Title theme={theme}>Управление платежами</Title>

      {error && <ErrorMessage>{error}</ErrorMessage>}

      <FiltersContainer theme={theme}>
        <FiltersTitle theme={theme}>Фильтры</FiltersTitle>
        
        <FiltersGrid>
          <FilterGroup>
            <FilterLabel theme={theme}>ID платежа</FilterLabel>
            <FilterInput
              theme={theme}
              type="text"
              value={filters.payment_id}
              onChange={(e) => handleFilterChange('payment_id', e.target.value)}
              placeholder="UUID платежа"
            />
          </FilterGroup>

          <FilterGroup>
            <FilterLabel theme={theme}>ID сессии</FilterLabel>
            <FilterInput
              theme={theme}
              type="text"
              value={filters.session_id}
              onChange={(e) => handleFilterChange('session_id', e.target.value)}
              placeholder="UUID сессии"
            />
          </FilterGroup>

          <FilterGroup>
            <FilterLabel theme={theme}>ID клиента</FilterLabel>
            <FilterInput
              theme={theme}
              type="text"
              value={filters.user_id}
              onChange={(e) => handleFilterChange('user_id', e.target.value)}
              placeholder="UUID клиента"
            />
          </FilterGroup>

          <FilterGroup>
            <FilterLabel theme={theme}>Статус</FilterLabel>
            <FilterSelect
              theme={theme}
              value={filters.status}
              onChange={(e) => handleFilterChange('status', e.target.value)}
            >
              {statusOptions.map(option => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </FilterSelect>
          </FilterGroup>

          <FilterGroup>
            <FilterLabel theme={theme}>Тип платежа</FilterLabel>
            <FilterSelect
              theme={theme}
              value={filters.payment_type}
              onChange={(e) => handleFilterChange('payment_type', e.target.value)}
            >
              {paymentTypeOptions.map(option => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </FilterSelect>
          </FilterGroup>

          <FilterGroup>
            <FilterLabel theme={theme}>Метод платежа</FilterLabel>
            <FilterSelect
              theme={theme}
              value={filters.payment_method}
              onChange={(e) => handleFilterChange('payment_method', e.target.value)}
            >
              {paymentMethodOptions.map(option => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </FilterSelect>
          </FilterGroup>

          <FilterGroup>
            <FilterLabel theme={theme}>Дата от</FilterLabel>
            <FilterInput
              theme={theme}
              type="datetime-local"
              value={localFilters.date_from}
              onChange={(e) => handleDateFilterChange('date_from', e.target.value)}
            />
          </FilterGroup>

          <FilterGroup>
            <FilterLabel theme={theme}>Дата до</FilterLabel>
            <FilterInput
              theme={theme}
              type="datetime-local"
              value={localFilters.date_to}
              onChange={(e) => handleDateFilterChange('date_to', e.target.value)}
            />
          </FilterGroup>

          <FilterGroup>
            <FilterLabel theme={theme}>Быстрые фильтры</FilterLabel>
            <FilterSelect
              theme={theme}
              onChange={(e) => {
                const filterType = e.target.value;
                
                if (filterType) {
                  const { dateFrom, dateTo } = getQuickFilterDates(filterType);
                  
                  if (dateFrom && dateTo) {
                    // Конвертируем UTC даты обратно в datetime-local формат для отображения
                    const localDateFrom = convertUTCToDateTimeLocal(dateFrom);
                    const localDateTo = convertUTCToDateTimeLocal(dateTo);
                    
                    setLocalFilters({
                      date_from: localDateFrom,
                      date_to: localDateTo
                    });
                  }
                } else {
                  // Очищаем фильтры
                  setLocalFilters({
                    date_from: '',
                    date_to: ''
                  });
                }
              }}
            >
              <option value="">Выберите период</option>
              <option value="today">Сегодня</option>
              <option value="yesterday">Вчера</option>
              <option value="this_week">Эта неделя</option>
              <option value="last_week">Прошлая неделя</option>
              <option value="this_month">Этот месяц</option>
              <option value="last_month">Прошлый месяц</option>
              <option value="this_year">Этот год</option>
              <option value="last_year">Прошлый год</option>
            </FilterSelect>
          </FilterGroup>
        </FiltersGrid>

        <ButtonGroup>
          <Button theme={theme} onClick={handleApplyFilters}>
            Применить фильтры
          </Button>
          <Button theme={theme} onClick={handleClearFilters}>
            Очистить фильтры
          </Button>
          <Button 
            theme={theme} 
            onClick={() => setShowStatistics(!showStatistics)}
            style={{ backgroundColor: showStatistics ? theme.secondaryColor : theme.primaryColor }}
          >
            {showStatistics ? 'Скрыть статистику' : 'Показать статистику'}
          </Button>
        </ButtonGroup>
      </FiltersContainer>

      {showStatistics && (
        <PaymentStatistics 
          filters={filters}
          localDateFilters={localFilters}
          onClose={() => setShowStatistics(false)}
        />
      )}

      <TableContainer theme={theme}>
        {loading ? (
          <LoadingSpinner theme={theme}>Загрузка платежей...</LoadingSpinner>
        ) : (
          <>
            <Table>
              <thead>
                <tr>
                  <TableHeader theme={theme}>ID</TableHeader>
                  <TableHeader theme={theme}>Сессия</TableHeader>
                  <TableHeader theme={theme}>Сумма</TableHeader>
                  <TableHeader theme={theme}>Статус</TableHeader>
                  <TableHeader theme={theme}>Тип</TableHeader>
                  <TableHeader theme={theme}>Метод</TableHeader>
                  <TableHeader theme={theme}>Возврат</TableHeader>
                  <TableHeader theme={theme}>Дата создания</TableHeader>
                  <TableHeader theme={theme}>Действия</TableHeader>
                </tr>
              </thead>
              <tbody>
                {payments.map(payment => (
                  <tr key={payment.id}>
                    <TableCell theme={theme}>
                      <code>{payment.id}</code>
                    </TableCell>
                    <TableCell theme={theme}>
                      <code>{payment.session_id}</code>
                    </TableCell>
                    <TableCell theme={theme}>
                      {formatAmount(payment.amount)}
                    </TableCell>
                    <TableCell theme={theme}>
                      <StatusBadge status={payment.status}>
                        {payment.status}
                      </StatusBadge>
                    </TableCell>
                    <TableCell theme={theme}>
                      {payment.payment_type === 'main' ? 'Основной' : 'Продление'}
                    </TableCell>
                    <TableCell theme={theme}>
                      {payment.payment_method === 'tinkoff' ? 'Tinkoff' : 
                       payment.payment_method === 'cashier' ? 'Кассир' : 
                       payment.payment_method || '-'}
                    </TableCell>
                    <TableCell theme={theme}>
                      {payment.refunded_amount > 0 ? formatAmount(payment.refunded_amount) : '-'}
                    </TableCell>
                    <TableCell theme={theme}>
                      {formatDate(payment.created_at)}
                    </TableCell>
                    <TableCell theme={theme}>
                      {canRefund(payment) && (
                        <RefundButton
                          onClick={() => handleRefund(payment.id, getRefundAmount(payment))}
                        >
                          Возврат
                        </RefundButton>
                      )}
                    </TableCell>
                  </tr>
                ))}
              </tbody>
            </Table>

            <PaginationContainer theme={theme}>
              <PaginationInfo theme={theme}>
                Показано {pagination.offset + 1}-{Math.min(pagination.offset + pagination.limit, pagination.total)} из {pagination.total} платежей
              </PaginationInfo>
              
              <div style={{ display: 'flex', gap: '10px' }}>
                <PaginationButton
                  theme={theme}
                  disabled={pagination.offset === 0}
                  onClick={() => setPagination(prev => ({ ...prev, offset: Math.max(0, prev.offset - prev.limit) }))}
                >
                  Назад
                </PaginationButton>
                <PaginationButton
                  theme={theme}
                  disabled={pagination.offset + pagination.limit >= pagination.total}
                  onClick={() => setPagination(prev => ({ ...prev, offset: prev.offset + prev.limit }))}
                >
                  Вперед
                </PaginationButton>
              </div>
            </PaginationContainer>
          </>
        )}
      </TableContainer>
    </Container>
  );
};

export default PaymentManagement; 