import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
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

const FilterInput = styled.input`
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
  
  &.pending {
    background-color: #fff3e0;
    color: #ef6c00;
  }
  
  &.processing {
    background-color: #e3f2fd;
    color: #1565c0;
  }
  
  &.completed {
    background-color: #e8f5e8;
    color: #2e7d32;
  }
  
  &.failed {
    background-color: #ffebee;
    color: #c62828;
  }
  
  &.refunded {
    background-color: #f3e5f5;
    color: #7b1fa2;
  }
  
  &.cancelled {
    background-color: #fafafa;
    color: #616161;
  }
`;

const TypeBadge = styled.span`
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  text-transform: uppercase;
  background-color: #e8f5e8;
  color: #2e7d32;
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
`;

const ErrorMessage = styled.div`
  color: #d32f2f;
  margin-bottom: 15px;
  font-size: 14px;
`;

const LoadingMessage = styled.div`
  color: ${props => props.theme.textColor};
  text-align: center;
  padding: 20px;
  font-size: 14px;
`;

const Pagination = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-top: 20px;
`;

const PageButton = styled.button`
  background-color: ${props => props.active ? props.theme.primaryColor : 'white'};
  color: ${props => props.active ? 'white' : props.theme.textColor};
  border: 1px solid #ddd;
  padding: 8px 12px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  
  &:hover {
    background-color: ${props => props.active ? props.theme.primaryColorDark : '#f5f5f5'};
  }
  
  &:disabled {
    background-color: #f5f5f5;
    color: #ccc;
    cursor: not-allowed;
  }
`;

// Модальное окно для деталей платежа
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
  width: 800px;
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

const PaymentDetails = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
`;

const DetailGroup = styled.div`
  display: flex;
  flex-direction: column;
  gap: 5px;
`;

const DetailLabel = styled.span`
  font-weight: 500;
  color: ${props => props.theme.textColor};
  font-size: 14px;
`;

const DetailValue = styled.span`
  color: ${props => props.theme.textColor};
  font-size: 14px;
`;

const RefundsSection = styled.div`
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #eee;
`;

const EventsSection = styled.div`
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #eee;
`;

const SectionTitle = styled.h4`
  margin: 0 0 15px 0;
  color: ${props => props.theme.textColor};
`;

const EventItem = styled.div`
  background-color: #f5f5f5;
  padding: 10px;
  border-radius: 5px;
  margin-bottom: 10px;
`;

const EventTime = styled.div`
  font-size: 12px;
  color: #666;
  margin-bottom: 5px;
`;

const EventDescription = styled.div`
  font-size: 14px;
  color: ${props => props.theme.textColor};
`;

/**
 * Компонент управления платежами
 * @returns {React.ReactNode} - Компонент управления платежами
 */
