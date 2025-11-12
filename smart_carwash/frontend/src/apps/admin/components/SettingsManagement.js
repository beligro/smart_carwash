import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { getTheme } from '../../../shared/styles/theme';
import ApiService from '../../../shared/services/ApiService';

const Container = styled.div`
  padding: 20px;
`;

const Title = styled.h1`
  margin: 0 0 30px 0;
  color: ${props => props.theme.textColor};
  font-size: 2rem;
`;

const ServiceSelector = styled.div`
  margin-bottom: 30px;
`;

const ServiceLabel = styled.label`
  display: block;
  margin-bottom: 10px;
  color: ${props => props.theme.textColor};
  font-weight: 500;
`;

const ServiceSelect = styled.select`
  padding: 10px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  font-size: 1rem;
  min-width: 200px;

  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }
`;

const SettingsContainer = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 25px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
`;

const SettingsTitle = styled.h2`
  margin: 0 0 20px 0;
  color: ${props => props.theme.textColor};
  font-size: 1.5rem;
`;

const FormGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 25px;
`;

const FormGroup = styled.div`
  display: flex;
  flex-direction: column;
`;

const FormLabel = styled.label`
  margin-bottom: 8px;
  color: ${props => props.theme.textColor};
  font-weight: 500;
  font-size: 0.9rem;
`;

const FormInput = styled.input`
  padding: 10px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  font-size: 1rem;

  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }

  &[type="number"] {
    width: 100%;
  }
`;

const SettingsField = styled.div`
  margin-bottom: 20px;
`;

const SettingsLabel = styled.label`
  display: block;
  margin-bottom: 8px;
  color: ${props => props.theme.textColor};
  font-weight: 500;
  font-size: 0.9rem;
`;

const SettingsInput = styled.input`
  padding: 10px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  font-size: 1rem;
  width: 100%;
  max-width: 200px;

  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

const SettingsDescription = styled.div`
  margin-top: 5px;
  color: ${props => props.theme.textColorSecondary};
  font-size: 0.8rem;
  line-height: 1.4;
`;

const RentalTimesContainer = styled.div`
  margin-top: 20px;
`;

const RentalTimesTitle = styled.h3`
  margin: 0 0 15px 0;
  color: ${props => props.theme.textColor};
  font-size: 1.1rem;
`;

const RentalTimesList = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 15px;
`;

const RentalTimeItem = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border-radius: 20px;
  font-size: 0.9rem;
`;

const RemoveButton = styled.button`
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  font-size: 1.2rem;
  padding: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;

  &:hover {
    opacity: 0.8;
  }
`;

const AddRentalTimeContainer = styled.div`
  display: flex;
  gap: 10px;
  align-items: center;
`;

const AddRentalTimeInput = styled.input`
  padding: 8px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  font-size: 0.9rem;
  width: 80px;

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
  gap: 15px;
  margin-top: 25px;
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

const SuccessMessage = styled.div`
  padding: 15px;
  background-color: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 4px;
  color: #16a34a;
  margin-bottom: 20px;
`;

const CarwashStatusContainer = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 25px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
  border: 2px solid ${props => props.isClosed ? '#dc3545' : '#28a745'};
`;

const CarwashStatusTitle = styled.h2`
  margin: 0 0 20px 0;
  color: ${props => props.theme.textColor};
  font-size: 1.5rem;
  display: flex;
  align-items: center;
  gap: 10px;
`;

const StatusIndicator = styled.span`
  font-size: 1.2rem;
`;

const WarningText = styled.div`
  padding: 15px;
  background-color: #fff3cd;
  border: 1px solid #ffc107;
  border-radius: 4px;
  color: #856404;
  margin-bottom: 20px;
  font-size: 0.9rem;
  line-height: 1.5;
`;

const ModalOverlay = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
`;

const ModalContent = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 30px;
  border-radius: 8px;
  max-width: 500px;
  width: 90%;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
`;

const ModalTitle = styled.h3`
  margin: 0 0 20px 0;
  color: ${props => props.theme.textColor};
  font-size: 1.3rem;
`;

const ModalText = styled.p`
  margin: 0 0 20px 0;
  color: ${props => props.theme.textColor};
  line-height: 1.5;
