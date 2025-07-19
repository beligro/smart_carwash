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

const SettingsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
`;

const SettingsCard = styled.div`
  background-color: ${props => props.theme.cardBackground};
  padding: 25px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
`;

const CardTitle = styled.h3`
  margin: 0 0 20px 0;
  color: ${props => props.theme.primaryColor};
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  gap: 10px;
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
  font-size: 14px;
`;

const Input = styled.input`
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-size: 14px;
  background-color: white;
  
  &:focus {
    outline: none;
    border-color: ${props => props.theme.primaryColor};
  }
`;

const Button = styled.button`
  background-color: ${props => props.theme.primaryColor};
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  
  &:hover {
    background-color: ${props => props.theme.primaryColorDark};
  }
  
  &:disabled {
    background-color: #ccc;
    cursor: not-allowed;
  }
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

const PriceInfo = styled.div`
  background-color: #f5f5f5;
  padding: 15px;
  border-radius: 5px;
  margin-bottom: 15px;
`;

const PriceInfoTitle = styled.h4`
  margin: 0 0 10px 0;
  color: ${props => props.theme.textColor};
  font-size: 14px;
`;

const PriceInfoText = styled.p`
  margin: 0;
  color: ${props => props.theme.textColor};
  font-size: 12px;
  line-height: 1.4;
`;

/**
 * Компонент управления настройками системы
 * @returns {React.ReactNode} - Компонент управления настройками
 */
