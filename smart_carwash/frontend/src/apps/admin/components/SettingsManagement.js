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

const SettingsManagement = () => {
  const theme = getTheme('light');
  const [selectedService, setSelectedService] = useState('wash');
  const [settings, setSettings] = useState({
    price_per_minute: 0,
    chemistry_price_per_minute: 0,
    available_rental_times: []
  });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [newRentalTime, setNewRentalTime] = useState('');
  const [chemistryTimeout, setChemistryTimeout] = useState(10);

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
        price_per_minute: response.price_per_minute || 0,
        chemistry_price_per_minute: response.chemistry_price_per_minute || 0,
        available_rental_times: response.available_rental_times || []
      });

      // Загружаем настройки химии
      try {
        const chemistryResponse = await ApiService.getChemistryTimeout(selectedService);
        setChemistryTimeout(chemistryResponse.chemistry_enable_timeout_minutes || 10);
      } catch (chemistryErr) {
        console.warn('Не удалось загрузить настройки химии:', chemistryErr);
        setChemistryTimeout(10); // Значение по умолчанию
      }
    } catch (err) {
      setError('Ошибка при загрузке настроек: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSettings();
  }, [selectedService]);

  const handleServiceChange = (serviceType) => {
    setSelectedService(serviceType);
    setSuccess('');
    setError('');
  };

  const handlePriceChange = (field, value) => {
    setSettings(prev => ({
      ...prev,
      [field]: parseInt(value) || 0
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
    setSaving(true);
    setError('');
    setSuccess('');

    try {
      await ApiService.updatePrices({
        serviceType: selectedService,
        pricePerMinute: settings.price_per_minute,
        chemistryPricePerMinute: settings.chemistry_price_per_minute
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
      setSuccess('Время аренды успешно обновлено');
    } catch (err) {
      setError('Ошибка при обновлении времени аренды: ' + (err.message || 'Неизвестная ошибка'));
    } finally {
      setSaving(false);
    }
  };

  const handleSaveChemistryTimeout = async () => {
    setSaving(true);
    setError('');
    setSuccess('');

    try {
      await ApiService.updateChemistryTimeout(selectedService, parseInt(chemistryTimeout));
      setSuccess('Настройки химии успешно обновлены');
    } catch (err) {
      setError('Ошибка при обновлении настроек химии: ' + (err.message || 'Неизвестная ошибка'));
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
            <SettingsTitle theme={theme}>Доступное время аренды</SettingsTitle>
            
            <RentalTimesContainer>
              <RentalTimesTitle theme={theme}>Текущее время аренды (в минутах):</RentalTimesTitle>
              
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
                {saving ? 'Сохранение...' : 'Сохранить время аренды'}
              </Button>
            </ButtonGroup>
          </SettingsContainer>

          {selectedService === 'wash' && (
            <SettingsContainer theme={theme}>
              <SettingsTitle theme={theme}>Настройки химии</SettingsTitle>
              
              <FormGrid>
                <FormGroup>
                  <FormLabel theme={theme}>Время доступности кнопки химии (в минутах)</FormLabel>
                  <FormInput
                    theme={theme}
                    type="number"
                    value={chemistryTimeout}
                    onChange={(e) => setChemistryTimeout(e.target.value)}
                    placeholder="Например: 10"
                    min="1"
                    max="60"
                  />
                  <small style={{ color: theme.textColor, opacity: 0.7 }}>
                    Время, в течение которого пользователь может включить химию после старта мойки
                  </small>
                </FormGroup>
              </FormGrid>

              <ButtonGroup>
                <Button theme={theme} onClick={handleSaveChemistryTimeout} disabled={saving}>
                  {saving ? 'Сохранение...' : 'Сохранить настройки химии'}
                </Button>
              </ButtonGroup>
            </SettingsContainer>
          )}
        </>
      )}
    </Container>
  );
};

export default SettingsManagement; 