const PaymentManagement = () => {
  const theme = getTheme('light');
  const [payments, setPayments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filters, setFilters] = useState({
    status: '',
    type: '',
    dateFrom: '',
    dateTo: '',
    userId: ''
  });
  const [pagination, setPagination] = useState({
    limit: 20,
    offset: 0,
    total: 0
  });
  
  // Состояние для модального окна деталей платежа
  const [showPaymentModal, setShowPaymentModal] = useState(false);
  const [selectedPayment, setSelectedPayment] = useState(null);
  const [paymentDetails, setPaymentDetails] = useState(null);
  const [paymentDetailsLoading, setPaymentDetailsLoading] = useState(false);

  // Загрузка платежей при изменении фильтров или пагинации
  useEffect(() => {
    fetchPayments();
  }, [filters, pagination.offset]);

  // Функция для загрузки платежей
  const fetchPayments = async () => {
    try {
      setLoading(true);
      setError('');
      
      const params = {
        limit: pagination.limit,
        offset: pagination.offset,
        ...filters
      };
      
      const response = await ApiService.getPayments(params);
      setPayments(response.payments || []);
      setPagination(prev => ({
        ...prev,
        total: response.total || 0
      }));
    } catch (err) {
      setError('Ошибка при загрузке платежей');
      console.error('Error fetching payments:', err);
    } finally {
      setLoading(false);
    }
  };

  // Функция для загрузки деталей платежа
  const fetchPaymentDetails = async (paymentId) => {
    try {
      setPaymentDetailsLoading(true);
      const response = await ApiService.getPaymentDetails(paymentId);
      setPaymentDetails(response);
    } catch (err) {
      setError('Ошибка при загрузке деталей платежа');
      console.error('Error fetching payment details:', err);
    } finally {
      setPaymentDetailsLoading(false);
    }
  };

  // Обработчик открытия модального окна с деталями платежа
  const handleViewPayment = async (payment) => {
    setSelectedPayment(payment);
    setShowPaymentModal(true);
    await fetchPaymentDetails(payment.id);
  };

  // Обработчик изменения фильтров
  const handleFilterChange = (filterName, value) => {
    setFilters(prev => ({ ...prev, [filterName]: value }));
    setPagination(prev => ({ ...prev, offset: 0 }));
  };

  // Обработчик изменения страницы
  const handlePageChange = (newOffset) => {
    setPagination(prev => ({ ...prev, offset: newOffset }));
  };

  // Функция для форматирования суммы в рублях
  const formatAmount = (kopecks) => {
    return (kopecks / 100).toFixed(2);
  };

  // Функция для форматирования даты
  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString('ru-RU');
  };

  // Функция для получения названия типа платежа
  const getPaymentTypeName = (type) => {
    const names = {
      queue_booking: 'Оплата очереди',
      session_extension: 'Продление сессии',
      refund: 'Возврат'
    };
    return names[type] || type;
  };

  // Функция для получения названия статуса
  const getStatusName = (status) => {
    const names = {
      pending: 'Ожидает',
      processing: 'Обрабатывается',
      completed: 'Завершен',
      failed: 'Ошибка',
      refunded: 'Возвращен',
      cancelled: 'Отменен'
    };
    return names[status] || status;
  };

  const totalPages = Math.ceil(pagination.total / pagination.limit);
  const currentPage = Math.floor(pagination.offset / pagination.limit) + 1;

  if (loading && payments.length === 0) {
    return <LoadingMessage theme={theme}>Загрузка платежей...</LoadingMessage>;
  }

  return (
    <Container>
      <Header>
        <Title theme={theme}>Управление платежами</Title>
      </Header>

      {error && <ErrorMessage>{error}</ErrorMessage>}

      <Filters>
        <FilterSelect
          value={filters.status}
          onChange={(e) => handleFilterChange('status', e.target.value)}
        >
          <option value="">Все статусы</option>
          <option value="pending">Ожидает</option>
          <option value="processing">Обрабатывается</option>
          <option value="completed">Завершен</option>
          <option value="failed">Ошибка</option>
          <option value="refunded">Возвращен</option>
          <option value="cancelled">Отменен</option>
        </FilterSelect>

        <FilterSelect
          value={filters.type}
          onChange={(e) => handleFilterChange('type', e.target.value)}
        >
          <option value="">Все типы</option>
          <option value="queue_booking">Оплата очереди</option>
          <option value="session_extension">Продление сессии</option>
          <option value="refund">Возврат</option>
        </FilterSelect>

        <FilterInput
          type="text"
          value={filters.userId}
          onChange={(e) => handleFilterChange('userId', e.target.value)}
          placeholder="ID пользователя"
        />

        <FilterInput
          type="date"
          value={filters.dateFrom}
          onChange={(e) => handleFilterChange('dateFrom', e.target.value)}
          placeholder="Дата от"
        />

        <FilterInput
          type="date"
          value={filters.dateTo}
          onChange={(e) => handleFilterChange('dateTo', e.target.value)}
          placeholder="Дата до"
        />
      </Filters>

      <Table theme={theme}>
        <thead>
          <tr>
            <Th theme={theme}>ID</Th>
            <Th theme={theme}>Пользователь</Th>
            <Th theme={theme}>Тип</Th>
            <Th theme={theme}>Сумма</Th>
            <Th theme={theme}>Статус</Th>
            <Th theme={theme}>Дата создания</Th>
            <Th theme={theme}>Действия</Th>
          </tr>
        </thead>
        <tbody>
          {payments.map((payment) => (
            <tr key={payment.id}>
              <Td>{payment.id}</Td>
              <Td>{payment.user_id}</Td>
              <Td>
                <TypeBadge>{getPaymentTypeName(payment.type)}</TypeBadge>
              </Td>
              <Td>{formatAmount(payment.amount_kopecks)} ₽</Td>
              <Td>
                <StatusBadge className={payment.status}>
                  {getStatusName(payment.status)}
                </StatusBadge>
              </Td>
              <Td>{formatDate(payment.created_at)}</Td>
              <Td>
                <ActionButton onClick={() => handleViewPayment(payment)} theme={theme}>
                  Детали
                </ActionButton>
              </Td>
            </tr>
          ))}
        </tbody>
      </Table>

      {payments.length === 0 && !loading && (
        <div style={{ textAlign: 'center', padding: '20px', color: theme.textColor }}>
          Платежи не найдены
        </div>
      )}

      {totalPages > 1 && (
        <Pagination>
          <PageButton
            onClick={() => handlePageChange(0)}
            disabled={currentPage === 1}
            theme={theme}
          >
            Первая
          </PageButton>
          
          <PageButton
            onClick={() => handlePageChange(pagination.offset - pagination.limit)}
            disabled={currentPage === 1}
            theme={theme}
          >
            Предыдущая
          </PageButton>
          
          <span style={{ color: theme.textColor }}>
            Страница {currentPage} из {totalPages}
          </span>
          
          <PageButton
            onClick={() => handlePageChange(pagination.offset + pagination.limit)}
            disabled={currentPage === totalPages}
            theme={theme}
          >
            Следующая
          </PageButton>
          
          <PageButton
            onClick={() => handlePageChange((totalPages - 1) * pagination.limit)}
            disabled={currentPage === totalPages}
            theme={theme}
          >
            Последняя
          </PageButton>
        </Pagination>
      )}

      {/* Модальное окно с деталями платежа */}
      {showPaymentModal && (
        <Modal>
          <ModalContent>
            <ModalHeader>
              <ModalTitle theme={theme}>Детали платежа</ModalTitle>
              <CloseButton onClick={() => setShowPaymentModal(false)} theme={theme}>
                ✕
              </CloseButton>
            </ModalHeader>

            {paymentDetailsLoading ? (
              <div>Загрузка деталей...</div>
            ) : paymentDetails ? (
              <>
                <PaymentDetails>
                  <DetailGroup>
                    <DetailLabel theme={theme}>ID платежа:</DetailLabel>
                    <DetailValue theme={theme}>{paymentDetails.payment.id}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Пользователь:</DetailLabel>
                    <DetailValue theme={theme}>{paymentDetails.payment.user_id}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Тип:</DetailLabel>
                    <DetailValue theme={theme}>{getPaymentTypeName(paymentDetails.payment.type)}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Сумма:</DetailLabel>
                    <DetailValue theme={theme}>{formatAmount(paymentDetails.payment.amount_kopecks)} ₽</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Статус:</DetailLabel>
                    <DetailValue theme={theme}>
                      <StatusBadge className={paymentDetails.payment.status}>
                        {getStatusName(paymentDetails.payment.status)}
                      </StatusBadge>
                    </DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Дата создания:</DetailLabel>
                    <DetailValue theme={theme}>{formatDate(paymentDetails.payment.created_at)}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Tinkoff ID:</DetailLabel>
                    <DetailValue theme={theme}>{paymentDetails.payment.tinkoff_payment_id || 'Не указан'}</DetailValue>
                  </DetailGroup>
                  
                  <DetailGroup>
                    <DetailLabel theme={theme}>Описание:</DetailLabel>
                    <DetailValue theme={theme}>{paymentDetails.payment.description || 'Не указано'}</DetailValue>
                  </DetailGroup>
                </PaymentDetails>

                {/* Возвраты */}
                {paymentDetails.refunds && paymentDetails.refunds.length > 0 && (
                  <RefundsSection>
                    <SectionTitle theme={theme}>Возвраты</SectionTitle>
                    {paymentDetails.refunds.map((refund, index) => (
                      <EventItem key={index}>
                        <EventTime>{formatDate(refund.created_at)}</EventTime>
                        <EventDescription theme={theme}>
                          Возврат: {formatAmount(refund.amount_kopecks)} ₽ - {refund.status}
                        </EventDescription>
                      </EventItem>
                    ))}
                  </RefundsSection>
                )}

                {/* События */}
                {paymentDetails.events && paymentDetails.events.length > 0 && (
                  <EventsSection>
                    <SectionTitle theme={theme}>События</SectionTitle>
                    {paymentDetails.events.map((event, index) => (
                      <EventItem key={index}>
                        <EventTime>{formatDate(event.created_at)}</EventTime>
                        <EventDescription theme={theme}>
                          {event.event_type}: {event.description}
                        </EventDescription>
                      </EventItem>
                    ))}
                  </EventsSection>
                )}
              </>
            ) : (
              <div>Ошибка загрузки деталей платежа</div>
            )}
          </ModalContent>
        </Modal>
      )}
    </Container>
  );
};

export default PaymentManagement; 