const SettingsManagement = () => {
  const theme = getTheme('light');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  
  // Состояние для настроек системы
  const [settings, setSettings] = useState({
    prices: {
      wash: {
        price_per_minute: 1000,
        chemistry_price_per_minute: 200
      },
      air_dry: {
        price_per_minute: 600,
        chemistry_price_per_minute: 100
      },
      vacuum: {
        price_per_minute: 400,
        chemistry_price_per_minute: 50
      }
    },
    rentalTimes: {
      wash: [5, 10, 15, 20],
      air_dry: [5, 10, 15],
      vacuum: [5, 10]
    },
    systemSettings: {
      max_queue_size: 10,
      session_timeout_minutes: 5,
      reservation_timeout_minutes: 3,
      notification_enabled: true
    }
  });

  // Состояние для промежуточного ввода времени аренды
  const [rentalTimeInputs, setRentalTimeInputs] = useState({
    wash: '5, 10, 15, 20',
    air_dry: '5, 10, 15',
    vacuum: '5, 10'
  });

  // Загрузка текущих настроек при монтировании компонента
  useEffect(() => {
    loadSettings();
  }, []);

  // Функция для загрузки настроек
  const loadSettings = async () => {
    try {
      setLoading(true);
      setError('');
      
      // Загружаем все настройки одной ручкой
      const allSettings = await ApiService.getAllSettings();
      
      const rentalTimes = allSettings.rental_times || {
        wash: [5, 10, 15, 20],
        air_dry: [5, 10, 15],
        vacuum: [5, 10]
      };

      setSettings({
        prices: allSettings.prices || {
          wash: { price_per_minute: 1000, chemistry_price_per_minute: 200 },
          air_dry: { price_per_minute: 600, chemistry_price_per_minute: 100 },
          vacuum: { price_per_minute: 400, chemistry_price_per_minute: 50 }
        },
        rentalTimes: rentalTimes,
        systemSettings: {
          max_queue_size: allSettings.system_settings?.max_queue_size || 10,
          session_timeout_minutes: allSettings.system_settings?.session_timeout_minutes || 5,
          reservation_timeout_minutes: allSettings.system_settings?.reservation_timeout_minutes || 3,
          notification_enabled: allSettings.system_settings?.notification_enabled !== undefined ? 
            allSettings.system_settings.notification_enabled === 1 : true
        }
      });

      // Инициализируем поля ввода времени аренды
      setRentalTimeInputs({
        wash: rentalTimes.wash.join(', '),
        air_dry: rentalTimes.air_dry.join(', '),
        vacuum: rentalTimes.vacuum.join(', ')
      });
    } catch (err) {
      setError('Ошибка при загрузке настроек');
      console.error('Error loading settings:', err);
    } finally {
      setLoading(false);
    }
  };

  // Обработчик изменения цены
  const handlePriceChange = (serviceType, priceType, value) => {
    setSettings(prev => ({
      ...prev,
      prices: {
        ...prev.prices,
        [serviceType]: {
          ...prev.prices[serviceType],
          [priceType]: parseInt(value) || 0
        }
      }
    }));
  };

  // Обработчик изменения времени аренды (промежуточный ввод)
  const handleRentalTimeInputChange = (serviceType, value) => {
    setRentalTimeInputs(prev => ({
      ...prev,
      [serviceType]: value
    }));
  };

  // Обработчик завершения ввода времени аренды (при потере фокуса или нажатии Enter)
  const handleRentalTimeInputBlur = (serviceType) => {
    const inputValue = rentalTimeInputs[serviceType];
    
    // Парсим введенные значения
    const times = inputValue
      .split(',')
      .map(t => t.trim())
      .filter(t => t !== '')
      .map(t => parseInt(t))
      .filter(t => !isNaN(t) && t > 0);
    
    // Обновляем настройки только если есть валидные значения
    if (times.length > 0) {
      setSettings(prev => ({
        ...prev,
        rentalTimes: {
          ...prev.rentalTimes,
          [serviceType]: times
        }
      }));
      
      // Обновляем отображаемый текст
      setRentalTimeInputs(prev => ({
        ...prev,
        [serviceType]: times.join(', ')
      }));
    } else {
      // Если нет валидных значений, возвращаем к предыдущему состоянию
      setRentalTimeInputs(prev => ({
        ...prev,
        [serviceType]: settings.rentalTimes[serviceType].join(', ')
      }));
    }
  };

  // Обработчик нажатия клавиш в поле ввода времени аренды
  const handleRentalTimeInputKeyDown = (serviceType, e) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleRentalTimeInputBlur(serviceType);
    }
  };

  // Обработчик изменения системной настройки
  const handleSystemSettingChange = (settingKey, value) => {
    setSettings(prev => ({
      ...prev,
      systemSettings: {
        ...prev.systemSettings,
        [settingKey]: typeof value === 'boolean' ? value : parseInt(value) || 0
      }
    }));
  };

  // Обработчик сохранения настроек цен
  const handleSavePriceSettings = async (serviceType) => {
    try {
      setError('');
      setSuccess('');
      
      const servicePrices = settings.prices[serviceType];
      
      // Сохраняем цены для данного типа услуги
      await ApiService.updateSetting(serviceType, 'price_per_minute', servicePrices.price_per_minute);
      await ApiService.updateSetting(serviceType, 'chemistry_price_per_minute', servicePrices.chemistry_price_per_minute);
      
      setSuccess(`Цены для ${getServiceTypeName(serviceType)} успешно обновлены`);
      
      // Очищаем сообщение об успехе через 3 секунды
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError(`Ошибка при сохранении цен для ${getServiceTypeName(serviceType)}`);
      console.error('Error saving price settings:', err);
    }
  };

  // Обработчик сохранения настроек времени аренды
  const handleSaveRentalTimeSettings = async (serviceType) => {
    try {
      setError('');
      setSuccess('');
      
      const rentalTimes = settings.rentalTimes[serviceType];
      
      // Сохраняем время аренды для данного типа услуги
      await ApiService.updateAvailableRentalTimes(serviceType, rentalTimes);
      
      setSuccess(`Время аренды для ${getServiceTypeName(serviceType)} успешно обновлено`);
      
      // Очищаем сообщение об успехе через 3 секунды
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError(`Ошибка при сохранении времени аренды для ${getServiceTypeName(serviceType)}`);
      console.error('Error saving rental time settings:', err);
    }
  };

  // Обработчик сохранения системных настроек
  const handleSaveSystemSettings = async () => {
    try {
      setError('');
      setSuccess('');
      
      const systemSettings = settings.systemSettings;
      
      // Сохраняем системные настройки
      await ApiService.updateSetting('system', 'max_queue_size', systemSettings.max_queue_size);
      await ApiService.updateSetting('system', 'session_timeout_minutes', systemSettings.session_timeout_minutes);
      await ApiService.updateSetting('system', 'reservation_timeout_minutes', systemSettings.reservation_timeout_minutes);
      await ApiService.updateSetting('system', 'notification_enabled', systemSettings.notification_enabled ? 1 : 0);
      
      setSuccess('Системные настройки успешно обновлены');
      
      // Очищаем сообщение об успехе через 3 секунды
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError('Ошибка при сохранении системных настроек');
      console.error('Error saving system settings:', err);
    }
  };

  // Функция для получения названия типа услуги
  const getServiceTypeName = (serviceType) => {
    const names = {
      wash: 'Мойка',
      air_dry: 'Обдув',
      vacuum: 'Пылесос'
    };
    return names[serviceType] || serviceType;
  };

  // Функция для получения иконки типа услуги
  const getServiceIcon = (serviceType) => {
    const icons = {
      wash: '🚗',
      air_dry: '💨',
      vacuum: '🧹'
    };
    return icons[serviceType] || '⚙️';
  };

  if (loading) {
    return <LoadingMessage theme={theme}>Загрузка настроек...</LoadingMessage>;
  }

  return (
    <Container>
      <Header>
        <Title theme={theme}>Управление настройками системы</Title>
      </Header>

      {error && <ErrorMessage>{error}</ErrorMessage>}
      {success && <SuccessMessage>{success}</SuccessMessage>}

      <SettingsGrid>
        {/* Настройки цен */}
        <SettingsCard theme={theme}>
          <CardTitle theme={theme}>
            💰 Настройки цен
          </CardTitle>
          
          <PriceInfo theme={theme}>
            <PriceInfoTitle theme={theme}>Информация о ценах</PriceInfoTitle>
            <PriceInfoText theme={theme}>
              Цены указаны в копейках за минуту. Химия добавляется к базовой цене при выборе соответствующей опции.
            </PriceInfoText>
          </PriceInfo>

          {Object.keys(settings.prices).map(serviceType => (
            <div key={serviceType} style={{ marginBottom: '20px', padding: '15px', border: '1px solid #eee', borderRadius: '5px' }}>
              <h4 style={{ margin: '0 0 10px 0', color: theme.textColor }}>
                {getServiceIcon(serviceType)} {getServiceTypeName(serviceType)}
              </h4>
              
              <FormGroup>
                <Label theme={theme}>Базовая цена за минуту (коп.)</Label>
                <Input
                  type="number"
                  value={settings.prices[serviceType].price_per_minute}
                  onChange={(e) => handlePriceChange(serviceType, 'price_per_minute', e.target.value)}
                  min="0"
                  step="100"
                />
              </FormGroup>

              <FormGroup>
                <Label theme={theme}>Цена химии за минуту (коп.)</Label>
                <Input
                  type="number"
                  value={settings.prices[serviceType].chemistry_price_per_minute}
                  onChange={(e) => handlePriceChange(serviceType, 'chemistry_price_per_minute', e.target.value)}
                  min="0"
                  step="50"
                />
              </FormGroup>

              <Button 
                onClick={() => handleSavePriceSettings(serviceType)}
                theme={theme}
                style={{ marginTop: '10px' }}
              >
                Сохранить цены для {getServiceTypeName(serviceType)}
              </Button>
            </div>
          ))}
        </SettingsCard>

        {/* Настройки времени аренды */}
        <SettingsCard theme={theme}>
          <CardTitle theme={theme}>
            ⏱️ Время аренды
          </CardTitle>
          
          <PriceInfo theme={theme}>
            <PriceInfoTitle theme={theme}>Информация о времени аренды</PriceInfoTitle>
            <PriceInfoText theme={theme}>
              Укажите доступные варианты времени аренды в минутах через запятую (например: 5, 10, 15, 20).
            </PriceInfoText>
          </PriceInfo>

          {Object.keys(settings.rentalTimes).map(serviceType => (
            <div key={serviceType} style={{ marginBottom: '20px', padding: '15px', border: '1px solid #eee', borderRadius: '5px' }}>
              <h4 style={{ margin: '0 0 10px 0', color: theme.textColor }}>
                {getServiceIcon(serviceType)} {getServiceTypeName(serviceType)}
              </h4>
              
              <FormGroup>
                <Label theme={theme}>Доступное время (минуты)</Label>
                <Input
                  type="text"
                  value={rentalTimeInputs[serviceType]}
                  onChange={(e) => handleRentalTimeInputChange(serviceType, e.target.value)}
                  onBlur={() => handleRentalTimeInputBlur(serviceType)}
                  onKeyDown={(e) => handleRentalTimeInputKeyDown(serviceType, e)}
                  placeholder="5, 10, 15, 20"
                />
                <small style={{ color: '#666', fontSize: '12px', marginTop: '5px' }}>
                  Введите числа через запятую (например: 5, 10, 15, 20). Нажмите Enter, Tab или кликните вне поля для применения.
                </small>
              </FormGroup>

              <Button 
                onClick={() => handleSaveRentalTimeSettings(serviceType)}
                theme={theme}
                style={{ marginTop: '10px' }}
              >
                Сохранить время для {getServiceTypeName(serviceType)}
              </Button>
            </div>
          ))}
        </SettingsCard>

        {/* Системные настройки */}
        <SettingsCard theme={theme}>
          <CardTitle theme={theme}>
            ⚙️ Системные настройки
          </CardTitle>
          
          <PriceInfo theme={theme}>
            <PriceInfoTitle theme={theme}>Информация о системных настройках</PriceInfoTitle>
            <PriceInfoText theme={theme}>
              Настройки, влияющие на работу всей системы автомойки.
            </PriceInfoText>
          </PriceInfo>

          <Form>
            <FormGroup>
              <Label theme={theme}>Максимальный размер очереди</Label>
              <Input
                type="number"
                value={settings.systemSettings.max_queue_size}
                onChange={(e) => handleSystemSettingChange('max_queue_size', e.target.value)}
                min="1"
                max="100"
              />
            </FormGroup>

            <FormGroup>
              <Label theme={theme}>Таймаут сессии (минуты)</Label>
              <Input
                type="number"
                value={settings.systemSettings.session_timeout_minutes}
                onChange={(e) => handleSystemSettingChange('session_timeout_minutes', e.target.value)}
                min="1"
                max="60"
              />
            </FormGroup>

            <FormGroup>
              <Label theme={theme}>Таймаут резервации (минуты)</Label>
              <Input
                type="number"
                value={settings.systemSettings.reservation_timeout_minutes}
                onChange={(e) => handleSystemSettingChange('reservation_timeout_minutes', e.target.value)}
                min="1"
                max="10"
              />
            </FormGroup>

            <FormGroup>
              <Label theme={theme}>
                <input
                  type="checkbox"
                  checked={settings.systemSettings.notification_enabled}
                  onChange={(e) => handleSystemSettingChange('notification_enabled', e.target.checked)}
                  style={{ marginRight: '8px' }}
                />
                Уведомления включены
              </Label>
            </FormGroup>

            <Button 
              onClick={handleSaveSystemSettings}
              theme={theme}
            >
              Сохранить системные настройки
            </Button>
          </Form>
        </SettingsCard>
      </SettingsGrid>
    </Container>
  );
};

export default SettingsManagement; 