`;

const ModalTextarea = styled.textarea`
  width: 100%;
  padding: 10px 12px;
  border: 1px solid ${props => props.theme.borderColor};
  border-radius: 4px;
  background-color: ${props => props.theme.inputBackground};
  color: ${props => props.theme.textColor};
  font-size: 1rem;
  min-height: 80px;
  margin-bottom: 20px;
  font-family: inherit;
  resize: vertical;

  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }
`;

const ModalButtonGroup = styled.div`
  display: flex;
  gap: 10px;
  justify-content: flex-end;
`;

const DangerButton = styled(Button)`
  background-color: #dc3545;
  
  &:hover {
    background-color: #c82333;
  }
  
  &:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }
`;

const SuccessButtonStyled = styled(Button)`
  background-color: #28a745;
  
  &:hover {
    background-color: #218838;
  }
  
  &:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }
`;

const SettingsManagement = () => {
  const theme = getTheme('light');
  const [selectedService, setSelectedService] = useState('wash');
  const [settings, setSettings] = useState({
    price_per_minute: '',
    chemistry_price_per_minute: '',
    available_rental_times: [],
    available_chemistry_times: []
  });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [newRentalTime, setNewRentalTime] = useState('');
  const [newChemistryTime, setNewChemistryTime] = useState('');
  const [cleaningTimeout, setCleaningTimeout] = useState('');
  const [cleaningTimeoutLoading, setCleaningTimeoutLoading] = useState(false);
  const [sessionTimeout, setSessionTimeout] = useState('');
  const [sessionTimeoutLoading, setSessionTimeoutLoading] = useState(false);
  const [cooldownTimeout, setCooldownTimeout] = useState('');
  const [cooldownTimeoutLoading, setCooldownTimeoutLoading] = useState(false);
  const [carwashStatus, setCarwashStatus] = useState(null);
  const [carwashStatusLoading, setCarwashStatusLoading] = useState(false);
  const [carwashActionLoading, setCarwashActionLoading] = useState(false);
  const [showCloseConfirm, setShowCloseConfirm] = useState(false);
  const [showOpenConfirm, setShowOpenConfirm] = useState(false);
  const [closeReason, setCloseReason] = useState('');

  const serviceOptions = [
    { value: 'wash', label: 'Мойка' },
    { value: 'air_dry', label: 'Обдув воздухом' },
    { value: 'vacuum', label: 'Пылесос' }
  ];

  const loadSettings = async () => {
    setLoading(true);
    setError('');

    try {
      const response = await ApiService.getSettings(selectedService);
      setSettings({
        price_per_minute: response.price_per_minute?.toString() || '',
        chemistry_price_per_minute: response.chemistry_price_per_minute?.toString() || '',
        available_rental_times: response.available_rental_times || [],
        available_chemistry_times: []
      });

      // Загружаем доступное время химии (только для wash)
      if (selectedService === 'wash') {
        try {
          const chemistryTimesResponse = await ApiService.getAdminAvailableChemistryTimes(selectedService);
          setSettings(prev => ({
            ...prev,
            available_chemistry_times: chemistryTimesResponse.available_chemistry_times || [3, 4, 5]
          }));
        } catch (chemistryTimesErr) {
          console.warn('Не удалось загрузить доступное время химии:', chemistryTimesErr);
          setSettings(prev => ({
            ...prev,
            available_chemistry_times: [3, 4, 5] // Значения по умолчанию
          }));
        }
      }
    } catch (err) {
      setError('Ошибка при загрузке настроек: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSettings();
    loadCleaningTimeout();
    loadSessionTimeout();
    loadCooldownTimeout();
    loadCarwashStatus();
  }, [selectedService]);

  const loadCarwashStatus = async () => {
    setCarwashStatusLoading(true);
    try {
      const response = await ApiService.getCarwashStatusAdmin();
      setCarwashStatus(response);
    } catch (err) {
      console.warn('Не удалось загрузить статус мойки:', err);
    } finally {
      setCarwashStatusLoading(false);
    }
  };

  const handleCloseCarwash = async () => {
    setCarwashActionLoading(true);
    setError('');
    setSuccess('');

    try {
      const response = await ApiService.closeCarwash(closeReason || null);
      setSuccess(`Мойка закрыта. Завершено сессий: ${response.completed_sessions}, отменено сессий: ${response.canceled_sessions}`);
      setShowCloseConfirm(false);
      setCloseReason('');
      await loadCarwashStatus();
    } catch (err) {
      setError('Ошибка при закрытии мойки: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setCarwashActionLoading(false);
    }
  };

  const handleOpenCarwash = async () => {
    setCarwashActionLoading(true);
    setError('');
    setSuccess('');

    try {
      await ApiService.openCarwash();
      setSuccess('Мойка открыта');
      setShowOpenConfirm(false);
      await loadCarwashStatus();
    } catch (err) {
      setError('Ошибка при открытии мойки: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setCarwashActionLoading(false);
    }
  };

  const loadCleaningTimeout = async () => {
    setCleaningTimeoutLoading(true);
    try {
      const response = await ApiService.getCleaningTimeout();
      setCleaningTimeout(response.timeout_minutes?.toString() || '');
    } catch (err) {
      console.warn('Не удалось загрузить время уборки:', err);
      setCleaningTimeout(''); // Значение по умолчанию
    } finally {
      setCleaningTimeoutLoading(false);
    }
  };

  const loadSessionTimeout = async () => {
    setSessionTimeoutLoading(true);
    try {
      const response = await ApiService.getSessionTimeout();
      setSessionTimeout(response.timeout_minutes?.toString() || '');
    } catch (err) {
      console.warn('Не удалось загрузить время ожидания старта мойки:', err);
      setSessionTimeout(''); // Значение по умолчанию
    } finally {
      setSessionTimeoutLoading(false);
    }
  };

  const loadCooldownTimeout = async () => {
    setCooldownTimeoutLoading(true);
    try {
      const response = await ApiService.getCooldownTimeout();
      setCooldownTimeout(response.timeout_minutes?.toString() || '');
    } catch (err) {
      console.warn('Не удалось загрузить время блокировки бокса:', err);
      setCooldownTimeout(''); // Значение по умолчанию
    } finally {
      setCooldownTimeoutLoading(false);
    }
  };

  const handleServiceChange = (serviceType) => {
    setSelectedService(serviceType);
    setSuccess('');
    setError('');
  };

  const handleCleaningTimeoutSave = async () => {
    // Проверяем, что поле не пустое
    if (!cleaningTimeout || cleaningTimeout.trim() === '') {
      setError('Время уборки не может быть пустым');
      return;
    }

    const timeoutValue = parseInt(cleaningTimeout);
    if (isNaN(timeoutValue) || timeoutValue < 1 || timeoutValue > 60) {
      setError('Время уборки должно быть числом от 1 до 60 минут');
      return;
    }

    setCleaningTimeoutLoading(true);
    setError('');

    try {
      await ApiService.updateCleaningTimeout(timeoutValue);
      setSuccess('Время уборки успешно обновлено');
    } catch (err) {
      setError('Ошибка при обновлении времени уборки: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setCleaningTimeoutLoading(false);
    }
  };

  const handleSessionTimeoutSave = async () => {
    // Проверяем, что поле не пустое
    if (!sessionTimeout || sessionTimeout.trim() === '') {
      setError('Время ожидания старта мойки не может быть пустым');
      return;
    }

    const timeoutValue = parseInt(sessionTimeout);
    if (isNaN(timeoutValue) || timeoutValue < 1 || timeoutValue > 60) {
      setError('Время ожидания старта мойки должно быть числом от 1 до 60 минут');
      return;
    }

    setSessionTimeoutLoading(true);
    setError('');

    try {
      await ApiService.updateSessionTimeout(timeoutValue);
      setSuccess('Время ожидания старта мойки успешно обновлено');
    } catch (err) {
      setError('Ошибка при обновлении времени ожидания старта мойки: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setSessionTimeoutLoading(false);
    }
  };

  const handleCooldownTimeoutSave = async () => {
    // Проверяем, что поле не пустое
    if (!cooldownTimeout || cooldownTimeout.trim() === '') {
      setError('Время блокировки бокса не может быть пустым');
      return;
    }

    const timeoutValue = parseInt(cooldownTimeout);
    if (isNaN(timeoutValue) || timeoutValue < 1 || timeoutValue > 60) {
      setError('Время блокировки бокса должно быть числом от 1 до 60 минут');
      return;
    }

    setCooldownTimeoutLoading(true);
    setError('');

    try {
      await ApiService.updateCooldownTimeout(timeoutValue);
      setSuccess('Время блокировки бокса успешно обновлено');
    } catch (err) {
      setError('Ошибка при обновлении времени блокировки бокса: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setCooldownTimeoutLoading(false);
    }
  };

  const handlePriceChange = (field, value) => {
    setSettings(prev => ({
      ...prev,
      [field]: value === '' ? '' : parseInt(value) || 0
    }));
  };

  const handleAddRentalTime = () => {
    const time = parseInt(newRentalTime);
    if (time && time > 0 && !settings.available_rental_times.includes(time)) {
      setSettings(prev => ({
        ...prev,
        available_rental_times: [...prev.available_rental_times, time].sort((a, b) => a - b)
      }));
      setNewRentalTime('');
    }
  };

  const handleRemoveRentalTime = (time) => {
    setSettings(prev => ({
      ...prev,
      available_rental_times: prev.available_rental_times.filter(t => t !== time)
    }));
  };

  const handleSavePrices = async () => {
    // Проверяем, что поля не пустые
    if (settings.price_per_minute === '' || settings.chemistry_price_per_minute === '') {
      setError('Все поля должны быть заполнены');
      return;
    }

    const pricePerMinute = parseInt(settings.price_per_minute);
    const chemistryPricePerMinute = parseInt(settings.chemistry_price_per_minute);

    if (isNaN(pricePerMinute) || isNaN(chemistryPricePerMinute) || pricePerMinute < 0 || chemistryPricePerMinute < 0) {
      setError('Цены должны быть положительными числами');
      return;
    }

    setSaving(true);
    setError('');
    setSuccess('');

    try {
      await ApiService.updatePrices({
        serviceType: selectedService,
        pricePerMinute: pricePerMinute,
        chemistryPricePerMinute: chemistryPricePerMinute
      });
      setSuccess('Цены успешно обновлены');
    } catch (err) {
      setError('Ошибка при обновлении цен: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setSaving(false);
    }
  };

  const handleSaveRentalTimes = async () => {
    setSaving(true);
    setError('');
    setSuccess('');

    try {
      await ApiService.updateRentalTimes({
        serviceType: selectedService,
        availableRentalTimes: settings.available_rental_times
      });
      setSuccess('Время мойки успешно обновлено');
    } catch (err) {
      setError('Ошибка при обновлении времени мойки: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setSaving(false);
    }
  };

  const handleAddChemistryTime = () => {
    const time = parseInt(newChemistryTime);
    if (time && time > 0 && !settings.available_chemistry_times.includes(time)) {
      setSettings(prev => ({
        ...prev,
        available_chemistry_times: [...prev.available_chemistry_times, time].sort((a, b) => a - b)
      }));
      setNewChemistryTime('');
    }
  };

  const handleRemoveChemistryTime = (time) => {
    setSettings(prev => ({
      ...prev,
      available_chemistry_times: prev.available_chemistry_times.filter(t => t !== time)
    }));
  };

  const handleSaveChemistryTimes = async () => {
    setSaving(true);
    setError('');
    setSuccess('');

    try {
      await ApiService.updateAvailableChemistryTimes({
        serviceType: selectedService,
        availableChemistryTimes: settings.available_chemistry_times
      });
      setSuccess('Доступное время химии успешно обновлено');
    } catch (err) {
      setError('Ошибка при обновлении времени химии: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setSaving(false);
    }
  };

  const formatPrice = (price) => {
    return `${(price / 100).toFixed(2)} руб.`;
  };

  return (
    <Container>
      <Title theme={theme}>Управление настройками</Title>

      {error && <ErrorMessage>{error}</ErrorMessage>}
      {success && <SuccessMessage>{success}</SuccessMessage>}

      {/* Кнопка управления статусом мойки */}
      <CarwashStatusContainer theme={theme} isClosed={carwashStatus?.is_closed}>
        <CarwashStatusTitle theme={theme}>
          <StatusIndicator>{carwashStatus?.is_closed ? '🔴' : '🟢'}</StatusIndicator>
          Статус мойки: {carwashStatus?.is_closed ? 'Закрыта' : 'Открыта'}
        </CarwashStatusTitle>
        
        {carwashStatus?.is_closed && carwashStatus?.closed_reason && (
          <WarningText>
            <strong>Причина закрытия:</strong> {carwashStatus.closed_reason}
          </WarningText>
        )}

        {!carwashStatus?.is_closed && (
          <WarningText>
            <strong>Внимание!</strong> При закрытии мойки будут выполнены следующие действия:
            <ul style={{ margin: '10px 0', paddingLeft: '20px' }}>
              <li>Все активные сессии будут завершены (выключится свет и химия в боксах)</li>
              <li>Все сессии в очереди/назначенные и созданные будут отменены с возвратом денег</li>
              <li>В mini app пользователи увидят сообщение о технических проблемах</li>
            </ul>
          </WarningText>
        )}

        <ButtonGroup>
          {carwashStatus?.is_closed ? (
            <SuccessButtonStyled
              theme={theme}
              onClick={() => setShowOpenConfirm(true)}
              disabled={carwashActionLoading || carwashStatusLoading}
            >
              {carwashActionLoading ? 'Открытие...' : 'Открыть мойку'}
            </SuccessButtonStyled>
          ) : (
            <DangerButton
              theme={theme}
              onClick={() => setShowCloseConfirm(true)}
              disabled={carwashActionLoading || carwashStatusLoading}
            >
              {carwashActionLoading ? 'Закрытие...' : 'Закрыть мойку'}
            </DangerButton>
          )}
        </ButtonGroup>
      </CarwashStatusContainer>

      {/* Модальное окно подтверждения закрытия */}
      {showCloseConfirm && (
        <ModalOverlay onClick={() => !carwashActionLoading && setShowCloseConfirm(false)}>
          <ModalContent theme={theme} onClick={(e) => e.stopPropagation()}>
            <ModalTitle theme={theme}>Подтверждение закрытия мойки</ModalTitle>
            <ModalText theme={theme}>
              Вы уверены, что хотите закрыть мойку? Это действие приведет к:
            </ModalText>
            <ul style={{ margin: '0 0 20px 0', paddingLeft: '20px', color: theme.textColor }}>
              <li>Завершению всех активных сессий (выключится свет и химия)</li>
              <li>Отмене всех сессий в очереди/назначенных/созданных с возвратом денег</li>
              <li>Скрытию всех кнопок в mini app для пользователей</li>
            </ul>
            <FormLabel theme={theme}>Причина закрытия (опционально):</FormLabel>
            <ModalTextarea
              theme={theme}
              value={closeReason}
              onChange={(e) => setCloseReason(e.target.value)}
              placeholder="Например: Отключена вода"
            />
            <ModalButtonGroup>
              <Button
                theme={theme}
                onClick={() => {
                  setShowCloseConfirm(false);
                  setCloseReason('');
                }}
                disabled={carwashActionLoading}
                style={{ backgroundColor: '#6c757d' }}
              >
                Отмена
              </Button>
              <DangerButton
                theme={theme}
                onClick={handleCloseCarwash}
                disabled={carwashActionLoading}
              >
                {carwashActionLoading ? 'Закрытие...' : 'Закрыть мойку'}
              </DangerButton>
            </ModalButtonGroup>
          </ModalContent>
        </ModalOverlay>
      )}

      {/* Модальное окно подтверждения открытия */}
      {showOpenConfirm && (
        <ModalOverlay onClick={() => !carwashActionLoading && setShowOpenConfirm(false)}>
          <ModalContent theme={theme} onClick={(e) => e.stopPropagation()}>
            <ModalTitle theme={theme}>Подтверждение открытия мойки</ModalTitle>
            <ModalText theme={theme}>
              Вы уверены, что хотите открыть мойку? После открытия пользователи снова смогут создавать сессии.
            </ModalText>
            <ModalButtonGroup>
              <Button
                theme={theme}
                onClick={() => setShowOpenConfirm(false)}
                disabled={carwashActionLoading}
                style={{ backgroundColor: '#6c757d' }}
              >
                Отмена
              </Button>
              <SuccessButtonStyled
                theme={theme}
                onClick={handleOpenCarwash}
                disabled={carwashActionLoading}
              >
                {carwashActionLoading ? 'Открытие...' : 'Открыть мойку'}
              </SuccessButtonStyled>
            </ModalButtonGroup>
          </ModalContent>
        </ModalOverlay>
      )}

      <ServiceSelector>
        <ServiceLabel theme={theme}>Выберите тип услуги:</ServiceLabel>
        <ServiceSelect
          theme={theme}
          value={selectedService}
          onChange={(e) => handleServiceChange(e.target.value)}
        >
          {serviceOptions.map(option => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </ServiceSelect>
      </ServiceSelector>

      {loading ? (
        <LoadingSpinner theme={theme}>Загрузка настроек...</LoadingSpinner>
      ) : (
        <>
          <SettingsContainer theme={theme}>
            <SettingsTitle theme={theme}>Настройки цен</SettingsTitle>
            
            <FormGrid>
              <FormGroup>
                <FormLabel theme={theme}>Цена за минуту (в копейках)</FormLabel>
                <FormInput
                  theme={theme}
                  type="number"
                  value={settings.price_per_minute}
                  onChange={(e) => handlePriceChange('price_per_minute', e.target.value)}
                  placeholder="Например: 100"
                />
                <small style={{ color: theme.textColor, opacity: 0.7 }}>
                  Текущая цена: {formatPrice(settings.price_per_minute)}
                </small>
              </FormGroup>

              {selectedService === 'wash' && (
                <FormGroup>
                  <FormLabel theme={theme}>Цена химии за минуту (в копейках)</FormLabel>
                  <FormInput
                    theme={theme}
                    type="number"
                    value={settings.chemistry_price_per_minute}
                    onChange={(e) => handlePriceChange('chemistry_price_per_minute', e.target.value)}
                    placeholder="Например: 50"
                  />
                  <small style={{ color: theme.textColor, opacity: 0.7 }}>
                    Текущая цена: {formatPrice(settings.chemistry_price_per_minute)}
                  </small>
              </FormGroup>
              )}
            </FormGrid>

            <ButtonGroup>
              <Button theme={theme} onClick={handleSavePrices} disabled={saving}>
                {saving ? 'Сохранение...' : 'Сохранить цены'}
              </Button>
            </ButtonGroup>
          </SettingsContainer>

          <SettingsContainer theme={theme}>
            <SettingsTitle theme={theme}>Доступное время мойки</SettingsTitle>
            
            <RentalTimesContainer>
              <RentalTimesTitle theme={theme}>Текущее время мойки (в минутах):</RentalTimesTitle>
              
              <RentalTimesList>
                {settings.available_rental_times.map(time => (
                  <RentalTimeItem key={time} theme={theme}>
                    {time} мин
                    <RemoveButton onClick={() => handleRemoveRentalTime(time)}>
                      ×
                    </RemoveButton>
                  </RentalTimeItem>
                ))}
              </RentalTimesList>

              <AddRentalTimeContainer>
                <AddRentalTimeInput
                  theme={theme}
                  type="number"
                  value={newRentalTime}
                  onChange={(e) => setNewRentalTime(e.target.value)}
                  placeholder="Время"
                  min="1"
                />
                <Button
                  theme={theme}
                  onClick={handleAddRentalTime}
                  disabled={!newRentalTime || parseInt(newRentalTime) <= 0}
                >
                  Добавить
                </Button>
              </AddRentalTimeContainer>
            </RentalTimesContainer>

            <ButtonGroup>
              <Button theme={theme} onClick={handleSaveRentalTimes} disabled={saving}>
                {saving ? 'Сохранение...' : 'Сохранить время мойки'}
              </Button>
            </ButtonGroup>
          </SettingsContainer>

          {selectedService === 'wash' && (
            <SettingsContainer theme={theme}>
              <SettingsTitle theme={theme}>Настройки химии</SettingsTitle>
              
              <RentalTimesContainer>
                <RentalTimesTitle theme={theme}>Доступное время химии (в минутах):</RentalTimesTitle>
                
                <RentalTimesList>
                  {settings.available_chemistry_times.map(time => (
                    <RentalTimeItem key={time} theme={theme}>
                      {time} мин
                      <RemoveButton onClick={() => handleRemoveChemistryTime(time)}>
                        ×
                      </RemoveButton>
                    </RentalTimeItem>
                  ))}
                </RentalTimesList>

                <AddRentalTimeContainer>
                  <AddRentalTimeInput
                    theme={theme}
                    type="number"
                    value={newChemistryTime}
                    onChange={(e) => setNewChemistryTime(e.target.value)}
                    placeholder="Минуты"
                    min="1"
                  />
                  <Button
                    theme={theme}
                    onClick={handleAddChemistryTime}
                    disabled={!newChemistryTime || parseInt(newChemistryTime) <= 0}
                  >
                    Добавить
                  </Button>
                </AddRentalTimeContainer>
                
                <small style={{ color: theme.textColor, opacity: 0.7, marginTop: '10px', display: 'block' }}>
                  Пользователь сможет выбрать одно из этих значений при оплате химии
                </small>
              </RentalTimesContainer>

              <ButtonGroup>
                <Button theme={theme} onClick={handleSaveChemistryTimes} disabled={saving}>
                  {saving ? 'Сохранение...' : 'Сохранить время химии'}
                </Button>
              </ButtonGroup>
            </SettingsContainer>
          )}

          <SettingsContainer theme={theme}>
            <SettingsTitle theme={theme}>Настройки уборки</SettingsTitle>
            
            <SettingsField>
              <SettingsLabel theme={theme}>Время уборки (в минутах):</SettingsLabel>
              <SettingsInput
                theme={theme}
                type="number"
                value={cleaningTimeout}
                onChange={(e) => setCleaningTimeout(e.target.value)}
                min="1"
                max="60"
                disabled={cleaningTimeoutLoading}
              />
              <SettingsDescription theme={theme}>
                Время, через которое уборка автоматически завершится (от 1 до 60 минут)
              </SettingsDescription>
            </SettingsField>

            <ButtonGroup>
              <Button 
                theme={theme} 
                onClick={handleCleaningTimeoutSave} 
                disabled={cleaningTimeoutLoading}
              >
                {cleaningTimeoutLoading ? 'Сохранение...' : 'Сохранить время уборки'}
              </Button>
            </ButtonGroup>
          </SettingsContainer>

          <SettingsContainer theme={theme}>
            <SettingsTitle theme={theme}>Настройки сессий</SettingsTitle>
            
            <SettingsField>
              <SettingsLabel theme={theme}>Время ожидания старта мойки (в минутах):</SettingsLabel>
              <SettingsInput
                theme={theme}
                type="number"
                value={sessionTimeout}
                onChange={(e) => setSessionTimeout(e.target.value)}
                min="1"
                max="60"
                disabled={sessionTimeoutLoading}
              />
              <SettingsDescription theme={theme}>
                Время, которое пользователь имеет для старта мойки после назначения бокса (от 1 до 60 минут)
              </SettingsDescription>
            </SettingsField>

            <ButtonGroup>
              <Button 
                theme={theme} 
                onClick={handleSessionTimeoutSave} 
                disabled={sessionTimeoutLoading}
              >
                {sessionTimeoutLoading ? 'Сохранение...' : 'Сохранить время ожидания старта мойки'}
              </Button>
            </ButtonGroup>
          </SettingsContainer>

          <SettingsContainer theme={theme}>
            <SettingsTitle theme={theme}>Настройки блокировки боксов</SettingsTitle>
            
            <SettingsField>
              <SettingsLabel theme={theme}>Время блокировки бокса после завершения сессии по времени (в минутах):</SettingsLabel>
              <SettingsInput
                theme={theme}
                type="number"
                value={cooldownTimeout}
                onChange={(e) => setCooldownTimeout(e.target.value)}
                min="1"
                max="60"
                disabled={cooldownTimeoutLoading}
              />
              <SettingsDescription theme={theme}>
                Время, на которое бокс блокируется для других пользователей после завершения сессии по времени (от 1 до 60 минут)
              </SettingsDescription>
            </SettingsField>

            <ButtonGroup>
              <Button 
                theme={theme} 
                onClick={handleCooldownTimeoutSave} 
                disabled={cooldownTimeoutLoading}
              >
                {cooldownTimeoutLoading ? 'Сохранение...' : 'Сохранить время блокировки бокса'}
              </Button>
            </ButtonGroup>
          </SettingsContainer>
        </>
      )}
    </Container>
  );
};

export default SettingsManagement; 