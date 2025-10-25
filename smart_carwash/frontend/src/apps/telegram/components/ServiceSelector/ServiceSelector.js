import React, { useState, useEffect } from 'react';
import styles from './ServiceSelector.module.css';
import { Card, Button } from '../../../../shared/components/UI';
import CarNumberInput from '../CarNumberInput';
import EmailInput from '../EmailInput';
import PriceCalculator from '../PriceCalculator';
import ApiService from '../../../../shared/services/ApiService';

/**
 * Компонент ServiceSelector - позволяет выбрать тип услуги и дополнительные опции
 * @param {Object} props - Свойства компонента
 * @param {Function} props.onSelect - Функция, вызываемая при выборе услуги
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Object} props.user - Данные пользователя (для получения сохраненного номера)
 */
const ServiceSelector = ({ onSelect, theme = 'light', user }) => {
  const [selectedService, setSelectedService] = useState(null);
  const [withChemistry, setWithChemistry] = useState(false);
  const [rentalTimes, setRentalTimes] = useState([]);
  const [selectedRentalTime, setSelectedRentalTime] = useState(null);
  const [chemistryTimes, setChemistryTimes] = useState([]);
  const [selectedChemistryTime, setSelectedChemistryTime] = useState(null);
  const [filteredChemistryTimes, setFilteredChemistryTimes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [loadingChemistryTimes, setLoadingChemistryTimes] = useState(false);
  const [carNumber, setCarNumber] = useState('');
  const [rememberCarNumber, setRememberCarNumber] = useState(false);
  const [savingCarNumber, setSavingCarNumber] = useState(false);
  const [email, setEmail] = useState('');
  const [rememberEmail, setRememberEmail] = useState(false);
  const [savingEmail, setSavingEmail] = useState(false);
  const [wantReceipt, setWantReceipt] = useState(false);
  
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  // Инициализация номера машины и email из данных пользователя
  useEffect(() => {
    try {
      if (user) {
        if (user.car_number) {
          setCarNumber(user.car_number);
        }
        if (user.email) {
          setEmail(user.email);
        }
      }
    } catch (error) {
      console.error('Ошибка при инициализации данных пользователя:', error);
    }
  }, [user]);
  
  // Типы услуг
  const serviceTypes = [
    { id: 'wash', name: 'Мойка', description: 'Стандартная мойка автомобиля', hasChemistry: true },
    { id: 'vacuum', name: 'Пылеводосос', description: 'Уборка салона пылеводососом', hasChemistry: false },
    { id: 'air_dry', name: 'Воздух для продувки', description: 'Сушка автомобиля воздухом', hasChemistry: false }
  ];
  
  // Загрузка доступного времени мойки при выборе услуги
  useEffect(() => {
    try {
      if (selectedService) {
        setLoading(true);
        ApiService.getAvailableRentalTimes(selectedService.id)
          .then(data => {
            setRentalTimes(data.available_times || [5]);
            setSelectedRentalTime(data.available_times && data.available_times.length > 0 ? data.available_times[0] : 5);
          })
          .catch(error => {
            console.error('Ошибка при загрузке времени мойки:', error);
            setRentalTimes([5]);
            setSelectedRentalTime(5);
          })
          .finally(() => {
            setLoading(false);
          });
      } else {
        setRentalTimes([]);
        setSelectedRentalTime(null);
      }
    } catch (error) {
      console.error('Ошибка в useEffect для загрузки времени мойки:', error);
      setRentalTimes([5]);
      setSelectedRentalTime(5);
      setLoading(false);
    }
  }, [selectedService]);
  
  // Фильтрация времени химии по выбранному времени мойки
  useEffect(() => {
    if (chemistryTimes.length > 0 && selectedRentalTime) {
      const filtered = chemistryTimes.filter(time => time <= selectedRentalTime);
      setFilteredChemistryTimes(filtered);
      
      // Если текущее выбранное время химии больше выбранного времени мойки, сбрасываем выбор
      if (selectedChemistryTime && selectedChemistryTime > selectedRentalTime) {
        const firstAvailable = filtered.length > 0 ? filtered[0] : null;
        setSelectedChemistryTime(firstAvailable);
      }
    } else {
      setFilteredChemistryTimes(chemistryTimes);
    }
  }, [chemistryTimes, selectedRentalTime, selectedChemistryTime]);

  // Обработчик выбора услуги
  const handleServiceSelect = async (serviceType) => {
    try {
      setSelectedService(serviceType);
      // Если выбрана мойка, включаем химию по умолчанию
      if (serviceType.id === 'wash') {
        setWithChemistry(true);
        // Загружаем доступное время химии для мойки
        setLoadingChemistryTimes(true);
        try {
          const data = await ApiService.getAvailableChemistryTimes(serviceType.id);
          setChemistryTimes(data.available_chemistry_times || [3, 4, 5]);
          setSelectedChemistryTime(data.available_chemistry_times ? data.available_chemistry_times[0] : 3);
        } catch (error) {
          console.error('Ошибка при загрузке времени химии для мойки:', error);
          setChemistryTimes([3, 4, 5]);
          setSelectedChemistryTime(3);
        } finally {
          setLoadingChemistryTimes(false);
        }
      } else if (!serviceType.hasChemistry) {
        // Если выбрана услуга без химии, сбрасываем флаг
        setWithChemistry(false);
        setChemistryTimes([]);
        setSelectedChemistryTime(null);
      }
    } catch (error) {
      console.error('Ошибка в handleServiceSelect:', error);
    }
  };

  // Обработчик отмены выбора услуги
  const handleServiceDeselect = () => {
    try {
      setSelectedService(null);
      setWithChemistry(false);
      setChemistryTimes([]);
      setSelectedChemistryTime(null);
      setRentalTimes([]);
      setSelectedRentalTime(null);
    } catch (error) {
      console.error('Ошибка в handleServiceDeselect:', error);
    }
  };
  
  // Обработчик переключения опции химии
  const handleChemistryToggle = async () => {
    try {
      const newWithChemistry = !withChemistry;
      setWithChemistry(newWithChemistry);
      
      // Если включаем химию, загружаем доступное время
      if (newWithChemistry && selectedService) {
        setLoadingChemistryTimes(true);
        try {
          const data = await ApiService.getAvailableChemistryTimes(selectedService.id);
          setChemistryTimes(data.available_chemistry_times || [3, 4, 5]);
          setSelectedChemistryTime(data.available_chemistry_times ? data.available_chemistry_times[0] : 3);
        } catch (error) {
          console.error('Ошибка при загрузке времени химии:', error);
          setChemistryTimes([3, 4, 5]);
          setSelectedChemistryTime(3);
        } finally {
          setLoadingChemistryTimes(false);
        }
      } else {
        // Если выключаем химию, сбрасываем выбор
        setChemistryTimes([]);
        setSelectedChemistryTime(null);
      }
    } catch (error) {
      console.error('Ошибка в handleChemistryToggle:', error);
    }
  };
  
  // Обработчик выбора времени мойки
  const handleRentalTimeSelect = (time) => {
    try {
      setSelectedRentalTime(time);
    } catch (error) {
      console.error('Ошибка в handleRentalTimeSelect:', error);
    }
  };

  // Обработчик выбора времени химии
  const handleChemistryTimeSelect = (time) => {
    try {
      setSelectedChemistryTime(time);
    } catch (error) {
      console.error('Ошибка в handleChemistryTimeSelect:', error);
    }
  };

  // Обработчик изменения номера машины
  const handleCarNumberChange = (value) => {
    try {
      setCarNumber(value || '');
    } catch (error) {
      console.error('Ошибка в handleCarNumberChange:', error);
      setCarNumber('');
    }
  };

  // Обработчик изменения чекбокса "запомнить номер"
  const handleRememberCarNumberChange = (checked) => {
    try {
      setRememberCarNumber(checked);
    } catch (error) {
      console.error('Ошибка в handleRememberCarNumberChange:', error);
    }
  };

  // Обработчик изменения email
  const handleEmailChange = (value) => {
    try {
      setEmail(value || '');
    } catch (error) {
      console.error('Ошибка в handleEmailChange:', error);
      setEmail('');
    }
  };

  // Обработчик изменения чекбокса "запомнить email"
  const handleRememberEmailChange = (checked) => {
    try {
      setRememberEmail(checked);
    } catch (error) {
      console.error('Ошибка в handleRememberEmailChange:', error);
    }
  };

  // Обработчик изменения галочки "Хочу получить чек"
  const handleWantReceiptChange = (checked) => {
    try {
      setWantReceipt(checked);
      // Если галочка выключена, очищаем email
      if (!checked) {
        setEmail('');
      }
    } catch (error) {
      console.error('Ошибка в handleWantReceiptChange:', error);
      setWantReceipt(false);
    }
  };

  // Безопасная проверка номера машины (гибкая валидация)
  const isValidCarNumber = (number) => {
    try {
      if (!number || typeof number !== 'string') {
        return false;
      }
      
      // Проверяем минимальную длину
      if (number.length < 6) {
        return false;
      }
      
      // Проверяем, что номер содержит только буквы и цифры
      const carNumberRegex = /^[А-ЯA-Z0-9]+$/;
      if (!carNumberRegex.test(number)) {
        return false;
      }
      
      // Проверяем, что есть хотя бы одна буква и одна цифра
      const hasLetter = /[А-ЯA-Z]/.test(number);
      const hasDigit = /[0-9]/.test(number);
      
      return hasLetter && hasDigit;
    } catch (error) {
      console.error('Ошибка в isValidCarNumber:', error);
      return false;
    }
  };

  // Сохранение номера машины пользователя
  const saveCarNumber = async () => {
    if (!user || !carNumber || !rememberCarNumber) {
      return;
    }

    setSavingCarNumber(true);
    try {
      await ApiService.updateCarNumber(user.id, carNumber);
      console.log('Номер машины сохранен');
    } catch (error) {
      console.error('Ошибка при сохранении номера машины:', error);
    } finally {
      setSavingCarNumber(false);
    }
  };

  // Сохранение email пользователя
  const saveEmail = async () => {
    if (!user || !email || !rememberEmail) {
      return;
    }

    setSavingEmail(true);
    try {
      await ApiService.updateEmail(user.id, email);
      console.log('Email сохранен');
    } catch (error) {
      console.error('Ошибка при сохранении email:', error);
    } finally {
      setSavingEmail(false);
    }
  };

  // Обработчик подтверждения выбора
  const handleConfirm = async () => {
    try {
      if (selectedService && selectedRentalTime && carNumber) {
        // Если пользователь хочет запомнить номер, сохраняем его
        if (rememberCarNumber) {
          await saveCarNumber();
        }

        // Если пользователь хочет запомнить email, сохраняем его
        if (rememberEmail) {
          await saveEmail();
        }

        // Определяем время химии
        let chemistryTime = 0;
        if (selectedService.hasChemistry && withChemistry) {
          chemistryTime = selectedChemistryTime || 0;
          
          // Дополнительная проверка на случай если время не выбрано
          if (chemistryTime === 0) {
            alert('Пожалуйста, выберите время химии');
            return;
          }
        }

        const serviceData = {
          serviceType: selectedService.id,
          withChemistry: selectedService.hasChemistry ? withChemistry : false,
          chemistryTimeMinutes: chemistryTime,
          rentalTimeMinutes: selectedRentalTime,
          carNumber: carNumber,
          email: wantReceipt ? email : null // Передаем email только если галочка включена
        };
        
        onSelect(serviceData);
      }
    } catch (error) {
      alert('Ошибка в handleConfirm: ' + error.message);
    }
  };

  // Проверяем, можно ли подтвердить выбор
  const canConfirm = selectedService && 
    selectedRentalTime && 
    carNumber && 
    carNumber.length >= 6 && 
    isValidCarNumber(carNumber) &&
    (!withChemistry || selectedChemistryTime) && // Если химия включена, должно быть выбрано время
    (!wantReceipt || (email && email.length > 0)); // Если галочка включена, email должен быть заполнен

  // Определяем, нужно ли показывать чекбокс "запомнить номер"
  const showRememberCheckbox = user && 
    carNumber && 
    carNumber !== user.car_number && 
    isValidCarNumber(carNumber);
  
  return (
    <div className={styles.container}>
      <h2 className={`${styles.title} ${themeClass}`}>Выберите услугу</h2>
      
      <div className={styles.serviceGrid}>
        {selectedService ? (
          // Показываем только выбранную услугу с крестиком
          <Card 
            theme={theme} 
            className={`${styles.serviceCard} ${styles.selected} ${styles.singleSelected}`}
          >
            <div className={styles.selectedServiceContent}>
              <h3 className={`${styles.serviceName} ${themeClass}`}>{selectedService.name}</h3>
              <p className={`${styles.serviceDescription} ${themeClass}`}>{selectedService.description}</p>
            </div>
            <button 
              className={styles.deselectButton}
              onClick={(e) => {
                e.stopPropagation();
                handleServiceDeselect();
              }}
              title="Отменить выбор"
            >
              ✕
            </button>
          </Card>
        ) : (
          // Показываем все услуги для выбора
          serviceTypes.map((service) => (
            <Card 
              key={service.id} 
              theme={theme} 
              className={styles.serviceCard}
              onClick={() => handleServiceSelect(service)}
            >
              <h3 className={`${styles.serviceName} ${themeClass}`}>{service.name}</h3>
              <p className={`${styles.serviceDescription} ${themeClass}`}>{service.description}</p>
            </Card>
          ))
        )}
      </div>
      
      {selectedService && (
        <div className={styles.optionsContainer}>
          {/* Ввод номера машины */}
          <CarNumberInput
            value={carNumber || ''}
            onChange={handleCarNumberChange}
            theme={theme}
            showRememberCheckbox={showRememberCheckbox}
            rememberChecked={rememberCarNumber}
            onRememberChange={handleRememberCarNumberChange}
            savedCarNumber={user?.car_number || ''}
          />


          {selectedService.hasChemistry && (
            <>
              <Card theme={theme} className={styles.optionCard}>
                <div className={styles.optionRow}>
                  <label className={`${styles.optionLabel} ${themeClass}`}>
                    <input 
                      type="checkbox" 
                      checked={withChemistry} 
                      onChange={handleChemistryToggle}
                      className={styles.checkbox}
                    />
                    <span className={styles.checkmark}></span>
                    Использовать химию
                  </label>
                </div>
                <p className={`${styles.optionDescription} ${themeClass}`}>
                  Химия помогает лучше очистить поверхность автомобиля от грязи и жира
                </p>
              </Card>

              {withChemistry && (
                <Card theme={theme} className={styles.optionCard}>
                  <h3 className={`${styles.optionTitle} ${themeClass}`}>Выберите время химии</h3>
                  <p className={`${styles.optionDescription} ${themeClass}`}>
                    Химия будет автоматически выключена через выбранное время
                  </p>
                  {loadingChemistryTimes ? (
                    <p className={`${styles.loadingText} ${themeClass}`}>Загрузка доступного времени...</p>
                  ) : filteredChemistryTimes.length === 0 ? (
                    <p className={`${styles.optionDescription} ${themeClass}`}>
                      Нет доступных вариантов химии для выбранного времени мойки
                    </p>
                  ) : (
                    <div className={styles.rentalTimeGrid}>
                      {filteredChemistryTimes.map((time) => (
                        <div 
                          key={time} 
                          className={`${styles.rentalTimeItem} ${selectedChemistryTime === time ? styles.selectedTime : ''}`}
                          onClick={() => handleChemistryTimeSelect(time)}
                        >
                          <span className={`${styles.rentalTimeValue} ${themeClass}`}>{time}</span>
                          <span className={`${styles.rentalTimeUnit} ${themeClass}`}>мин</span>
                        </div>
                      ))}
                    </div>
                  )}
                </Card>
              )}
            </>
          )}
          
          <Card theme={theme} className={styles.optionCard}>
            <h3 className={`${styles.optionTitle} ${themeClass}`}>Выберите время мойки</h3>
            {loading ? (
              <p className={`${styles.loadingText} ${themeClass}`}>Загрузка доступного времени...</p>
            ) : (
              <div className={styles.rentalTimeGrid}>
                {(rentalTimes || []).map((time) => (
                  <div 
                    key={time} 
                    className={`${styles.rentalTimeItem} ${selectedRentalTime === time ? styles.selectedTime : ''}`}
                    onClick={() => handleRentalTimeSelect(time)}
                  >
                    <span className={`${styles.rentalTimeValue} ${themeClass}`}>{time}</span>
                    <span className={`${styles.rentalTimeUnit} ${themeClass}`}>мин</span>
                  </div>
                ))}
              </div>
            )}
          </Card>

          {/* Отображение цены */}
          {selectedService && selectedRentalTime && (
            <PriceCalculator
              serviceType={selectedService.id}
              withChemistry={withChemistry}
              chemistryTimeMinutes={withChemistry ? selectedChemistryTime : 0}
              rentalTimeMinutes={selectedRentalTime}
              theme={theme}
            />
          )}

          {/* Галочка "Хочу получить чек" */}
          {selectedService && selectedRentalTime && (
            <Card theme={theme} className={styles.optionCard}>
              <div className={styles.optionRow}>
                <label className={`${styles.optionLabel} ${themeClass}`}>
                  <input 
                    type="checkbox" 
                    checked={wantReceipt} 
                    onChange={(e) => handleWantReceiptChange(e.target.checked)}
                    className={styles.checkbox}
                  />
                  <span className={styles.checkmark}></span>
                  Хочу получить чек
                </label>
              </div>
              <p className={`${styles.optionDescription} ${themeClass}`}>
                Чек будет отправлен на указанный email после оплаты
              </p>
            </Card>
          )}

          {/* Ввод email - показывается только если галочка включена */}
          {selectedService && selectedRentalTime && wantReceipt && (
            <EmailInput
              value={email || ''}
              onChange={handleEmailChange}
              theme={theme}
              showRememberCheckbox={user && 
                email && 
                email !== user.email && 
                email.length > 0}
              rememberChecked={rememberEmail}
              onRememberChange={handleRememberEmailChange}
              savedEmail={user?.email || ''}
            />
          )}
        </div>
      )}
      
      <div className={styles.buttonContainer}>
        <Button 
          theme={theme} 
          onClick={handleConfirm}
          disabled={!canConfirm || loading || savingCarNumber || savingEmail}
          className={styles.confirmButton}
        >
          {savingCarNumber || savingEmail ? 'Сохранение...' : 'Подтвердить выбор'}
        </Button>
      </div>
    </div>
  );
};

export default ServiceSelector;
