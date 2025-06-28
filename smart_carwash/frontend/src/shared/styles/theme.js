// Файл с общими темами и цветами для приложения

const lightTheme = {
    backgroundColor: '#F5F5F5',
    cardBackground: '#FFFFFF',
    textColor: '#000000',
    secondaryTextColor: '#666666',
    borderColor: '#EEEEEE',
    primaryColor: '#007BFF',
    primaryColorHover: '#0056b3',
    secondaryColor: '#6c757d',
    secondaryColorHover: '#5a6268',
    dangerColor: '#dc3545',
    dangerColorHover: '#bd2130',
    successColor: '#28a745',
    errorColor: '#dc3545',
    disabledColor: '#cccccc',
    inputBackground: '#FFFFFF',
    
    // Статусы
    statusColors: {
      created: {
        background: '#E3F2FD',
        text: '#0D47A1'
      },
      assigned: {
        background: '#BBDEFB',
        text: '#1565C0'
      },
      active: {
        background: '#E8F5E9',
        text: '#2E7D32'
      },
      complete: {
        background: '#F1F8E9',
        text: '#558B2F'
      },
      canceled: {
        background: '#FFEBEE',
        text: '#C62828'
      },
      free: {
        background: '#E8F5E9',
        text: '#2E7D32'
      },
      reserved: {
        background: '#BBDEFB',
        text: '#1565C0'
      },
      busy: {
        background: '#FFEBEE',
        text: '#C62828'
      },
      maintenance: {
        background: '#FFF3E0',
        text: '#F57C00'
      }
    },
    
    // Таймер
    timerColors: {
      normal: '#000000',
      warning: '#F57F17',
      danger: '#C62828'
    }
  };
  
  const darkTheme = {
    backgroundColor: '#1E1E1E',
    cardBackground: '#2C2C2C',
    textColor: '#FFFFFF',
    secondaryTextColor: '#E0E0E0',
    borderColor: '#444444',
    primaryColor: '#0D6EFD',
    primaryColorHover: '#0b5ed7',
    secondaryColor: '#6c757d',
    secondaryColorHover: '#5a6268',
    dangerColor: '#dc3545',
    dangerColorHover: '#bd2130',
    successColor: '#28a745',
    errorColor: '#dc3545',
    disabledColor: '#555555',
    inputBackground: '#3C3C3C',
    
    // Статусы
    statusColors: {
      created: {
        background: '#0D47A1',
        text: '#FFFFFF'
      },
      assigned: {
        background: '#1565C0',
        text: '#FFFFFF'
      },
      active: {
        background: '#2E7D32',
        text: '#FFFFFF'
      },
      complete: {
        background: '#558B2F',
        text: '#FFFFFF'
      },
      canceled: {
        background: '#C62828',
        text: '#FFFFFF'
      },
      free: {
        background: '#2E7D32',
        text: '#FFFFFF'
      },
      reserved: {
        background: '#1565C0',
        text: '#FFFFFF'
      },
      busy: {
        background: '#C62828',
        text: '#FFFFFF'
      },
      maintenance: {
        background: '#F57C00',
        text: '#FFFFFF'
      }
    },
    
    // Таймер
    timerColors: {
      normal: '#FFFFFF',
      warning: '#F57F17',
      danger: '#C62828'
    }
  };
  
  // Функция для получения темы в зависимости от режима
  const getTheme = (mode) => {
    return mode === 'dark' ? darkTheme : lightTheme;
  };
  
  export { lightTheme, darkTheme, getTheme };
