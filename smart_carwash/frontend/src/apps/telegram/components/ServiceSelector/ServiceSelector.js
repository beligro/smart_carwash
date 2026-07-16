import React, { useState, useEffect } from 'react';
import styles from './ServiceSelector.module.css';
import { Card, Button } from '../../../../shared/components/UI';
import CarNumberInput from '../CarNumberInput';
import PriceCalculator from '../PriceCalculator';
import ApiService from '../../../../shared/services/ApiService';
import { validateAndNormalizeLicensePlate } from '../../../../shared/utils/licensePlateUtils';

/**
 * Компонент ServiceSelector - позволяет выбрать тип услуги и дополнительные опции
 * @param {Object} props - Свойства компонента
 * @param {Function} props.onSelect - Функция, вызываемая при выборе услуги
 * @param {string} props.theme - Тема оформления ('light' или 'dark')
 * @param {Object} props.user - Данные пользователя (для получения сохраненного номера)
 */
const ServiceSelector = ({ onSelect, theme = 'light', user, initialServiceType }) => {
  const [selectedService, setSelectedService] = useState(null);
  const [withChemistry, setWithChemistry] = useState(false);
  const [rentalTimes, setRentalTimes] = useState([]);
  const [selectedRentalTime, setSelectedRentalTime] = useState(null);
  const [chemistryTimes, setChemistryTimes] = useState([]);
  const [selectedChemistryTime, setSelectedChemistryTime] = useState(null);
  const [filteredChemistryTimes, setFilteredChemistryTimes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [loadingChemistryTimes, setLoadingChemistryTimes] = useState(false);
  // Блокировка кнопки на время создания сессии (защита от повторных нажатий/дублей запросов)
  const [submitting, setSubmitting] = useState(false);
  const [carNumber, setCarNumber] = useState('');
  const [carNumberCountry, setCarNumberCountry] = useState('RUS');
  const [rememberCarNumber, setRememberCarNumber] = useState(false);
  const [savingCarNumber, setSavingCarNumber] = useState(false);
  const [noCarNumber, setNoCarNumber] = useState(false);
  const [showDisclaimer, setShowDisclaimer] = useState(false);
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [confirmationChecked, setConfirmationChecked] = useState(false);
  const [savedCarNumber, setSavedCarNumber] = useState(''); // Сохраняем номер при переключении "нет номера"
  
  // Состояния для улучшенного UX валидации госномера
  const [carNumberError, setCarNumberError] = useState('');
  const [showCarNumberError, setShowCarNumberError] = useState(false);
  
  const themeClass = theme === 'dark' ? styles.dark : styles.light;
  
  // Инициализация номера машины и email из данных пользователя
  useEffect(() => {
    try {
      if (user) {
        if (user.car_number) {
          setCarNumber(user.car_number);
          // Для сохраненного номера не показываем дисклеймер и подтверждение
          // Он будет работать без дополнительных подтверждений
        }
        if (user.car_number_country) {
          setCarNumberCountry(user.car_number_country);
        }
      }
    } catch (error) {
      console.error('Ошибка при инициализации данных пользователя:', error);
    }
  }, [user]);

  // Сброс состояния "нет номера" при изменении пользователя
  useEffect(() => {
    if (user && user.car_number) {
      setNoCarNumber(false);
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

  // Предвыбор типа услуги (например, при возврате «Оплатить и остаться»)
  useEffect(() => {
    if (initialServiceType && !selectedService) {
      const match = serviceTypes.find((s) => s.id === initialServiceType);
      if (match) {
        handleServiceSelect(match);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialServiceType]);

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

  // Обработчик изменения чекбокса "запомнить номер"
  const handleRememberCarNumberChange = (checked) => {
    try {
      setRememberCarNumber(checked);
    } catch (error) {
      console.error('Ошибка в handleRememberCarNumberChange:', error);
    }
  };

  // Обработчик изменения чекбокса "нет номера"
  const handleNoCarNumberChange = (checked) => {
    try {
      setNoCarNumber(checked);
      
      if (checked) {
        // Если включаем "нет номера", сохраняем текущий номер и очищаем поля
        if (carNumber) {
          setSavedCarNumber(carNumber);
        }
        setCarNumber('');
        setCarNumberError('');
        setShowCarNumberError(false);
        setShowDisclaimer(false);
        setShowConfirmation(false);
        setConfirmationChecked(false);
      } else {
        // Если выключаем "нет номера", восстанавливаем сохраненный номер
        if (savedCarNumber) {
          setCarNumber(savedCarNumber);
          // Для восстановленного номера показываем дисклеймер/подтверждение только если это НЕ сохраненный номер
          if (savedCarNumber !== user?.car_number) {
            const validation = validateCarNumberWithDetails(savedCarNumber);
            if (validation.isValid) {
              setShowDisclaimer(true);
              setShowConfirmation(true);
              setConfirmationChecked(true);
            }
          }
        }
      }
    } catch (error) {
      console.error('Ошибка в handleNoCarNumberChange:', error);
    }
  };

  // Обработчик изменения чекбокса подтверждения
  const handleConfirmationChange = (checked) => {
    try {
      setConfirmationChecked(checked);
      // Скрываем дисклеймер при подтверждении, показываем при снятии галочки
      setShowDisclaimer(!checked);
    } catch (error) {
      console.error('Ошибка в handleConfirmationChange:', error);
    }
  };


  // Улучшенная валидация госномера с детальными сообщениями
  const validateCarNumberWithDetails = (number) => {
    try {
      if (!number || typeof number !== 'string') {
        return {
          isValid: false,
          error: 'Введите номер автомобиля',
          suggestion: 'Номер должен содержать буквы и цифры'
        };
      }

      const validation = validateAndNormalizeLicensePlate(number, carNumberCountry);
      
      if (!validation.isValid) {
        // Детализируем ошибку для лучшего UX
        let error = validation.error;
        let suggestion = '';
        
        if (number.length < 8) {
          suggestion = 'Номер слишком короткий. Пример: А123ВС77';
        } else if (number.length > 9) {
          suggestion = 'Номер слишком длинный. Пример: А123ВС77';
        } else if (!/[АВЕКМНОРСТУХA-Z]/.test(number)) {
          suggestion = 'Используйте только поддерживаемые буквы: А, В, Е, К, М, Н, О, Р, С, Т, У, Х';
        } else if (!/\d/.test(number)) {
          suggestion = 'Номер должен содержать цифры. Пример: А123ВС77';
        } else {
          suggestion = 'Проверьте формат номера. Пример: А123ВС77';
        }
        
        return {
          isValid: false,
          error: error,
          suggestion: suggestion
        };
      }

      return {
        isValid: true,
        error: '',
        suggestion: ''
      };
    } catch (error) {
      console.error('Ошибка в validateCarNumberWithDetails:', error);
      return {
        isValid: false,
        error: 'Ошибка валидации номера',
        suggestion: 'Попробуйте ввести номер заново'
      };
    }
  };

  // Безопасная проверка номера машины (используем новую утилиту)
  const isValidCarNumber = (number) => {
    try {
      if (!number || typeof number !== 'string') {
        return false;
      }
      
      const validation = validateAndNormalizeLicensePlate(number, carNumberCountry);
      return validation.isValid;
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
      // Нормализуем номер перед сохранением
      const validation = validateAndNormalizeLicensePlate(carNumber, carNumberCountry);
      if (!validation.isValid) {
        console.error('Номер машины невалидный:', validation.error);
        return;
      }
      
      await ApiService.updateCarNumber(user.id, validation.normalized, carNumberCountry);
      console.log('Номер машины сохранен:', validation.normalized, 'страна:', carNumberCountry);
    } catch (error) {
      console.error('Ошибка при сохранении номера машины:', error);
    } finally {
      setSavingCarNumber(false);
    }
  };


  // Обработчик изменения госномера с валидацией в реальном времени
  const handleCarNumberChange = (newCarNumber) => {
    setCarNumber(newCarNumber);
    
    // Валидируем в реальном времени только если НЕ выбрано "нет номера"
    if (!noCarNumber) {
      const validation = validateCarNumberWithDetails(newCarNumber);
      if (!validation.isValid) {
        setCarNumberError(validation.suggestion); // Показываем подсказку вместо ошибки
        setShowDisclaimer(false);
        setShowConfirmation(false);
        setConfirmationChecked(false);
      } else {
        setCarNumberError(''); // Очищаем ошибку при валидном номере
        setShowDisclaimer(true); // Показываем дисклеймер при валидном номере
        setShowConfirmation(true); // Показываем подтверждение
      }
    } else {
      setCarNumberError(''); // Очищаем ошибку если выбрано "нет номера"
      setShowDisclaimer(false);
      setShowConfirmation(false);
      setConfirmationChecked(false);
    }
    
    // Скрываем ошибку при начале ввода
    if (showCarNumberError) {
      setShowCarNumberError(false);
    }
  };
  const handleConfirm = async () => {
    try {
      // Проверяем валидность госномера только если номер указан
      if (!noCarNumber && carNumber) {
        const carNumberValidation = validateCarNumberWithDetails(carNumber);
        if (!carNumberValidation.isValid) {
          setCarNumberError(carNumberValidation.error);
          setShowCarNumberError(true);
          // Показываем ошибку на 5 секунд
          setTimeout(() => {
            setShowCarNumberError(false);
          }, 5000);
          return;
        }
      }

      if (selectedService && selectedRentalTime && (noCarNumber || carNumber)) {
        // Если пользователь хочет запомнить номер, сохраняем его
        if (rememberCarNumber && carNumber) {
          await saveCarNumber();
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

        // Нормализуем номер машины перед отправкой (только если номер указан)
        let normalizedCarNumber = '';
        if (!noCarNumber && carNumber) {
          const validation = validateAndNormalizeLicensePlate(carNumber, carNumberCountry);
          if (!validation.isValid) {
            alert('Неверный формат номера машины: ' + validation.error);
            return;
          }
          normalizedCarNumber = validation.normalized;
        }

        const serviceData = {
          serviceType: selectedService.id,
          withChemistry: selectedService.hasChemistry ? withChemistry : false,
          chemistryTimeMinutes: chemistryTime,
          rentalTimeMinutes: selectedRentalTime,
          carNumber: normalizedCarNumber, // Используем нормализованный номер или пустую строку
          carNumberCountry: noCarNumber ? '' : carNumberCountry, // Передаем страну только если номер указан
          email: null // Email больше не используется
        };

        // Блокируем кнопку на время запроса, чтобы повторные нажатия не создавали
        // дубли запросов создания сессии. Ошибку показывает родитель (onSelect).
        setSubmitting(true);
        try {
          await onSelect(serviceData);
        } finally {
          setSubmitting(false);
        }
      }
    } catch (error) {
      alert('Ошибка в handleConfirm: ' + error.message);
    }
  };

  // Проверяем, можно ли подтвердить выбор
  const canConfirm = selectedService && 
    selectedRentalTime && 
    (noCarNumber || (carNumber && carNumber.length >= 6 && isValidCarNumber(carNumber) && 
      (carNumber === user?.car_number || confirmationChecked))) && // Для сохраненного номера подтверждение не требуется
    (!withChemistry || selectedChemistryTime); // Если химия включена, должно быть выбрано время

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
            country={carNumberCountry}
            onCountryChange={setCarNumberCountry}
            theme={theme}
            showRememberCheckbox={showRememberCheckbox}
            rememberChecked={rememberCarNumber}
            onRememberChange={handleRememberCarNumberChange}
            savedCarNumber={user?.car_number || ''}
            noCarNumber={noCarNumber}
            onNoCarNumberChange={handleNoCarNumberChange}
            showDisclaimer={showDisclaimer}
            showConfirmation={showConfirmation}
            confirmationChecked={confirmationChecked}
            onConfirmationChange={handleConfirmationChange}
          />

          {/* Отображение ошибки/подсказки для госномера - только если НЕ выбрано "нет номера" */}
          {!noCarNumber && carNumberError && (
            <div className={`${styles.carNumberError} ${themeClass} ${showCarNumberError ? styles.errorVisible : styles.suggestionVisible}`}>
              <div className={styles.errorIcon}>
                {showCarNumberError ? '⚠️' : '💡'}
              </div>
              <div className={styles.errorText}>
                {showCarNumberError ? 'Указан некорректный номер машины. Проверьте его, пожалуйста' : carNumberError}
              </div>
            </div>
          )}


          {/* Показываем опции только если НЕ выбрано "нет номера" */}
          {!noCarNumber && (
            <>
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
            </>
          )}

        </div>
      )}
      
      {/* Показываем кнопку только если НЕ выбрано "нет номера" */}
      {!noCarNumber && (
        <div className={styles.buttonContainer}>
          <Button 
            theme={theme} 
            onClick={handleConfirm}
            disabled={!canConfirm || loading || savingCarNumber || submitting}
            className={styles.confirmButton}
          >
            {submitting ? 'Создание…' : savingCarNumber ? 'Сохранение...' : 'Подтвердить выбор'}
          </Button>
        </div>
      )}
    </div>
  );
};

export default ServiceSelector;
